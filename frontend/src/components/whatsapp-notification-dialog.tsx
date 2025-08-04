import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import {
	Select,
	SelectContent,
	SelectGroup,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import {
	DEFAULT_WHATSAPP_MESSAGE_TEMPLATE,
	months,
} from "@/constants/whatsapp-message-template";

interface WhatsAppNotificationDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	onSend: (
		messageTemplate: string,
		month: string,
		year: number,
		includeTransactions: boolean,
	) => void;
	sendingNotification: boolean;
}

export function WhatsAppNotificationDialog({
	open,
	onOpenChange,
	onSend,
	sendingNotification,
}: WhatsAppNotificationDialogProps) {
	const [messageTemplate, setMessageTemplate] = useState<string>(
		() => DEFAULT_WHATSAPP_MESSAGE_TEMPLATE,
	);
	const [selectedMonth, setSelectedMonth] = useState<string>(() => {
		const now = new Date();
		return now.toLocaleString("default", { month: "long" });
	});
	const [selectedYear, setSelectedYear] = useState<number>(() => {
		return new Date().getFullYear();
	});
	const [selectedDuration, setSelectedDuration] =
		useState<string>("Half month");
	const [includeTransactions, setIncludeTransactions] = useState<boolean>(true);

	const years = Array.from(
		{ length: 5 },
		(_, i) => new Date().getFullYear() - 2 + i,
	);

	const handleSend = () => {
		const finalMessage = messageTemplate
			.replace(/\{month\}/g, selectedMonth)
			.replace(/\{year\}/g, selectedYear.toString())
			.replace(/\{duration\}/g, selectedDuration);
		onSend(finalMessage, selectedMonth, selectedYear, includeTransactions);
	};

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent>
				<DialogHeader>
					<DialogTitle>Edit WhatsApp Message</DialogTitle>
					<DialogDescription>
						You can edit the message template that will be sent to the user.
					</DialogDescription>
				</DialogHeader>
				<div className="space-y-4">
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
					<div className="flex items-center space-x-2">
						<Checkbox
							id="include-transactions"
							checked={includeTransactions}
							onCheckedChange={(checked) =>
								setIncludeTransactions(checked as boolean)
							}
						/>
						<Label htmlFor="include-transactions">
							Include transaction history
						</Label>
					</div>
					<Textarea
						rows={6}
						value={messageTemplate}
						onChange={(e) => setMessageTemplate(e.target.value)}
						className="w-full min-h-[120px] font-mono"
					/>
				</div>
				<DialogFooter>
					<Button onClick={handleSend} disabled={sendingNotification}>
						Send Notification
					</Button>
					<Button variant="outline" onClick={() => onOpenChange(false)}>
						Cancel
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}
