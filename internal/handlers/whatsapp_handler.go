package handlers

import (
	"context"
	"fmt"
	"maya-canteen/internal/database"
	"maya-canteen/internal/handlers/common"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

type Client = whatsmeow.Client

// type JID = types.JID

// type Instance struct {
// 	ID              string
// 	Client          *Client
// 	QrCodeRateLimit uint16
// }

// type Message struct {
// 	InstanceID string
// 	Body       string
// 	SenderJID  string
// 	ChatJID    string
// 	MessageID  string
// 	FromMe     bool
// 	Timestamp  time.Time
// 	MediaType  *MediaType
// 	Media      *[]byte
// 	Mimetype   *string
// }

// type MediaType int

// const (
// 	Audio MediaType = iota
// 	Image
// 	Document
// 	Sticker
// 	// Video
// )

// func (m MediaType) String() string {
// 	switch m {
// 	case Audio:
// 		return "audio"
// 	case Document:
// 		return "document"
// 	case Sticker:
// 		return "sticker"
// 	// case Video:
// 	// 	return "video"
// 	case Image:
// 		return "image"
// 	}
// 	return "unknown"
// }

// type ContactInfo struct {
// 	Phone   string `json:"phone"`
// 	Name    string `json:"name"`
// 	Status  string `json:"status"`
// 	Picture string `json:"picture"`
// }

// type MessageResponse struct {
// 	ID        string
// 	Sender    JID
// 	Timestamp time.Time
// }

// type DownloadResponse struct {
// 	Data     []byte
// 	Type     MediaType
// 	Mimetype string
// }

// type UploadResponse struct {
// 	URL           string
// 	DirectPath    string
// 	Mimetype      MediaType
// 	MediaKey      []byte
// 	FileEncSHA256 []byte
// 	FileSHA256    []byte
// 	FileLength    uint64
// }

// type IsOnWhatsAppResponse struct {
// 	Query        string `json:"query"`
// 	Phone        string `json:"phone"`
// 	IsRegistered bool   `json:"is_registered"`
// }

// type WhatsApp interface {
// 	CreateInstance(id string) *Instance
// 	CreateInstanceFromDevice(id string, jid JID) *Instance
// 	IsLoggedIn(instance *Instance) bool
// 	IsConnected(instance *Instance) bool
// 	Disconnect(instance *Instance)
// 	Logout(instance *Instance) error
// 	EventHandler(instance *Instance, handler func(evt any))
// 	InitInstance(instance *Instance, qrcodeHandler func(evt string, qrcode string, err error)) error
// 	SendTextMessage(instance *Instance, jid JID, text string) (MessageResponse, error)
// 	SendAudioMessage(instance *Instance, jid JID, audioURL *dataurl.DataURL, mimitype string) (MessageResponse, error)
// 	SendImageMessage(instance *Instance, jid JID, imageURL *dataurl.DataURL, mimitype string) (MessageResponse, error)
// 	SendDocumentMessage(instance *Instance, jid JID, documentURL *dataurl.DataURL, mimitype string, filename string) (MessageResponse, error)
// 	GetContactInfo(instance *Instance, jid JID) (*ContactInfo, error)
// 	ParseEventMessage(instance *Instance, message *events.Message) (Message, error)
// 	IsOnWhatsApp(instance *Instance, phones []string) ([]IsOnWhatsAppResponse, error)
// }

// type whatsApp struct {
// 	container *sqlstore.Container
// }

// func NewWhatsApp(databasePath string) *whatsApp {
// 	var level = "DEBUG"
// 	dbLog := waLog.Stdout("Database", level, true)

// 	container, err := sqlstore.New("sqlite3", "file:"+databasePath+"?_foreign_keys=on", dbLog)
// 	if err != nil {
// 		fmt.Println("Failed to create database container")
// 		return nil
// 	}
// 	return &whatsApp{container: container}
// }

// func (w *whatsApp) CreateInstance(id string) *Instance {
// 	client := w.createClient(w.container.NewDevice())
// 	return &Instance{
// 		ID:              id,
// 		Client:          client,
// 		QrCodeRateLimit: 10,
// 	}
// }

// func (w *whatsApp) CreateInstanceFromDevice(id string, jid JID) *Instance {
// 	device, _ := w.container.GetDevice(JID{
// 		User:       jid.User,
// 		RawAgent:   jid.RawAgent,
// 		Device:     jid.Device,
// 		Server:     jid.Server,
// 		Integrator: jid.Integrator,
// 	})
// 	if device != nil {
// 		client := w.createClient(device)
// 		return &Instance{
// 			ID:              id,
// 			Client:          client,
// 			QrCodeRateLimit: 10,
// 		}
// 	}
// 	return w.CreateInstance(id)
// }

// func (w *whatsApp) EventHandler(instance *Instance, handler func(evt any)) {
// 	instance.Client.AddEventHandler(handler)
// }

// func (w *whatsApp) InitInstance(instance *Instance, qrcodeHandler func(evt string, qrcode string, err error)) error {
// 	if instance.Client.Store.ID == nil {
// 		go w.generateQrcode(instance, qrcodeHandler)
// 	} else {
// 		err := instance.Client.Connect()
// 		if err != nil {
// 			return err
// 		}

// 		if !instance.Client.WaitForConnection(5 * time.Second) {
// 			return errors.New("websocket didn't reconnect within 5 seconds of failed")
// 		}
// 	}

// 	return nil
// }

// func (w *whatsApp) SendTextMessage(instance *Instance, jid JID, text string) (MessageResponse, error) {
// 	message := &waProto.Message{
// 		ExtendedTextMessage: &waProto.ExtendedTextMessage{
// 			Text: &text,
// 		},
// 	}
// 	return w.sendMessage(instance, jid, message)
// }

// func (w *whatsApp) SendAudioMessage(instance *Instance, jid JID, audioURL *dataurl.DataURL, mimitype string) (MessageResponse, error) {
// 	uploaded, err := w.uploadMedia(instance, audioURL, Audio)
// 	if err != nil {
// 		return MessageResponse{}, err
// 	}
// 	message := &waProto.Message{
// 		AudioMessage: &waProto.AudioMessage{
// 			PTT:           proto.Bool(true),
// 			URL:           proto.String(uploaded.URL),
// 			DirectPath:    proto.String(uploaded.DirectPath),
// 			MediaKey:      uploaded.MediaKey,
// 			Mimetype:      proto.String(mimitype),
// 			FileEncSHA256: uploaded.FileEncSHA256,
// 			FileSHA256:    uploaded.FileSHA256,
// 			FileLength:    proto.Uint64(uint64(len(audioURL.Data))),
// 		},
// 	}
// 	return w.sendMessage(instance, jid, message)
// }

// func (w *whatsApp) SendImageMessage(instance *Instance, jid JID, imageURL *dataurl.DataURL, mimitype string) (MessageResponse, error) {
// 	uploaded, err := w.uploadMedia(instance, imageURL, Image)
// 	if err != nil {
// 		return MessageResponse{}, err
// 	}
// 	message := &waProto.Message{
// 		ImageMessage: &waProto.ImageMessage{
// 			URL:           proto.String(uploaded.URL),
// 			DirectPath:    proto.String(uploaded.DirectPath),
// 			MediaKey:      uploaded.MediaKey,
// 			Mimetype:      proto.String(mimitype),
// 			FileEncSHA256: uploaded.FileEncSHA256,
// 			FileSHA256:    uploaded.FileSHA256,
// 			FileLength:    proto.Uint64(uint64(len(imageURL.Data))),
// 		},
// 	}
// 	return w.sendMessage(instance, jid, message)
// }

// func (w *whatsApp) SendDocumentMessage(
// 	instance *Instance, jid JID, documentURL *dataurl.DataURL, mimitype string, filename string) (MessageResponse, error) {
// 	uploaded, err := w.uploadMedia(instance, documentURL, Document)
// 	if err != nil {
// 		return MessageResponse{}, err
// 	}

// 	message := &waProto.Message{
// 		DocumentMessage: &waProto.DocumentMessage{
// 			URL:           proto.String(uploaded.URL),
// 			FileName:      &filename,
// 			DirectPath:    proto.String(uploaded.DirectPath),
// 			MediaKey:      uploaded.MediaKey,
// 			Mimetype:      proto.String(mimitype),
// 			FileEncSHA256: uploaded.FileEncSHA256,
// 			FileSHA256:    uploaded.FileSHA256,
// 			FileLength:    proto.Uint64(uint64(len(documentURL.Data))),
// 		},
// 	}
// 	return w.sendMessage(instance, jid, message)
// }

// func (w *whatsApp) IsOnWhatsApp(instance *Instance, phones []string) ([]IsOnWhatsAppResponse, error) {
// 	isOnWhatsAppResponse, err := instance.Client.IsOnWhatsApp(phones)
// 	if err != nil {
// 		return nil, err
// 	}
// 	data := make([]IsOnWhatsAppResponse, 0, len(isOnWhatsAppResponse))
// 	for _, resp := range isOnWhatsAppResponse {
// 		data = append(data, IsOnWhatsAppResponse{
// 			Query:        resp.Query,
// 			IsRegistered: resp.IsIn,
// 			Phone:        resp.JID.User,
// 		})
// 	}

// 	return data, nil
// }

// func (w *whatsApp) sendMessage(instance *Instance, jid JID, message *waProto.Message) (MessageResponse, error) {
// 	resp, err := instance.Client.SendMessage(context.Background(), jid, message)
// 	if err != nil {
// 		return MessageResponse{}, err
// 	}

// 	return MessageResponse{
// 		ID:        resp.ID,
// 		Sender:    *instance.Client.Store.ID,
// 		Timestamp: resp.Timestamp,
// 	}, nil
// }

// func (w *whatsApp) GetContactInfo(instance *Instance, jid JID) (*ContactInfo, error) {
// 	userInfo, err := instance.Client.GetUserInfo([]JID{jid})
// 	if err != nil {
// 		return nil, err
// 	}

// 	contactInfo, err := instance.Client.Store.Contacts.GetContact(jid)
// 	if err != nil {
// 		return nil, err
// 	}

// 	profilePictureInfo, _ := instance.Client.GetProfilePictureInfo(
// 		jid,
// 		&whatsmeow.GetProfilePictureParams{},
// 	)

// 	profilePictureURL := ""
// 	if profilePictureInfo != nil {
// 		profilePictureURL = profilePictureInfo.URL
// 	}

// 	return &ContactInfo{
// 		Phone:   jid.User,
// 		Name:    contactInfo.PushName,
// 		Status:  userInfo[jid].Status,
// 		Picture: profilePictureURL,
// 	}, nil
// }

// func (w *whatsApp) ParseEventMessage(instance *Instance, message *events.Message) (Message, error) {
// 	media, err := w.downloadMedia(
// 		instance,
// 		message.Message,
// 	)

// 	if err != nil && media == nil {
// 		return Message{}, err
// 	}

// 	text := w.getTextMessage(message.Message)
// 	base := Message{
// 		InstanceID: instance.ID,
// 		Body:       text,
// 		MessageID:  message.Info.ID,
// 		ChatJID:    message.Info.Chat.User,
// 		SenderJID:  message.Info.Sender.User,
// 		FromMe:     message.Info.MessageSource.IsFromMe,
// 		Timestamp:  message.Info.Timestamp,
// 	}

// 	if media != nil && err == nil {
// 		base.MediaType = &media.Type
// 		base.Mimetype = &media.Mimetype
// 		base.Media = &media.Data
// 		return base, nil
// 	}

// 	return base, nil
// }

// func (w *whatsApp) createClient(deviceStore *store.Device) *whatsmeow.Client {
// 	var level = "DEBUG"
// 	log := waLog.Stdout("Client", level, true)
// 	return whatsmeow.NewClient(deviceStore, log)
// }

// func (w *whatsApp) uploadMedia(instance *Instance, media *dataurl.DataURL, mediaType MediaType) (*UploadResponse, error) {
// 	var mType whatsmeow.MediaType
// 	switch mediaType {
// 	case Image:
// 		mType = whatsmeow.MediaImage
// 	case Audio:
// 		mType = whatsmeow.MediaAudio
// 	case Document:
// 		mType = whatsmeow.MediaDocument
// 	default:
// 		return nil, errors.New("unknown media type")
// 	}

// 	uploaded, err := instance.Client.Upload(context.Background(), media.Data, mType)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &UploadResponse{
// 		URL:           uploaded.URL,
// 		Mimetype:      mediaType,
// 		DirectPath:    uploaded.DirectPath,
// 		MediaKey:      uploaded.MediaKey,
// 		FileEncSHA256: uploaded.FileEncSHA256,
// 		FileSHA256:    uploaded.FileSHA256,
// 		FileLength:    uploaded.FileLength,
// 	}, nil
// }

// func (w *whatsApp) downloadMedia(instance *Instance, message *waProto.Message) (*DownloadResponse, error) {
// 	document := message.GetDocumentMessage()
// 	if document != nil {
// 		data, err := instance.Client.Download(document)
// 		if err != nil {
// 			return &DownloadResponse{Type: Document}, err
// 		}

// 		return &DownloadResponse{
// 			Data:     data,
// 			Type:     Document,
// 			Mimetype: document.GetMimetype(),
// 		}, nil
// 	}

// 	audio := message.GetAudioMessage()
// 	if audio != nil {
// 		data, err := instance.Client.Download(audio)
// 		if err != nil {
// 			return &DownloadResponse{Type: Audio}, err
// 		}

// 		return &DownloadResponse{
// 			Data:     data,
// 			Type:     Audio,
// 			Mimetype: audio.GetMimetype(),
// 		}, nil
// 	}

// 	image := message.GetImageMessage()
// 	if image != nil {
// 		data, err := instance.Client.Download(image)
// 		if err != nil {
// 			return &DownloadResponse{Type: Image}, err
// 		}

// 		return &DownloadResponse{
// 			Data:     data,
// 			Type:     Image,
// 			Mimetype: image.GetMimetype(),
// 		}, nil
// 	}

// 	sticker := message.GetStickerMessage()
// 	if sticker != nil {
// 		data, err := instance.Client.Download(sticker)
// 		if err != nil {
// 			return &DownloadResponse{Type: Sticker}, err
// 		}

// 		return &DownloadResponse{
// 			Data:     data,
// 			Type:     Sticker,
// 			Mimetype: sticker.GetMimetype(),
// 		}, nil
// 	}

// 	return nil, nil
// }

// func (w *whatsApp) getTextMessage(message *waProto.Message) string {
// 	extendedTextMessage := message.GetExtendedTextMessage()
// 	if extendedTextMessage != nil {
// 		return *extendedTextMessage.Text
// 	}
// 	return message.GetConversation()
// }

// func (w *whatsApp) generateQrcode(instance *Instance, qrcodeHandler func(evt string, qrcode string, err error)) {
// 	qrChan, err := instance.Client.GetQRChannel(context.Background())
// 	if err != nil {
// 		if !errors.Is(err, whatsmeow.ErrQRStoreContainsID) {
// 			errMessage := fmt.Sprintf("Failed to get qr channel. %s", err)
// 			qrcodeHandler("error", "", errors.New(errMessage))
// 		}
// 	} else {
// 		err = instance.Client.Connect()
// 		if err != nil {
// 			errMessage := fmt.Sprintf("Failed to connect client to WhatsApp websocket. %s", err)
// 			qrcodeHandler("error", "", errors.New(errMessage))
// 		} else {
// 			for evt := range qrChan {
// 				if instance.QrCodeRateLimit == 0 {
// 					qrcodeHandler("rate-limit", "", nil)
// 					return
// 				}

//					switch evt.Event {
//					case "code":
//						instance.QrCodeRateLimit -= 1
//						qrcodeHandler("code", evt.Code, nil)
//					default:
//						qrcodeHandler(evt.Event, "", evt.Error)
//					}
//				}
//			}
//		}
//	}
//
// WhatsAppHandler manages the WhatsApp integration with our application
type WhatsAppHandler struct {
	common.BaseHandler
	whatsappClient *Client
}

// NewWhatsAppHandler creates a new WhatsApp handler with the given database service and client
func NewWhatsAppHandler(db database.Service, client *whatsmeow.Client) *WhatsAppHandler {
	return &WhatsAppHandler{
		BaseHandler:    common.NewBaseHandler(db),
		whatsappClient: client,
	}
}

// SendWhatsAppMessage sends a message to a user's WhatsApp number
func (h *WhatsAppHandler) SendWhatsAppMessage(phoneNumber, message string) error {
	// Check if WhatsApp client is connected
	if h.whatsappClient == nil || !h.whatsappClient.IsLoggedIn() {
		return fmt.Errorf("WhatsApp client is not connected")
	}

	// Parse phone number as JID (WhatsApp ID)
	recipient, err := types.ParseJID(phoneNumber + "@s.whatsapp.net")
	if err != nil {
		return fmt.Errorf("invalid phone number format: %v", err)
	}

	// Create message with current timestamp
	msg := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: proto.String(message),
		},
	}

	// Send message with 10-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = h.whatsappClient.SendMessage(ctx, recipient, msg)
	if err != nil {
		return fmt.Errorf("failed to send WhatsApp message: %v", err)
	}

	return nil
}

