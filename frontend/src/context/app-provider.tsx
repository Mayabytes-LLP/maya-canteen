import { transactionService } from "@/services/transaction-service";
import { useEffect, useRef, useState, type FC } from "react";
import { toast } from "sonner";
import { AppContext, initialState } from "./app-context";

type Props = {
  children?: React.ReactNode;
};

export const AppProvider: FC<Props> = ({ children, ...props }) => {
  const [admin, setAdmin] = useState(initialState.admin);
  const [currentPage, setCurrentPage] = useState(initialState.currentPage);
  const [zkDeviceStatus, setZkDeviceStatus] = useState<boolean>(
    initialState.zkDeviceStatus,
  );
  const [currentUser, setCurrentUser] = useState(initialState.currentUser);
  const [whatsappQR, setWhatsappQR] = useState<string | null>(null);
  const [whatsappStatus, setWhatsappStatus] = useState(
    initialState.whatsappStatus,
  );

  const ws = useRef<WebSocket | null>(null);

  useEffect(() => {
    const connectWebSocket = () => {
      ws.current = new WebSocket("ws://localhost:8080/ws");

      ws.current.onopen = () => {
        console.log("WebSocket connected");
      };

      ws.current.onmessage = async (event) => {
        try {
          const message = JSON.parse(event.data);
          console.log("WebSocket message:", message);

          switch (message.type) {
            case "ping":
              ws.current?.send(JSON.stringify({ type: "pong" }));
              break;
            case "device_status": {
              const { status } = message.payload;
              if (status !== "connected") {
                setZkDeviceStatus(false);
                break;
              }
              setZkDeviceStatus(true);
              break;
            }
            case "attendance_event": {
              console.log("Attendance event:", message);
              const { user_id } = message.payload;
              const user = await transactionService.getUser(user_id);
              if (!user) {
                toast.error("User not found");
                return;
              }
              setCurrentUser(user);
              toast.success("User logged in successfully");
              break;
            }
            case "whatsapp_qr": {
              console.log("WhatsApp QR code received:", message);
              const { qr_code_base64, logged_in } = message.payload;

              if (logged_in) {
                // Already logged in, no need for QR code
                setWhatsappQR(null);
                toast.success("WhatsApp is already logged in");
              } else if (!qr_code_base64 || qr_code_base64 === "") {
                // No QR code or empty QR code
                setWhatsappQR(null);
              } else {
                // Valid QR code received, set it for display
                setWhatsappQR(qr_code_base64);
                toast.info("WhatsApp QR code refreshed. Please scan to login.");
              }
              break;
            }
            case "whatsapp_status": {
              console.log("WhatsApp status:", message);
              const { status, message: statusMessage } = message.payload;
              setWhatsappStatus({
                connected: status === "connected",
                message:
                  statusMessage ||
                  (status === "connected" ? "Connected" : "Disconnected"),
              });

              if (status === "connected") {
                toast.success(statusMessage || "WhatsApp connected");
              } else {
                toast.error(statusMessage || "WhatsApp disconnected");
              }
              break;
            }
          }
        } catch (error) {
          console.error("WebSocket message error:", error);
        }
      };

      ws.current.onclose = () => {
        console.log("WebSocket disconnected");
        // Attempt to reconnect after 5 seconds
        setTimeout(() => {
          if (!ws.current || ws.current.readyState === WebSocket.CLOSED) {
            console.log("Attempting to reconnect...");
            connectWebSocket();
          }
        }, 5000);
      };

      ws.current.onerror = (error) => {
        console.error("WebSocket error:", error);
        if (ws.current) {
          ws.current.close();
        }
      };
    };

    connectWebSocket();

    return () => {
      if (ws.current) {
        ws.current.close();
      }
    };
  }, []);
  useEffect(() => {
    console.log("currentUser", currentUser);
    if (!currentUser?.id) {
      setCurrentPage("screenSaver");
      setAdmin(false);
      return;
    }
    if (
      currentUser &&
      currentUser.id &&
      ["10081", "1023", "10024", "10091"].includes(currentUser.employee_id)
    ) {
      setCurrentPage("canteen");
      setAdmin(true);
    } else {
      setCurrentPage("canteen");
    }
    return;
  }, [admin, currentUser]);
  const value = {
    admin,
    currentPage,
    setCurrentPage,
    currentUser,
    setCurrentUser,
    setAdmin,
    zkDeviceStatus,
    ws,
    whatsappQR,
    whatsappStatus,
  };

  return (
    <AppContext.Provider {...props} value={value}>
      {children}
    </AppContext.Provider>
  );
};
