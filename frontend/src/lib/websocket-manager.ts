export interface WebSocketMessage {
	type: string;
	payload: unknown;
}

export interface ConnectionStats {
	total_connections: number;
	timestamp: number;
}

export interface ReconnectingWebSocketOptions {
	url: string;
	maxAttempts?: number;
	baseDelay?: number;
	maxDelay?: number;
	jitterRange?: number;
	pingInterval?: number;
	pongTimeout?: number;
	onOpen?: () => void;
	onClose?: () => void;
	onError?: (error: Event) => void;
	onMessage?: (message: WebSocketMessage) => void;
	onReconnectAttempt?: (attempt: number, delay: number) => void;
}

export class ReconnectingWebSocket {
	private ws: WebSocket | null = null;
	private url: string;
	private maxAttempts: number;
	private baseDelay: number;
	private maxDelay: number;
	private jitterRange: number;
	private pingInterval: number;
	private pongTimeout: number;

	private reconnectAttempts = 0;
	private reconnectTimeout: number | null = null;
	private pingIntervalId: number | null = null;
	private pongTimeoutId: number | null = null;
	private isManuallyDisconnected = false;
	private lastPongReceived = Date.now();

	// Event handlers
	private onOpen?: () => void;
	private onClose?: () => void;
	private onError?: (error: Event) => void;
	private onMessage?: (message: WebSocketMessage) => void;
	private onReconnectAttempt?: (attempt: number, delay: number) => void;

	constructor(options: ReconnectingWebSocketOptions) {
		this.url = options.url;
		this.maxAttempts = options.maxAttempts ?? 10;
		this.baseDelay = options.baseDelay ?? 500; // 500ms
		this.maxDelay = options.maxDelay ?? 30000; // 30 seconds
		this.jitterRange = options.jitterRange ?? 1000; // 1 second
		this.pingInterval = options.pingInterval ?? 30000; // 30 seconds
		this.pongTimeout = options.pongTimeout ?? 60000; // 60 seconds

		this.onOpen = options.onOpen;
		this.onClose = options.onClose;
		this.onError = options.onError;
		this.onMessage = options.onMessage;
		this.onReconnectAttempt = options.onReconnectAttempt;

		this.connect();
	}

	private connect(): void {
		if (this.isManuallyDisconnected) {
			return;
		}

		console.log(
			`Attempting WebSocket connection to ${this.url} (attempt ${this.reconnectAttempts + 1})`,
		);

		try {
			this.ws = new WebSocket(this.url);
			this.setupEventHandlers();
		} catch (error) {
			console.error("Failed to create WebSocket connection:", error);
			this.scheduleReconnect();
		}
	}

	private setupEventHandlers(): void {
		if (!this.ws) return;

		this.ws.onopen = () => {
			console.log("WebSocket connected successfully");
			this.reconnectAttempts = 0; // Reset on successful connection
			this.lastPongReceived = Date.now();
			this.startPingInterval();
			this.onOpen?.();
		};

		this.ws.onclose = (event) => {
			console.log(`WebSocket disconnected: ${event.code} - ${event.reason}`);
			this.cleanup();

			if (!this.isManuallyDisconnected) {
				this.scheduleReconnect();
			}

			this.onClose?.();
		};

		this.ws.onerror = (error) => {
			console.error("WebSocket error:", error);
			this.onError?.(error);
		};

		this.ws.onmessage = (event) => {
			try {
				const message: WebSocketMessage = JSON.parse(event.data);

				// Handle ping/pong messages for connection health
				if (message.type === "ping") {
					this.send({ type: "pong" });
					return;
				} else if (message.type === "pong") {
					this.lastPongReceived = Date.now();
					return;
				}

				this.onMessage?.(message);
			} catch (error) {
				console.error("Failed to parse WebSocket message:", error);
			}
		};
	}

	private startPingInterval(): void {
		this.stopPingInterval();

		this.pingIntervalId = window.setInterval(() => {
			if (this.isConnected()) {
				this.send({ type: "ping" });

				// Start pong timeout
				this.pongTimeoutId = window.setTimeout(() => {
					console.warn("Pong timeout - connection appears dead");
					this.disconnect(false); // Force reconnection
				}, this.pongTimeout);
			}
		}, this.pingInterval);
	}

	private stopPingInterval(): void {
		if (this.pingIntervalId) {
			clearInterval(this.pingIntervalId);
			this.pingIntervalId = null;
		}

		if (this.pongTimeoutId) {
			clearTimeout(this.pongTimeoutId);
			this.pongTimeoutId = null;
		}
	}

	private scheduleReconnect(): void {
		if (
			this.isManuallyDisconnected ||
			this.reconnectAttempts >= this.maxAttempts
		) {
			if (this.reconnectAttempts >= this.maxAttempts) {
				console.error(
					`Max reconnection attempts (${this.maxAttempts}) reached. Giving up.`,
				);
			}
			return;
		}

		const delay = this.calculateBackoffDelay(this.reconnectAttempts);
		console.log(
			`Scheduling reconnection in ${delay}ms (attempt ${this.reconnectAttempts + 1}/${this.maxAttempts})`,
		);

		this.onReconnectAttempt?.(this.reconnectAttempts + 1, delay);

		this.reconnectTimeout = window.setTimeout(() => {
			this.reconnectAttempts++;
			this.connect();
		}, delay);
	}

	private calculateBackoffDelay(attempt: number): number {
		// Exponential backoff with jitter
		const exponentialDelay = this.baseDelay * 2 ** attempt;
		const jitter = Math.random() * this.jitterRange;
		const delay = Math.min(exponentialDelay + jitter, this.maxDelay);
		return Math.floor(delay);
	}

	private cleanup(): void {
		this.stopPingInterval();

		if (this.reconnectTimeout) {
			clearTimeout(this.reconnectTimeout);
			this.reconnectTimeout = null;
		}
	}

	// Public methods
	public send(message: WebSocketMessage | Record<string, unknown>): boolean {
		if (!this.isConnected()) {
			console.warn("Cannot send message: WebSocket not connected");
			return false;
		}

		try {
			const messageStr =
				typeof message === "string" ? message : JSON.stringify(message);
			if (this.ws) {
				this.ws.send(messageStr);
			}
			return true;
		} catch (error) {
			console.error("Failed to send WebSocket message:", error);
			return false;
		}
	}

	public isConnected(): boolean {
		return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
	}

	public getReadyState(): number | null {
		return this.ws?.readyState ?? null;
	}

	public disconnect(manual: boolean = true): void {
		this.isManuallyDisconnected = manual;
		this.cleanup();

		if (this.ws) {
			this.ws.close(1000, "Manual disconnect");
			this.ws = null;
		}

		if (!manual) {
			// If not manual, allow reconnection
			this.isManuallyDisconnected = false;
		}

		console.log(`WebSocket ${manual ? "manually" : "forcibly"} disconnected`);
	}

	public reconnect(): void {
		console.log("Manual reconnection requested");
		this.disconnect(false);
		this.reconnectAttempts = 0; // Reset attempts on manual reconnect
		this.connect();
	}

	public getConnectionStats(): {
		connected: boolean;
		readyState: number | null;
		reconnectAttempts: number;
		lastPongReceived: number;
		url: string;
	} {
		return {
			connected: this.isConnected(),
			readyState: this.getReadyState(),
			reconnectAttempts: this.reconnectAttempts,
			lastPongReceived: this.lastPongReceived,
			url: this.url,
		};
	}
}