// NotifyUserBalance sends a balance notification to a specific user
func (h *WhatsAppHandler) NotifyUserBalance(w http.ResponseWriter, r *http.Request) {
	// Extract employee ID from URL params
	vars := mux.Vars(r)
	employeeID, err := h.ParseID(vars, "id")
	if err != nil {
		common.RespondWithError(w, http.StatusBadRequest, "Employee ID is required")
		return
	}

	// Get user and balance information
	user, err := h.DB.GetUser(employeeID)
	if err != nil {
		common.RespondWithError(w, http.StatusNotFound, fmt.Sprintf("User with employee ID %s not found", strconv.FormatInt(employeeID, 10)))
		return
	}

	// Validate that user has a phone number
	if user.Phone == "" {
		common.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("User with employee ID %s does not have a phone number", strconv.FormatInt(employeeID, 10)))
		return
	}

	// Get balance for the user
	userBalance, err := h.DB.GetUserBalanceByUserID(user.ID)
	if err != nil {
		common.RespondWithError(w, http.StatusInternalServerError, "Failed to get user balance")
		return
	}

	message := fmt.Sprintf(
		"**Balance Update** \n\nDear %s,\nYour current canteen balance is: *PKR %.2f*\n\nPlease pay online via Jazz Cash 03422949447 (Syed Kazim Raza) half month of Canteen bill\n\nThis is an automated message from Maya Canteen Management System.",
		user.Name,
		float64(userBalance.Balance), // Assuming 'Amount' is the numeric field in models.UserBalance
	)

	// Send the message via WhatsApp
	err = h.SendWhatsAppMessage(user.Phone, message)
	if err != nil {
		log.Printf("Error sending WhatsApp balance notification to %s: %v", user.Phone, err)
		common.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to send WhatsApp message: %v", err))
		return
	}

	common.RespondWithJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": fmt.Sprintf("Balance notification sent to %s", user.Name),
	})
}

