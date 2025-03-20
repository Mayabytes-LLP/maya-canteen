import { transactionService, User } from "@/services/transaction-service";
import { createContext, useEffect, useRef, useState, type FC } from "react";
import { toast } from "sonner";

export interface AppState {
  admin: boolean;
  currentPage:
    | "canteen"
    | "transactions"
    | "products"
    | "users"
    | "screenSaver";
  currentUser: User | null;
  setCurrentPage: (
    page: "canteen" | "products" | "users" | "screenSaver" | "transactions"
  ) => void;
  setCurrentUser: (user: User | null) => void;
  setAdmin: (admin: boolean) => void;
  ws: React.RefObject<WebSocket | null>;
}

export const initialState: AppState = {
  admin: false,
  currentPage: "canteen",
  currentUser: null,
  setCurrentPage: () => null,
  setCurrentUser: () => null,
  setAdmin: () => null,
  ws: { current: null },
};

export const AppContext = createContext<AppState>(initialState);

type Props = {
  children?: React.ReactNode;
};

export const AppProvider: FC<Props> = ({ children, ...props }) => {
  const [admin, setAdmin] = useState(initialState.admin);
  const [currentPage, setCurrentPage] = useState(initialState.currentPage);
  const [currentUser, setCurrentUser] = useState(initialState.currentUser);
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

  const value = {
    admin,
    currentPage,
    setCurrentPage,
    currentUser,
    setCurrentUser,
    setAdmin,
    ws,
  };

  const admins = ["10058", "10037", "10024"];

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
      admins.includes(currentUser.employee_id)
    ) {
      setCurrentPage("canteen");
      setAdmin(true);
    } else {
      setCurrentPage("canteen");
    }
    return;
  }, [admin, currentUser]);

  useEffect(() => {}, [currentUser]);

  return (
    <AppContext.Provider {...props} value={value}>
      {children}
    </AppContext.Provider>
  );
};
