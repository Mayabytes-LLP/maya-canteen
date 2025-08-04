import { createContext } from "react";
import type { User } from "@/services/transaction-service";

export interface AppState {
	admin: boolean;
	currentUser: User | null;
	zkDeviceStatus: boolean;
	whatsappStatus: {
		connected: boolean;
		message: string;
	};
	whatsappClientInfo?: Record<string, unknown> | null;
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
	whatsappClientInfo: null,
	setCurrentUser: () => null,
	setAdmin: () => null,
	ws: { current: null },
	whatsappQR: null,
};

export const AppContext = createContext<AppState>(initialState);
