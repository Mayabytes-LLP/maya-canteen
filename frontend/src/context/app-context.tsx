import { User } from "@/services/transaction-service";
import { createContext } from "react";

export interface AppState {
  admin: boolean;
  currentPage:
    | "canteen"
    | "transactions"
    | "products"
    | "users"
    | "screenSaver";
  currentUser: User | null;
  zkDeviceStatus: boolean;
  whatsappStatus: {
    connected: boolean;
    message: string;
  };
  setCurrentPage: (
    page: "canteen" | "products" | "users" | "screenSaver" | "transactions"
  ) => void;
  setCurrentUser: (user: User | null) => void;
  setAdmin: (admin: boolean) => void;
  ws: React.RefObject<WebSocket | null>;
  whatsappQR: string | null;
}

export const initialState: AppState = {
  admin: false,
  currentPage: "canteen",
  currentUser: null,
  zkDeviceStatus: false,
  whatsappStatus: {
    connected: false,
    message: "Disconnected",
  },
  setCurrentPage: () => null,
  setCurrentUser: () => null,
  setAdmin: () => null,
  ws: { current: null },
  whatsappQR: null,
};

export const AppContext = createContext<AppState>(initialState);
