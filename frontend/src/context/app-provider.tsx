import { type FC, useEffect, useRef, useState } from "react";
import { useLocation, useNavigate } from "react-router";
import { toast } from "sonner";
import {
	ReconnectingWebSocket,
	type WebSocketMessage,
} from "@/lib/websocket-manager";
import { transactionService } from "@/services/transaction-service";
import { AppContext, initialState } from "./app-context";

type Props = {
	children?: React.ReactNode;
};

const VITE_WS_URL = import.meta.env.VITE_WS_URL || "ws://localhost:8080/ws";
export const AppProvider: FC<Props> = ({ children, ...props }) => {
	const [admin, setAdmin] = useState(initialState.admin);
	const [zkDeviceStatus, setZkDeviceStatus] = useState<boolean>(
		initialState.zkDeviceStatus,
	);
	const [currentUser, setCurrentUser] = useState(initialState.currentUser);
	const [whatsappQR, setWhatsappQR] = useState<string | null>(null);
	const [whatsappStatus, setWhatsappStatus] = useState(
		initialState.whatsappStatus,
	);
	interface WhatsappClientInfo {
		id?: string;
		name?: string;
		phoneNumber?: string;
		[key: string]: unknown; // Allow additional properties
	}

	const [whatsappClientInfo, setWhatsappClientInfo] =
		useState<WhatsappClientInfo | null>(null);
	const location = useLocation();
	const navigate = useNavigate();

	const ws = useRef<ReconnectingWebSocket | null>(null);

	useEffect(() => {
		const connectWebSocket = () => {
			// Prevent duplicate connections
			if (ws.current) {
				console.warn(
					"WebSocket is already initialized. Skipping new connection.",
				);
				return;
			}

			ws.current = new ReconnectingWebSocket({
				url: VITE_WS_URL,
				maxAttempts: 10,
				baseDelay: 1000,
				maxDelay: 30000,
				jitterRange: 2000,
				pingInterval: 30000,
				pongTimeout: 60000,
				onOpen: () => {
					console.log("WebSocket connected successfully");
					toast.success("Connected to server", { duration: 2000 });
				},
				onClose: () => {
					console.log("WebSocket disconnected");
					toast.warning("Disconnected from server", { duration: 3000 });
				},
				onError: (error) => {
					console.error("WebSocket error:", error);
					toast.error("Connection error occurred", { duration: 3000 });
				},
				onReconnectAttempt: (attempt, delay) => {
					console.log(`Reconnection attempt ${attempt} in ${delay}ms`);
					if (attempt === 1) {
						toast.info("Attempting to reconnect...", { duration: 2000 });
					} else if (attempt % 3 === 0) {
						toast.warning(`Reconnection attempt ${attempt}...`, {
							duration: 2000,
						});
					}
				},
				onMessage: async (message: WebSocketMessage) => {
					try {
						console.log("WebSocket message:", message);

						switch (message.type) {
							case "device_status": {
								const { status } = message.payload as { status: string };
								if (status !== "connected") {
									setZkDeviceStatus(false);
									break;
								}
								setZkDeviceStatus(true);
								break;
							}
							case "attendance_event": {
								console.log("Attendance event:", message);
								const { user_id } = message.payload as { user_id: string };
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
								const { qr_code_base64, logged_in } = message.payload as {
									qr_code_base64: string;
									logged_in: boolean;
								};

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
									toast.info(
										"WhatsApp QR code refreshed. Please scan to login.",
									);
								}
								break;
							}
							case "whatsapp_status": {
								console.log("WhatsApp status:", message);
								const {
									status,
									message: statusMessage,
									client_info,
								} = message.payload as {
									status: string;
									message: string;
									client_info: Record<string, unknown>;
								};
								setWhatsappStatus({
									connected: status === "connected",
									message:
										statusMessage ||
										(status === "connected" ? "Connected" : "Disconnected"),
								});
								setWhatsappClientInfo(client_info || null);

								if (status === "connected") {
									toast.success(statusMessage || "WhatsApp connected");
								} else {
									toast.error(statusMessage || "WhatsApp disconnected");
								}
								break;
							}
							case "connection_status": {
								// Handle connection status updates from server
								const { total_connections } = message.payload as {
									total_connections: number;
								};
								console.log(
									`Server reports ${total_connections} active connections`,
								);
								break;
							}
							case "connected": {
								// Initial connection message from server
								const { message: msg, client_id } = message.payload as {
									message: string;
									client_id: string;
								};
								console.log(
									`Connected to server: ${msg}, client ID: ${client_id}`,
								);
								break;
							}
						}
					} catch (error) {
						console.error("WebSocket message error:", error);
					}
				},
			});
		};

		connectWebSocket();

		return () => {
			if (ws.current) {
				ws.current.disconnect(true);
				ws.current = null;
			}
		};
	}, []);
	useEffect(() => {
		if (!currentUser?.id) {
			if (location.pathname !== "/login") {
				navigate("/login");
			}
			setAdmin(false);
			return;
		}
		if (
			currentUser?.id &&
			["10081", "1023", "10024", "10091"].includes(currentUser.employee_id)
		) {
			setAdmin(true);
		}
		return;
	}, [currentUser, location.pathname, navigate]);

	const value = {
		admin,
		currentUser,
		setCurrentUser,
		setAdmin,
		zkDeviceStatus,
		ws,
		whatsappQR,
		whatsappStatus,
		whatsappClientInfo,
	};

	return (
		<AppContext.Provider {...props} value={value}>
			{children}
		</AppContext.Provider>
	);
};
