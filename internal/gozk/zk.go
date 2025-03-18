package gozk

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/canhlinh/log4go"
)

const (
	DefaultTimezone = "Asia/Ho_Chi_Minh"
)

var (
	KeepAlivePeriod   = time.Second * 60
	ReadSocketTimeout = 3 * time.Second
)

type ZK struct {
	conn      *net.TCPConn
	sessionID int
	replyID   int
	host      string
	port      int
	pin       int
	loc       *time.Location
	lastData  []byte
	disabled  bool
	capturing chan bool
	recordCap int
}

func NewZK(host string, port int, pin int, timezone string) *ZK {
	return &ZK{
		host:      host,
		port:      port,
		pin:       pin,
		loc:       LoadLocation(timezone),
		sessionID: 0,
		replyID:   USHRT_MAX - 1,
	}
}

func (zk *ZK) Connect() error {
	if zk.conn != nil {
		return errors.New("already connected")
	}

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(zk.host, strconv.Itoa(zk.port)), 3*time.Second)
	if err != nil {
		return err
	}

	tcpConnection := conn.(*net.TCPConn)
	if err := tcpConnection.SetKeepAlive(true); err != nil {
		return err
	}

	if err := tcpConnection.SetKeepAlivePeriod(KeepAlivePeriod); err != nil {
		return err
	}

	zk.conn = tcpConnection

	res, err := zk.sendCommand(CMD_CONNECT, nil, 8)
	if err != nil {
		return err
	}

	zk.sessionID = res.CommandID
	//
	// if res.Code == CMD_ACK_UNAUTH {
	// 	commandString, _ := makeCommKey(zk.pin, zk.sessionID, 50)
	// 	res, err := zk.sendCommand(CMD_AUTH, commandString, 8)
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	if !res.Status {
	// 		return errors.New("unauthorized")
	// 	}
	// }
	//
	log.Println("Connected with session_id", zk.sessionID)
	return nil
}

func (zk *ZK) sendCommand(command int, commandString []byte, responseSize int) (*Response, error) {
	if commandString == nil {
		commandString = make([]byte, 0)
	}

	header, err := createHeader(command, commandString, zk.sessionID, zk.replyID)
	if err != nil {
		return nil, err
	}

	top, err := createTCPTop(header)
	if err != nil && err != io.EOF {
		return nil, err
	}

	if n, err := zk.conn.Write(top); err != nil {
		return nil, err
	} else if n == 0 {
		return nil, errors.New("failed to write command")
	}

	zk.conn.SetReadDeadline(time.Now().Add(ReadSocketTimeout))
	tcpDataRecieved := make([]byte, responseSize+8)
	bytesReceived, err := zk.conn.Read(tcpDataRecieved)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("GOT ERROR %s ON COMMAND %d", err.Error(), command)
	}
	tcpLength := testTCPTop(tcpDataRecieved)
	if bytesReceived == 0 || tcpLength == 0 {
		return nil, errors.New("TCP packet invalid")
	}
	receivedHeader, err := newBP().UnPack([]string{"H", "H", "H", "H"}, tcpDataRecieved[8:16])
	if err != nil {
		return nil, err
	}

	resCode := receivedHeader[0].(int)
	commandID := receivedHeader[2].(int)
	zk.replyID = receivedHeader[3].(int)
	zk.lastData = tcpDataRecieved[16:bytesReceived]

	switch resCode {
	case CMD_ACK_OK, CMD_PREPARE_DATA, CMD_DATA:
		return &Response{
			Status:    true,
			Code:      resCode,
			TCPLength: tcpLength,
			CommandID: commandID,
			Data:      zk.lastData,
			ReplyID:   zk.replyID,
		}, nil
	default:
		return &Response{
			Status:    false,
			Code:      resCode,
			TCPLength: tcpLength,
			CommandID: commandID,
			Data:      zk.lastData,
			ReplyID:   zk.replyID,
		}, nil
	}
}

