import { CheckCircle2, LoaderCircle, Phone, QrCode, RefreshCw } from "lucide-react";
import { QRCodeSVG } from "qrcode.react";
import { useContext, useEffect, useState } from "react";
import { toast } from "sonner";
import DepositForm from "@/components/deposit-form";
import ErrorBoundary from "@/components/error-boundary";
import TransactionList from "@/components/transaction-list";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardFooter,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { AppContext } from "@/context";

function statusBadgeVariant(status: {
	connected: boolean;
	message: string;
}): "default" | "secondary" | "destructive" | "outline" {
	if (status.connected) return "default";
	if (status.message?.toLowerCase().includes("connecting")) return "secondary";
	return "destructive";
}

function statusBadgeClass(status: {
	connected: boolean;
	message: string;
}): string {
	if (status.connected) return "bg-green-600 text-white";
	if (status.message?.toLowerCase().includes("connecting")) return "";
	return "bg-red-600 text-white";
}

function statusLabel(status: { connected: boolean; message: string }): string {
	if (status.connected) return "Connected";
	if (status.message?.toLowerCase().includes("connecting")) return "Connecting…";
	return "Disconnected";
}

export default function TransactionsPage() {
	const [refreshTrigger, setRefreshTrigger] = useState(0);
	const [transactionLimit, setTransactionLimit] = useState(50);
	const [inputLimit, setInputLimit] = useState("10");

	const {
		whatsappQR,
		whatsappPairingCode,
		whatsappStatus,
		ws,
		whatsappClientInfo,
	} = useContext(AppContext);

	const [pairingMode, setPairingMode] = useState<"qr" | "phone">("qr");
	const [phoneNumber, setPhoneNumber] = useState("");
	const [isPairing, setIsPairing] = useState(false);
	const [isRefreshing, setIsRefreshing] = useState(false);

	useEffect(() => {
		if (whatsappStatus.connected || whatsappQR || whatsappPairingCode) {
			setIsRefreshing(false);
		}
	}, [whatsappStatus.connected, whatsappQR, whatsappPairingCode]);

	useEffect(() => {
		if (whatsappPairingCode || whatsappStatus.connected) {
			setIsPairing(false);
		}
	}, [whatsappPairingCode, whatsappStatus.connected]);

	const handleTransactionAdded = () => {
		setRefreshTrigger((prev) => prev + 1);
	};

	const handleLimitChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		setInputLimit(e.target.value);
	};

	const applyLimit = () => {
		const limit = parseInt(inputLimit);
		if (!Number.isNaN(limit) && limit > 0) {
			setTransactionLimit(limit);
		} else {
			setInputLimit(transactionLimit.toString());
		}
	};

	const pairWithPhone = () => {
		const phone = phoneNumber.trim();
		if (!phone) {
			toast.error("Please enter a phone number");
			return;
		}
		if (ws.current?.isConnected()) {
			setIsPairing(true);
			const success = ws.current.send({
				type: "pair_phone",
				phone: phone,
			});
			if (success) {
				toast.info("Initiating phone pairing...");
			} else {
				toast.error("Failed to send pairing request");
				setIsPairing(false);
			}
		} else {
			toast.error("WebSocket not connected");
		}
	};

	const refreshWhatsApp = () => {
		if (whatsappStatus.connected) {
			toast.info("WhatsApp is already connected.");
			return;
		}
		if (ws.current?.isConnected()) {
			setIsRefreshing(true);
			const success = ws.current.send({ type: "refresh_whatsapp" });
			if (!success) {
				toast.error("Failed to send refresh command");
				setIsRefreshing(false);
			}
		} else {
			toast.error("WebSocket not connected");
		}
	};

	const isConnecting =
		isRefreshing ||
		isPairing ||
		whatsappStatus.message?.toLowerCase().includes("connecting");

	return (
		<div className="container mx-auto py-6 space-y-6">
			<h1 className="text-3xl font-bold">Transactions Management</h1>

			<div className="grid grid-cols-1 md:grid-cols-2 gap-6">
				<div>
					<DepositForm onTransactionAdded={handleTransactionAdded} />
				</div>
				<div className="space-y-4">
					<Card>
						<CardHeader>
							<div className="flex justify-between items-center">
								<div>
									<CardTitle>WebSocket Connection</CardTitle>
									<CardDescription>Real-time connection status</CardDescription>
								</div>
								<Badge
									variant={
										ws.current?.isConnected() ? "default" : "destructive"
									}
									className={`${
										ws.current?.isConnected() ? "bg-green-600" : "bg-red-600"
									} text-white`}
								>
									{ws.current?.isConnected() ? "Connected" : "Disconnected"}
								</Badge>
							</div>
						</CardHeader>
						<CardContent>
							{ws.current && (
								<div className="space-y-2 text-sm">
									<div className="flex justify-between">
										<span>Ready State:</span>
										<span className="font-mono">
											{ws.current.getReadyState() ?? "Unknown"}
										</span>
									</div>
									<div className="flex justify-between">
										<span>Reconnect Attempts:</span>
										<span className="font-mono">
											{ws.current.getConnectionStats().reconnectAttempts}
										</span>
									</div>
									<div className="flex justify-between">
										<span>Last Activity:</span>
										<span className="font-mono">
											{new Date(
												ws.current.getConnectionStats().lastPongReceived,
											).toLocaleTimeString()}
										</span>
									</div>
								</div>
							)}
						</CardContent>
						<CardFooter className="flex gap-2">
							<Button
								variant="outline"
								size="sm"
								onClick={() => ws.current?.reconnect()}
								disabled={ws.current?.isConnected()}
							>
								<RefreshCw className="h-4 w-4 mr-2" />
								Reconnect
							</Button>
							<Button
								variant="outline"
								size="sm"
								onClick={() => {
									if (ws.current) {
										const stats = ws.current.getConnectionStats();
										console.log("WebSocket stats:", stats);
										toast.info("Check console for connection details");
									}
								}}
							>
								Debug
							</Button>
						</CardFooter>
					</Card>

					<Card>
						<CardHeader>
							<div className="flex justify-between items-center">
								<div>
									<CardTitle>WhatsApp Connection</CardTitle>
									<CardDescription>
										{whatsappPairingCode
											? "Enter the pairing code on your phone"
											: whatsappQR
												? "Scan the QR code to link WhatsApp"
												: whatsappStatus.connected
													? "WhatsApp is ready to send messages"
													: isConnecting
														? whatsappStatus.message || "Connecting…"
														: "Connect WhatsApp via QR code or phone number"}
									</CardDescription>
								</div>
								<Badge
									variant={statusBadgeVariant(whatsappStatus)}
									className={statusBadgeClass(whatsappStatus)}
								>
									{statusLabel(whatsappStatus)}
								</Badge>
							</div>
						</CardHeader>
						<CardContent className="flex flex-col items-center min-h-[200px] justify-center">
							{whatsappStatus.connected ? (
								<div className="text-center p-4 space-y-3">
									<CheckCircle2 className="h-10 w-10 text-green-500 mx-auto" />
									<p className="font-medium text-green-700 dark:text-green-400">
										WhatsApp is connected
									</p>
									{whatsappClientInfo && (
										<div className="text-xs text-muted-foreground space-y-0.5 mt-2">
											{whatsappClientInfo.connected === true && (
												<>
													{whatsappClientInfo.user != null && (
														<div>
															<span className="font-medium">Phone:</span>{" "}
															<span className="font-mono">
																{String(whatsappClientInfo.user)}
															</span>
														</div>
													)}
													{whatsappClientInfo.platform != null && (
														<div>
															<span className="font-medium">Device:</span>{" "}
															{String(whatsappClientInfo.platform)}
														</div>
													)}
												</>
											)}
										</div>
									)}
								</div>
							) : isConnecting && !whatsappQR && !whatsappPairingCode ? (
								<div className="text-center p-4 space-y-3">
									<LoaderCircle className="h-10 w-10 text-muted-foreground animate-spin mx-auto" />
									<p className="font-medium text-muted-foreground">
										{whatsappStatus.message || "Connecting to WhatsApp…"}
									</p>
									<p className="text-xs text-muted-foreground/60">
										Please wait while we establish a connection
									</p>
								</div>
							) : whatsappPairingCode ? (
								<div className="flex flex-col items-center p-6 border rounded-lg w-full max-w-sm">
									<Phone className="h-8 w-8 text-blue-500 mb-2" />
									<p className="text-sm text-muted-foreground mb-4 text-center">
										Enter this code on your phone to link WhatsApp:
									</p>
									<div className="bg-muted px-8 py-4 rounded-md w-full">
										<p className="text-4xl font-mono font-bold tracking-[0.3em] text-center select-all">
											{whatsappPairingCode}
										</p>
									</div>
									<div className="mt-4 text-xs text-muted-foreground space-y-1">
										<p className="text-center">
											Open <span className="font-medium">WhatsApp</span> →{" "}
											<span className="font-medium">Linked Devices</span> →{" "}
											<span className="font-medium">
												Link with phone number
											</span>
										</p>
									</div>
								</div>
							) : whatsappQR ? (
								<div className="flex flex-col items-center w-full p-4">
									<div className="bg-white p-3 rounded-xl shadow-md border">
										<QRCodeSVG
											value={whatsappQR}
											size={280}
											level="Q"
											bgColor="#ffffff"
											fgColor="#000000"
											marginSize={8}
										/>
									</div>
									<p className="mt-4 text-sm text-muted-foreground">
										Scan with WhatsApp to link this device
									</p>
									<p className="text-xs text-muted-foreground/60 mt-1">
										Open WhatsApp → Linked Devices → Link a Device
									</p>
								</div>
							) : (
								<>
									{!whatsappStatus.connected && (
										<div className="flex gap-2 mb-6 w-full max-w-xs">
											<Button
												variant={pairingMode === "qr" ? "default" : "outline"}
												size="sm"
												className="flex-1"
												onClick={() => setPairingMode("qr")}
											>
												<QrCode className="h-4 w-4 mr-1" />
												QR Code
											</Button>
											<Button
												variant={
													pairingMode === "phone" ? "default" : "outline"
												}
												size="sm"
												className="flex-1"
												onClick={() => setPairingMode("phone")}
											>
												<Phone className="h-4 w-4 mr-1" />
												Phone Number
											</Button>
										</div>
									)}

									{pairingMode === "phone" && !whatsappStatus.connected ? (
										<div className="flex flex-col items-center gap-4 w-full max-w-sm">
											<p className="text-sm text-muted-foreground text-center">
												Enter your phone number to receive a pairing code
											</p>
											<div className="flex gap-2 w-full">
												<Input
													type="tel"
													placeholder="923001234567"
													value={phoneNumber}
													onChange={(e) => setPhoneNumber(e.target.value)}
													disabled={isPairing}
													className="flex-1"
												/>
												<Button
													onClick={pairWithPhone}
													disabled={isPairing || !phoneNumber.trim()}
												>
													{isPairing ? (
														<>
															<LoaderCircle className="h-4 w-4 mr-1 animate-spin" />
															Pairing…
														</>
													) : (
														"Link"
													)}
												</Button>
											</div>
											<p className="text-xs text-muted-foreground text-center">
												Use international format without + or leading 0 (e.g.,{" "}
												<code className="font-mono">923001234567</code>)
											</p>
										</div>
									) : null}
								</>
							)}
						</CardContent>
						<CardFooter className="flex justify-center">
							<Button
								variant="outline"
								onClick={refreshWhatsApp}
								disabled={isConnecting}
							>
								{isConnecting ? (
									<>
										<LoaderCircle className="h-4 w-4 mr-2 animate-spin" />
										Connecting…
									</>
								) : (
									<>
										<RefreshCw className="h-4 w-4 mr-2" />
										{whatsappStatus.connected
											? "Refresh"
											: "Connect WhatsApp"}
									</>
								)}
							</Button>
						</CardFooter>
					</Card>
				</div>
			</div>

			<div className="space-y-4">
				<div className="flex items-end gap-2">
					<div className="space-y-2">
						<Label htmlFor="limit">Number of transactions to show</Label>
						<Input
							id="limit"
							type="number"
							min="1"
							value={inputLimit}
							onChange={handleLimitChange}
							className="w-24"
						/>
					</div>
					<Button onClick={applyLimit}>Apply</Button>
				</div>
				<ErrorBoundary>
					<TransactionList
						limit={transactionLimit}
						refreshTrigger={refreshTrigger}
					/>
				</ErrorBoundary>
			</div>
		</div>
	);
}
