import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { AppContext } from "@/context";
import { transactionService } from "@/services/transaction-service";
import { MessageCircle } from "lucide-react";
import { useContext, useState } from "react";
import { toast } from "sonner";

export default function SendAllBalance() {
  const { whatsappStatus } = useContext(AppContext);

  const [allMessageDialogOpen, setAllMessageDialogOpen] = useState(false);
  // Add state for month and year selection
  const [selectedMonth, setSelectedMonth] = useState<string>(() => {
    const now = new Date();
    return now.toLocaleString("default", { month: "long" });
  });
  const [selectedYear, setSelectedYear] = useState<number>(() => {
    return new Date().getFullYear();
  });
  const [sendingAllNotifications, setSendingAllNotifications] = useState(false);

  const [selectedDuration, setSelectedDuration] =
    useState<string>("Half month");
  const months = [
    "January",
    "February",
    "March",
    "April",
    "May",
    "June",
    "July",
    "August",
    "September",
    "October",
    "November",
    "December",
  ];
  const years = Array.from(
    { length: 5 },
    (_, i) => new Date().getFullYear() - 2 + i
  );
  const defaultTemplate =
    "**Balance Update** \n\nDear {name},\nYour current canteen balance is: *PKR {balance}*\n\nPlease pay online via Jazz Cash 03422949447 (Syed Kazim Raza) {duration} of Canteen bill for {month} {year}\n\nThis is an automated message from Maya Canteen Management System.";

  const [allMessageTemplate, setAllMessageTemplate] =
    useState<string>(defaultTemplate);

  // Function to send balance notifications to all users
  const sendAllBalanceNotifications = async (
    messageTemplate?: string,
    month?: string,
    year?: number
  ) => {
    setSendingAllNotifications(true);
    try {
      const response = await transactionService.sendAllBalanceNotifications(
        messageTemplate,
        month,
        year
      );
      if (response.success && response.data) {
        // Show success message with details
        toast.success(
          <div className="space-y-2">
            <p>{response.data.message}</p>
            {response.data.details.failed_users.length > 0 && (
              <div className="text-sm">
                <p className="font-semibold">Failed Users:</p>
                <ul className="list-disc pl-4">
                  {response.data.details.failed_users.map(
                    (user: string, index: number) => (
                      <li key={index}>{user}</li>
                    )
                  )}
                </ul>
              </div>
            )}
          </div>
        );
      } else {
        toast.error("Failed to send balance notifications to all users");
      }
    } catch (error) {
      console.error("Error sending all balance notifications:", error);
      toast.error("Failed to send balance notifications to all users");
    } finally {
      setSendingAllNotifications(false);
    }
  };
  return (
    <Dialog open={allMessageDialogOpen} onOpenChange={setAllMessageDialogOpen}>
      <DialogTrigger asChild>
        <Button
          onClick={() => setAllMessageDialogOpen(true)}
          disabled={sendingAllNotifications || !whatsappStatus.connected}
          className="flex items-center gap-2"
        >
          {sendingAllNotifications ? (
            "Sending..."
          ) : (
            <>
              <MessageCircle className="h-4 w-4" /> Send All Balances
            </>
          )}
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit WhatsApp Message for All Users</DialogTitle>
        </DialogHeader>
        <DialogDescription>
          You can edit the message template that will be sent to all users. The
          message will be sent to all users with a balance.
        </DialogDescription>
        <DialogClose />
        <div className="space-y-2">
          <div className="text-xs text-muted-foreground">
            You can use <code>{"{name}"}</code>, <code>{"{balance}"}</code>,{" "}
            <code>{"{month}"}</code> and <code>{"{year}"}</code> as
            placeholders.
          </div>
          <div className="flex gap-2">
            <Select
              value={selectedMonth}
              onValueChange={(value) => setSelectedMonth(value)}
            >
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Select Month" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  {months.map((month) => (
                    <SelectItem key={month} value={month}>
                      {month}
                    </SelectItem>
                  ))}
                </SelectGroup>
              </SelectContent>
            </Select>
            <Select
              value={selectedYear.toString()}
              onValueChange={(value) => setSelectedYear(parseInt(value, 10))}
            >
              <SelectTrigger className="w-[100px]">
                <SelectValue placeholder="Year" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  {years.map((year) => (
                    <SelectItem key={year} value={year.toString()}>
                      {year}
                    </SelectItem>
                  ))}
                </SelectGroup>
              </SelectContent>
            </Select>

            <Select
              value={selectedDuration}
              onValueChange={(value) => setSelectedDuration(value)}
            >
              <SelectTrigger className="w-[100px]">
                <SelectValue placeholder="Duration" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem key="half-month" value="Half month">
                    Half Month
                  </SelectItem>
                  <SelectItem key="full-month" value="Full month">
                    Full Month
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </div>
          <Textarea
            rows={6}
            value={allMessageTemplate}
            onChange={(e) => setAllMessageTemplate(e.target.value)}
            className="w-full min-h-[120px] font-mono"
          />
        </div>
        <DialogFooter>
          <Button
            onClick={() => {
              const finalMessage = allMessageTemplate
                .replace(/\{month\}/g, selectedMonth)
                .replace(/\{year\}/g, selectedYear.toString())
                .replace(/\{duration\}/g, selectedDuration);
              sendAllBalanceNotifications(
                finalMessage,
                selectedMonth,
                selectedYear
              );
              setAllMessageDialogOpen(false);
            }}
            disabled={sendingAllNotifications}
          >
            Send Notifications
          </Button>
          <Button
            variant="outline"
            onClick={() => setAllMessageDialogOpen(false)}
          >
            Cancel
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