// Disconnect disconnects out of the machine fingerprint
func (zk *ZK) Disconnect() error {
	if zk.conn == nil {
		return errors.New("already disconnected")
	}

	if _, err := zk.sendCommand(CMD_EXIT, nil, 8); err != nil {
		return err
	}

	if err := zk.conn.Close(); err != nil {
		return err
	}

	zk.conn = nil
	return nil
}

// EnableDevice enables the connected device
func (zk *ZK) EnableDevice() error {
	res, err := zk.sendCommand(CMD_ENABLEDEVICE, nil, 8)
	if err != nil {
		return err
	}

	if !res.Status {
		return errors.New("failed to enable device")
	}

	zk.disabled = false
	return nil
}

// DisableDevice disable the connected device
func (zk *ZK) DisableDevice() error {
	res, err := zk.sendCommand(CMD_DISABLEDEVICE, nil, 8)
	if err != nil {
		return err
	}

	if !res.Status {
		return errors.New("failed to disable device")
	}

	zk.disabled = true
	return nil
}

func (zk *ZK) GetZktecoUsers() ([]*User, error) {
	var (
		records   int
		err       error
		userdata  []byte
		size      int
		totalSize int
		users     = make([]*User, 0)
		v         []interface{}
	)

	if records, err = zk.readSize(); err != nil {
		fmt.Printf("zk read size error: %s", err)
		return nil, err
	}

	userdata, size, err = zk.readWithBuffer(CMD_USERTEMP_RRQ, FCT_USER, 0)
	if err != nil {
		fmt.Printf("zk readWithBuffer for userdata error: %s", err)
		return nil, err
	}

	if size <= 4 {
		fmt.Printf("size too short can't been read .")
		return nil, errors.New("size too short can't been read")
	}

	totalSize = mustUnpack([]string{"I"}, userdata[:4])[0].(int)
	if totalSize/records == 8 || totalSize/records == 16 {
		fmt.Printf("Sorry I don't support this kind of device. I'm lazy!  totalSize = %d ; size = %d\n", totalSize, size)
		return nil, errors.New("sorry I don't support this kind of device. I'm lazy")
	}

	// 重新赋值
	userdata = userdata[4:]
	for len(userdata) >= 72 { // 只处理72
		v, err = newBP().UnPack([]string{"H", "B", "8s", "24s", "I", "7s", "24s"}, userdata[:72])
		if err != nil {
			fmt.Printf("userdata unpack err : %v\n", err)
			return nil, err
		}
		name := string([]byte(v[3].(string)))
		users = append(users, &User{
			Name: strings.Replace(name, "\u0000", "", -1),
			Uid:  strings.Replace(v[6].(string), "\x00", "", -1),
		})
		userdata = userdata[72:]
	}

	return users, nil
}

// GetAttendances returns total attendances from the connected device
func (zk *ZK) GetAttendances() ([]*Attendance, error) {
	if err := zk.GetUsers(); err != nil {
		return nil, err
	}

	properties, err := zk.GetProperties()
	if err != nil {
		return nil, err
	}

	data, size, err := zk.readWithBuffer(CMD_ATTLOG_RRQ, 0, 0)
	if err != nil {
		return nil, err
	}

	if size < 4 {
		return []*Attendance{}, nil
	}

	totalSizeByte := data[:4]
	data = data[4:]

	totalSize := mustUnpack([]string{"I"}, totalSizeByte)[0].(int)
	recordSize := totalSize / properties.TotalRecords
	attendances := []*Attendance{}

	if recordSize == 8 || recordSize == 16 {
		return nil, errors.New("sorry I don't support this kind of device. I'm lazy")
	}

	for len(data) >= 40 {

		v, err := newBP().UnPack([]string{"H", "24s", "B", "4s", "B", "8s"}, data[:40])
		if err != nil {
			return nil, err
		}

		timestamp, err := zk.decodeTime([]byte(v[3].(string)))
		if err != nil {
			return nil, err
		}

		// userID, err := strconv.ParseInt(strings.Replace(v[1].(string), "\x00", "", -1), 10, 64)
		userID := strings.Replace(v[1].(string), "\x00", "", -1)
		if err != nil {
			return nil, err
		}

		attendances = append(attendances, &Attendance{AttendedAt: timestamp, UserID: userID})
		data = data[40:]
	}

	return attendances, nil
}

