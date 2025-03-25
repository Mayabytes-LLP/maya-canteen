import { User } from "@/services/transaction-service";
import { createContext } from "react";

export interface AppState {
  admin: boolean;
  currentUser: User | null;
  zkDeviceStatus: boolean;
  whatsappStatus: {
    connected: boolean;
    message: string;
  };
  setCurrentUser: (user: User | null) => void;
  setAdmin: (admin: boolean) => void;
  ws: React.RefObject<WebSocket | null>;
  whatsappQR: string | null;
}

export const initialState: AppState = {
  admin: false,
  currentUser: null,
  zkDeviceStatus: false,
  whatsappStatus: {
    connected: false,
    message: "Disconnected",
  },
  setCurrentUser: () => null,
  setAdmin: () => null,
  ws: { current: null },
  whatsappQR: null,
};

export const AppContext = createContext<AppState>(initialState);