// NotifyAllUsersBalances sends balance notifications to all users
func (h *WhatsAppHandler) NotifyAllUsersBalances(w http.ResponseWriter, r *http.Request) {
	// Get all users with balances
	userBalances, err := h.DB.GetUsersBalances()
	if err != nil {
		common.RespondWithError(w, http.StatusInternalServerError, "Failed to get user's balance")
		return
	}

	successCount := 0
	failCount := 0
	failedUsers := []string{}

	// Send notification to each user with a phone number
	for _, balance := range userBalances {
		// Skip users without phone numbers
		if balance.Phone == "" {
			failCount++
			failedUsers = append(failedUsers, fmt.Sprintf("%s (no phone number)", balance.UserName))
			continue
		}

		// Format message
		message := fmt.Sprintf(
			"**Balance Update** \n\nDear %s,\nYour current canteen balance is: *PKR %.2f*\n\nPlease pay online via Jazz Cash 03422949447 (Syed Kazim Raza) half month of Canteen bill\n\nThis is an automated message from Maya Canteen Management System.",
			balance.UserName,
			balance.Balance,
		)

		// Send message
		err = h.SendWhatsAppMessage(balance.Phone, message)
		if err != nil {
			log.Printf("Failed to send WhatsApp notification to %s (%s): %v", balance.UserName, balance.Phone, err)
			failCount++
			failedUsers = append(failedUsers, fmt.Sprintf("%s (%v)", balance.UserName, err))
		} else {
			successCount++
		}

		// Add a small delay between messages to avoid rate limiting
		time.Sleep(500 * time.Millisecond)
	}

	common.RespondWithJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": fmt.Sprintf("Sent %d notifications, %d failed", successCount, failCount),
		"details": map[string]any{
			"success_count": successCount,
			"fail_count":    failCount,
			"failed_users":  failedUsers,
		},
	})
}