// GetUsers returns a list of users
// For now, just run this func. I'll implement this function later on.
func (zk *ZK) GetUsers() error {
	_, err := zk.readSize()
	if err != nil {
		return err
	}
	_, size, err := zk.readWithBuffer(CMD_USERTEMP_RRQ, FCT_USER, 0)
	if err != nil {
		return err
	}

	if size < 4 {
		return nil
	}
	return nil
}

func (zk *ZK) LiveCapture() (chan *Attendance, error) {
	if zk.capturing != nil {
		return nil, errors.New("is capturing")
	}

	users, err := zk.GetZktecoUsers()
	if err != nil {
		return nil, err
	}

	if err := zk.verifyUser(); err != nil {
		return nil, err
	}

	// First disable the device to ensure no pending operations
	if !zk.disabled {
		if err := zk.DisableDevice(); err != nil {
			return nil, err
		}
	}

	// Clear any existing event registrations
	if err := zk.regEvent(0); err != nil {
		return nil, err
	}

	// Register for attendance log events
	if err := zk.regEvent(EF_ATTLOG); err != nil {
		return nil, err
	}

	// Re-enable the device
	if err := zk.EnableDevice(); err != nil {
		return nil, err
	}

	log4go.Info("Start capturing")
	zk.capturing = make(chan bool, 1)
	c := make(chan *Attendance, 1)

	go func() {
		defer func() {
			log4go.Info("Stopped capturing")
			zk.regEvent(0)
			close(c)
		}()

		for {
			select {
			case <-zk.capturing:
				return
			default:
				data, err := zk.receiveData(1032, KeepAlivePeriod)
				if err != nil {
					if strings.Contains(err.Error(), "timeout") {
						// Timeout is expected, send keep-alive
						_, err := zk.sendCommand(CMD_REG_EVENT, nil, 8)
						if err != nil {
							log4go.Error("Failed to send keep-alive:", err)
							return
						}
						continue
					}
					log4go.Error("Error receiving data:", err)
					return
				}

				// Send acknowledgment
				if err := zk.ackOK(); err != nil {
					log4go.Error("Failed to send ACK:", err)
					return
				}

				if len(data) == 0 {
					log4go.Info("Empty data received, continuing")
					continue
				}

				header := mustUnpack([]string{"H", "H", "H", "H"}, data[8:16])
				data = data[16:]

				if header[0].(int) != CMD_REG_EVENT {
					log4go.Info("Not an event, skipping")
					continue
				}

				// Print the data in a more readable format for debugging
				dataStr := ""
				for _, b := range data {
					if b >= 32 && b <= 126 { // Printable ASCII
						dataStr += string(b)
					} else {
						dataStr += fmt.Sprintf("\\x%02x", b)
					}
				}
				log4go.Info("Data as string:", dataStr)
				for len(data) >= 10 {
					var userID string
					var status, punch int
					var timehex string
					var unpack []interface{}

					if len(data) == 10 {
						unpack = mustUnpack([]string{"H", "B", "B", "6s"}, data)
						userID = fmt.Sprintf("%d", unpack[0].(int))
						status = unpack[1].(int)
						punch = unpack[2].(int)
						timehex = unpack[3].(string)
						data = data[10:]
					} else if len(data) == 12 {
						unpack = mustUnpack([]string{"I", "B", "B", "6s"}, data)
						userID = fmt.Sprintf("%d", unpack[0].(int))
						status = unpack[1].(int)
						punch = unpack[2].(int)
						timehex = unpack[3].(string)
						data = data[12:]
					} else if len(data) == 14 {
						unpack = mustUnpack([]string{"H", "B", "B", "6s", "4s"}, data)
						userID = fmt.Sprintf("%d", unpack[0].(int))
						status = unpack[1].(int)
						punch = unpack[2].(int)
						timehex = unpack[3].(string)
						data = data[14:]
					} else if len(data) == 32 {
						unpack = mustUnpack([]string{"24s", "B", "B", "6s"}, data[:32])
						userID = strings.Replace(unpack[0].(string), "\x00", "", -1)
						status = unpack[1].(int)
						punch = unpack[2].(int)
						timehex = unpack[3].(string)
						data = data[32:]
					} else if len(data) == 36 {
						// First try to parse as a 24-byte user ID format
						unpack = mustUnpack([]string{"24s", "B", "B", "6s", "4s"}, data[:36])
						rawUserID := unpack[0].(string)
						// Check if the first few bytes contain ASCII digits (common for numeric user IDs)
						if rawUserID[0] >= '0' && rawUserID[0] <= '9' {
							// This appears to be a numeric ID stored as ASCII in the first part
							// Extract until we hit a null byte or non-numeric character
							numericPart := ""
							for i := 0; i < len(rawUserID); i++ {
								if rawUserID[i] == 0 || !(rawUserID[i] >= '0' && rawUserID[i] <= '9') {
									break
								}
								numericPart += string(rawUserID[i])
							}
							if numericPart != "" {
								userID = numericPart
							} else {
								userID = strings.Replace(rawUserID, "\x00", "", -1)
							}
						} else {
							userID = strings.Replace(rawUserID, "\x00", "", -1)
						}
						status = unpack[1].(int)
						punch = unpack[2].(int)
						timehex = unpack[3].(string)
						data = data[36:]
					} else if len(data) == 37 {
						unpack = mustUnpack([]string{"24s", "B", "B", "6s", "5s"}, data[:37])
						userID = strings.Replace(unpack[0].(string), "\x00", "", -1)
						status = unpack[1].(int)
						punch = unpack[2].(int)
						timehex = unpack[3].(string)
						data = data[37:]
					} else if len(data) >= 52 {
						unpack = mustUnpack([]string{"24s", "B", "B", "6s", "20s"}, data[:52])
						userID = strings.Replace(unpack[0].(string), "\x00", "", -1)
						status = unpack[1].(int)
						punch = unpack[2].(int)
						timehex = unpack[3].(string)
						data = data[52:]
					}

					timestamp := zk.decodeTimeHex([]byte(timehex))

					uid := ""
					// Find matching user by user_id
					for _, user := range users {
						if user.Uid == userID {
							uid = userID
							break
						}
					}
					if uid == "" {
						uid = userID
					}

					c <- &Attendance{UserID: userID, AttendedAt: timestamp}
					log.Printf("UserID %v timestamp %v status %v punch %v\n", userID, timestamp, status, punch)
				}
			}
		}
	}()

	return c, nil
}

