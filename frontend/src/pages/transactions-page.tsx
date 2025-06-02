import { useContext, useState } from "react";

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
import { LoaderCircle, RefreshCw } from "lucide-react";
import { QRCodeSVG } from "qrcode.react";
import { toast } from "sonner";

export default function TransactionsPage() {
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  const [transactionLimit, setTransactionLimit] = useState(50);
  const [inputLimit, setInputLimit] = useState("10");
  const [isRefreshing, setIsRefreshing] = useState(false);

  const { whatsappQR, whatsappStatus, ws } = useContext(AppContext);

  // Function to trigger a refresh of the transaction list
  const handleTransactionAdded = () => {
    setRefreshTrigger((prev) => prev + 1);
  };

  // Handle limit change
  const handleLimitChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setInputLimit(e.target.value);
  };

  // Apply the new limit
  const applyLimit = () => {
    const limit = parseInt(inputLimit);
    if (!isNaN(limit) && limit > 0) {
      setTransactionLimit(limit);
    } else {
      setInputLimit(transactionLimit.toString());
    }
  };

  // Function to manually refresh the WhatsApp connection
  const refreshWhatsApp = () => {
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      setIsRefreshing(true);
      // Send refresh command to backend
      ws.current.send(JSON.stringify({ type: "refresh_whatsapp" }));
      toast.info("Connecting to WhatsApp...");

      // Set a timeout to reset the refreshing state after a reasonable time
      setTimeout(() => {
        setIsRefreshing(false);
      }, 30000); // Reset after 30 seconds max
    } else {
      toast.error("WebSocket not connected");
    }
  };

  return (
    <div className="container mx-auto py-6 space-y-6">
      <h1 className="text-3xl font-bold">Transactions Management</h1>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div>
          <DepositForm onTransactionAdded={handleTransactionAdded} />
        </div>
        <div>
          <Card>
            <CardHeader>
              <div className="flex justify-between items-center">
                <div>
                  <CardTitle>WhatsApp Connection</CardTitle>
                  <CardDescription>
                    {whatsappQR
                      ? "Scan the QR code to login with WhatsApp"
                      : whatsappStatus.connected
                      ? "WhatsApp is connected and ready to send messages"
                      : "Click refresh to connect WhatsApp"}
                  </CardDescription>
                </div>
                <Badge
                  variant={whatsappStatus.connected ? "default" : "destructive"}
                  className={`${
                    whatsappStatus.connected ? "bg-green-600" : "bg-red-600"
                  } text-white`}
                >
                  {whatsappStatus.connected ? "Connected" : "Disconnected"}
                </Badge>
              </div>
            </CardHeader>
            <CardContent className="flex flex-col items-center">
              {whatsappQR ? (
                <div className="flex flex-col items-center w-full bg-white p-4 rounded-lg">
                  <div className="p-4 rounded-lg shadow-lg">
                    <QRCodeSVG
                      value={whatsappQR}
                      size={400}
                      level="L"
                      bgColor="#ffffff"
                      fgColor="#000000"
                      marginSize={10}
                    />
                  </div>
                  <p className="mt-4 text-sm text-muted-foreground">
                    Scan with WhatsApp to login
                  </p>
                </div>
              ) : whatsappStatus.connected ? (
                <div className="text-center p-4"></div>
              ) : (
                <div className="text-center p-4">
                  <p className="text-lg">WhatsApp QR Code not available</p>
                  <p className="text-sm text-muted-foreground mt-1">
                    {whatsappStatus.message ||
                      "WhatsApp authentication is pending."}
                  </p>
                </div>
              )}
            </CardContent>
            <CardFooter className="flex justify-center">
              <Button
                variant="outline"
                onClick={refreshWhatsApp}
                className="flex gap-2 items-center"
                disabled={isRefreshing}
              >
                {isRefreshing ? (
                  <>
                    <LoaderCircle className="h-4 w-4 animate-spin" />
                    Connecting...
                  </>
                ) : (
                  <>
                    <RefreshCw className="h-4 w-4" />
                    {whatsappStatus.connected
                      ? "Refresh Connection"
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
