import {
	Command,
	CommandEmpty,
	CommandGroup,
	CommandInput,
	CommandItem,
	CommandList,
} from "@components/ui/command";
import {
	Popover,
	PopoverContent,
	PopoverTrigger,
} from "@components/ui/popover";
import { zodResolver } from "@hookform/resolvers/zod";
import { Check, ChevronsUpDown } from "lucide-react";
import { useCallback, useContext, useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import type * as z from "zod";
import SendAllBalance from "@/components/send-all-balance";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import {
	Dialog,
	DialogClose,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import {
	Form,
	FormControl,
	FormDescription,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { WhatsAppNotificationDialog } from "@/components/whatsapp-notification-dialog";
import { AppContext } from "@/context";
import { cn } from "@/lib/utils";
import {
	Departments,
	zodUserSchema as formSchema,
	transactionService,
	type User,
	type UserBalance,
} from "@/services/transaction-service";
import { Button } from "../ui/button";
import { BalanceTable } from "./user-data-table";
import UserTransactions from "./user-transactions";

type FormValues = z.infer<typeof formSchema>;

interface UserBalancesProps {
	refreshTrigger?: number;
}

export default function UserBalances({
	refreshTrigger = 0,
}: UserBalancesProps) {
	const [balances, setBalances] = useState<UserBalance[]>([]);
	const [loading, setLoading] = useState(true);

	const [deleteUserId, setDeleteUserId] = useState<number | null>(null);
	const [selectedUserId, setSelectedUserId] = useState<string | null>(null);

	const [editUser, setEditUser] = useState<User | null>(null);

	const [openEditDialog, setOpenEditDialog] = useState(false);
	const [openDeleteDialog, setOpenDeleteDialog] = useState(false);
	const [openTransactionsDialog, setOpenTransactionsDialog] = useState(false);

	const [isSubmitting, setIsSubmitting] = useState(false);
	const [userPopover, setUserPopover] = useState(false);
	const [sendingNotification, setSendingNotification] = useState(false);

	const { admin, whatsappStatus } = useContext(AppContext);

	const [notificationDialogOpen, setNotificationDialogOpen] = useState(false);
	const [selectedEmployeeId, setSelectedEmployeeId] = useState<string | null>(
		null,
	);

	// Initialize form
	const form = useForm<FormValues>({
		resolver: zodResolver(formSchema),
		defaultValues: {
			name: "",
			employee_id: "",
			phone: "",
			active: true,
		},
	});

	const fetchBalances = useCallback(async () => {
		setLoading(true);
		try {
			const balancesData = await transactionService.getUsersBalances();
			setBalances(balancesData ?? []);
		} catch (error) {
			toast.error("Failed to load user's balance");
			console.error("Error fetching user balances:", error);
		} finally {
			setLoading(false);
		}
	}, []);

	// biome-ignore lint/correctness/useExhaustiveDependencies: refresh trigger
	useEffect(() => {
		fetchBalances();
		// refreshTrigger is intentionally included to trigger refresh from parent component
	}, [refreshTrigger, fetchBalances]);

	// Function to send balance notification to a single user
	const sendUserBalanceNotification = async (
		employeeId: string,
		messageTemplate?: string,
		month?: string,
		year?: number,
		includeTransactions?: boolean,
	) => {
		setSendingNotification(true);
		try {
			const response = await transactionService.sendBalanceNotification(
				employeeId,
				messageTemplate,
				month,
				year,
				includeTransactions,
			);
			if (response.success) {
				toast.success(
					`Balance notification sent to user with ID ${employeeId}`,
				);
				setNotificationDialogOpen(false);
			} else {
				toast.error("Failed to send balance notification");
			}
		} catch (error) {
			console.error("Error sending balance notification:", error);
			toast.error("Failed to send balance notification");
		} finally {
			setSendingNotification(false);
		}
	};

	const handleEdit = async (user: UserBalance) => {
		const currentUser = await transactionService.getUser(user.employee_id);
		if (!currentUser) {
			toast.error("User not found");
			return;
		}

		setEditUser(currentUser);

		form.setValue("name", currentUser.name ?? "");
		form.setValue("employee_id", currentUser.employee_id ?? "");
		form.setValue("phone", currentUser.phone ?? "");
		form.setValue(
			"department",
			currentUser.department as FormValues["department"],
		);
		form.setValue("active", currentUser.active);
		setOpenEditDialog(true);
	};

	const handleUpdate = async (data: FormValues) => {
		if (!editUser) return;

		setIsSubmitting(true);
		try {
			await transactionService.updateUser({
				id: editUser.id,
				name: data.name,
				employee_id: data.employee_id,
				department: data.department,
				phone: data.phone,
				active: data.active,
			});

			// Refetch balances to ensure we have the latest data
			await fetchBalances();

			toast.success("User updated successfully");
			setEditUser(null);
			setOpenEditDialog(false);
		} catch (error) {
			console.error("Error updating user:", error);
			toast.error("Failed to update user");
		} finally {
			setIsSubmitting(false);
		}
	};

	const confirmDelete = (userId: number) => {
		setDeleteUserId(userId);
		setOpenDeleteDialog(true);
	};

	const handleDelete = async () => {
		if (!deleteUserId) return;

		setIsSubmitting(true);
		try {
			await transactionService.deleteUser(deleteUserId);

			// Refetch balances to ensure we have the latest data
			await fetchBalances();

			toast.success("User deleted successfully");
			setDeleteUserId(null);
			setOpenDeleteDialog(false);
		} catch (error) {
			console.error("Error deleting user:", error);
			toast.error("Failed to delete user");
		} finally {
			setIsSubmitting(false);
		}
	};

	const showUserTransactions = (userId: string) => {
		setSelectedUserId(userId);
		setOpenTransactionsDialog(true);
	};

	return (
		<Card className="w-full">
			<CardHeader className="flex flex-row items-center justify-between">
				<CardTitle>User's Balance</CardTitle>
				{admin && <SendAllBalance />}
			</CardHeader>
			<CardContent>
				{loading ? (
					<div className="flex justify-center items-center h-40">
						<p>Loading balances...</p>
					</div>
				) : balances.length === 0 ? (
					<div className="text-center py-4">
						<p>No balances found.</p>
					</div>
				) : (
					<div className="overflow-x-auto">
						<BalanceTable
							data={balances}
							admin={admin}
							whatsappStatus={whatsappStatus}
							onViewTransactions={showUserTransactions}
							onEdit={handleEdit}
							onDelete={confirmDelete}
							sendingNotification={sendingNotification}
							onSendBalanceNotification={(employeeId: string) => {
								setSelectedEmployeeId(employeeId);
								setNotificationDialogOpen(true);
							}}
						/>
					</div>
				)}
			</CardContent>
			{/* Edit User Dialog */}
			<Dialog open={openEditDialog} onOpenChange={setOpenEditDialog}>
				<DialogContent>
					<Form {...form}>
						<form
							onSubmit={form.handleSubmit(handleUpdate)}
							className="space-y-4"
						>
							<DialogHeader>
								<DialogTitle>Edit User</DialogTitle>
								<DialogDescription>
									Update the user information below.
								</DialogDescription>
							</DialogHeader>

							<FormField
								control={form.control}
								name="name"
								render={({ field }) => (
									<FormItem>
										<FormLabel>Name</FormLabel>
										<FormControl>
											<Input placeholder="Enter user name" {...field} />
										</FormControl>
										<FormMessage />
									</FormItem>
								)}
							/>

							<FormField
								control={form.control}
								name="employee_id"
								render={({ field }) => (
									<FormItem>
										<FormLabel>Employee ID</FormLabel>
										<FormControl>
											<Input placeholder="Enter employee ID" {...field} />
										</FormControl>
										<FormMessage />
									</FormItem>
								)}
							/>
							<FormField
								control={form.control}
								name="department"
								render={({ field }) => (
									<FormItem className="flex flex-col items-stretch">
										<FormLabel>Department</FormLabel>
										<Popover open={userPopover} onOpenChange={setUserPopover}>
											<PopoverTrigger asChild>
												<FormControl>
													{/** biome-ignore lint/a11y/useSemanticElements: shadcn */}
													<Button
														variant="outline"
														role="combobox"
														className={cn(
															"w-full justify-between",
															!field.value && "text-muted-foreground",
														)}
													>
														{field.value || "Select Department"}
														<ChevronsUpDown className="opacity-50" />
													</Button>
												</FormControl>
											</PopoverTrigger>
											<PopoverContent className="w-full p-0">
												<Command>
													<CommandInput
														placeholder="Search Employee..."
														className="h-9"
													/>
													<CommandList>
														<CommandEmpty>No Employee found.</CommandEmpty>
														<CommandGroup>
															{Departments.map((department) => (
																<CommandItem
																	key={department}
																	value={department}
																	onSelect={() => {
																		form.setValue("department", department);
																		setUserPopover(false);
																	}}
																>
																	{department}
																	<Check
																		className={cn(
																			"ml-auto",
																			department === field.value
																				? "opacity-100"
																				: "opacity-0",
																		)}
																	/>
																</CommandItem>
															))}
														</CommandGroup>
													</CommandList>
												</Command>
											</PopoverContent>
										</Popover>
										<FormDescription>Select The Department</FormDescription>
										<FormMessage />
									</FormItem>
								)}
							/>
							<FormField
								control={form.control}
								name="phone"
								render={({ field }) => (
									<FormItem>
										<FormLabel>Phone Number format ( +923452324442 )</FormLabel>
										<FormControl>
											<Input placeholder="+92 345 2324442" {...field} />
										</FormControl>
										<FormMessage />
									</FormItem>
								)}
							/>

							<FormField
								control={form.control}
								name="active"
								render={({ field }) => (
									<FormItem className="flex flex-row items-start space-x-3 space-y-0 rounded-md p-4 border">
										<FormControl>
											<Checkbox
												checked={field.value}
												onCheckedChange={field.onChange}
											/>
										</FormControl>
										<div className="space-y-1 leading-none">
											<FormLabel>Active</FormLabel>
											<FormDescription>
												Mark user as active to allow system access
											</FormDescription>
										</div>
									</FormItem>
								)}
							/>

							<DialogFooter>
								<DialogClose asChild>
									<Button variant="outline">Cancel</Button>
								</DialogClose>

								<Button type="submit" disabled={isSubmitting}>
									{isSubmitting ? "Updating..." : "Save Changes"}
								</Button>
							</DialogFooter>
						</form>
					</Form>
				</DialogContent>
			</Dialog>

			{/* Delete User Confirmation Dialog */}
			<Dialog open={openDeleteDialog} onOpenChange={setOpenDeleteDialog}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Confirm Deletion</DialogTitle>
						<DialogDescription>
							Are you sure you want to delete this user? This action cannot be
							undone.
						</DialogDescription>
					</DialogHeader>
					<DialogFooter>
						<DialogClose asChild>
							<Button variant="outline">Cancel</Button>
						</DialogClose>
						<Button
							variant="destructive"
							onClick={handleDelete}
							disabled={isSubmitting}
						>
							{isSubmitting ? "Deleting..." : "Delete User"}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>

			{/* User Transactions Dialog */}
			<Dialog
				open={openTransactionsDialog}
				onOpenChange={setOpenTransactionsDialog}
			>
				<DialogContent className="max-w-3xl">
					<DialogHeader>
						<DialogTitle>User Transactions</DialogTitle>
						<DialogDescription>
							Transactions for the selected user
						</DialogDescription>
					</DialogHeader>
					{selectedUserId && (
						<UserTransactions
							userId={selectedUserId}
							onClose={() => setOpenTransactionsDialog(false)}
						/>
					)}
				</DialogContent>
			</Dialog>

			{/* WhatsAppNotificationDialog */}
			<WhatsAppNotificationDialog
				open={notificationDialogOpen}
				onOpenChange={setNotificationDialogOpen}
				onSend={(messageTemplate, month, year, includeTransactions) => {
					if (selectedEmployeeId) {
						sendUserBalanceNotification(
							selectedEmployeeId,
							messageTemplate,
							month,
							year,
							includeTransactions,
						);
					}
				}}
				sendingNotification={sendingNotification}
			/>
		</Card>
	);
}