func (zk ZK) StopCapture() {
	zk.capturing <- false
}

func (zk *ZK) IsConnected() bool {
	if zk.conn == nil {
		return false
	}

	// Try sending a simple command to check the connection
	_, err := zk.sendCommand(CMD_GET_TIME, nil, 8)
	return err == nil
}

func (zk *ZK) Reconnect() error {
	if zk.IsConnected() {
		return nil
	}

	return zk.Connect()
}

func (zk ZK) Clone() *ZK {
	return &ZK{
		host:      zk.host,
		port:      zk.port,
		pin:       zk.pin,
		loc:       zk.loc,
		sessionID: 0,
		replyID:   USHRT_MAX - 1,
	}
}

func (zk *ZK) GetTime() (time.Time, error) {
	res, err := zk.sendCommand(CMD_GET_TIME, nil, 1032)
	if err != nil {
		return time.Now(), err
	}
	if !res.Status {
		return time.Now(), errors.New("can not get time")
	}

	return zk.decodeTime(res.Data[:4])
}

func (zk *ZK) SetTime(t time.Time) error {
	truncatedTime := t.Truncate(time.Second)
	log.Println("Set new time:", truncatedTime)

	commandString, err := newBP().Pack([]string{"I"}, []interface{}{zk.encodeTime(truncatedTime)})
	if err != nil {
		return err
	}
	res, err := zk.sendCommand(CMD_SET_TIME, commandString, 8)
	if err != nil {
		return err
	}
	if !res.Status {
		return errors.New("can not set time")
	}
	return nil
}
