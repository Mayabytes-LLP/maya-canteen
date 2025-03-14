import { User } from "@/services/transaction-service";
import { createContext, useEffect, useRef, useState, type FC } from "react";

export interface AppState {
  admin: boolean;
  currentPage: "canteen" | "products" | "users" | "screenSaver";
  currentUser: User | null;
  setCurrentPage: (
    page: "canteen" | "products" | "users" | "screenSaver"
  ) => void;
  setCurrentUser: (user: User) => void;
  setAdmin: (admin: boolean) => void;
}

export const initialState: AppState = {
  admin: false,
  currentPage: "canteen",
  currentUser: null,
  setCurrentPage: () => null,
  setCurrentUser: () => null,
  setAdmin: () => null,
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

      ws.current.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data);
          console.log("WebSocket message:", message);

          switch (message.type) {
            case "ping":
              ws.current?.send(JSON.stringify({ type: "pong" }));
              break;
            case "user_data":
              // Handle user data updates
              if (message.payload) {
                setCurrentUser(message.payload);
              }
              break;
          }
        } catch (error) {
          console.error("WebSocket message error:", error);
        }
      };

      ws.current.onclose = () => {
        console.log("WebSocket disconnected");
        // Attempt to reconnect after 5 seconds
        setTimeout(() => {
          console.log("Attempting to reconnect...");
          if (!ws.current || ws.current.readyState === WebSocket.CLOSED) {
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
  };

  const admins = ["00058", "00092"];

  useEffect(() => {
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
