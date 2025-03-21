import { useContext, useEffect, useState } from "react";
import { toast } from "sonner";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { cn } from "@/lib/utils";
import {
  Departments,
  zodUserSchema as formSchema,
  transactionService,
  User,
  UserBalance,
} from "@/services/transaction-service";
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
import {
  Check,
  ChevronsUpDown,
  CreditCard,
  MoreHorizontal,
  Pencil,
  Trash2,
} from "lucide-react";
import { useForm } from "react-hook-form";
import * as z from "zod";
import { AppContext } from "../canteen-provider";
import { Button } from "../ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";
import UserTransactions from "./user-transactions";

import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { CopyButton } from "../ui/copy-button";

type FormValues = z.infer<typeof formSchema>;

interface UserBalancesProps {
  refreshTrigger?: number;
}

export default function UserBalances({
  refreshTrigger = 0,
}: UserBalancesProps) {
  const [balances, setBalances] = useState<UserBalance[]>([]);
  const [loading, setLoading] = useState(true);
  const [users, setUsers] = useState<User[]>([]);

  const [deleteUserId, setDeleteUserId] = useState<number | null>(null);
  const [selectedUserId, setSelectedUserId] = useState<string | null>(null);

  const [editUser, setEditUser] = useState<User | null>(null);

  const [openEditDialog, setOpenEditDialog] = useState(false);
  const [openDeleteDialog, setOpenDeleteDialog] = useState(false);
  const [openTransactionsDialog, setOpenTransactionsDialog] = useState(false);

  const [isSubmitting, setIsSubmitting] = useState(false);
  const [userPopover, setUserPopover] = useState(false);

  const { admin } = useContext(AppContext);

  // Initialize form
  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: "",
      employee_id: "",
      phone: "",
    },
  });

  useEffect(() => {
    const fetchBalances = async () => {
      setLoading(true);
      try {
        const balancesData = await transactionService.getUsersBalances();
        setBalances(balancesData ?? []);
      } catch (error) {
        toast.error("Failed to load users' balances");
        console.error("Error fetching user balances:", error);
      } finally {
        setLoading(false);
      }
    };

    fetchBalances();
  }, [refreshTrigger]);

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
      });

      // Update the local state
      setUsers(
        users.map((user) =>
          user.id === editUser.id
            ? {
                ...user,
                name: data.name,
                department: data.department,
                employee_id: data.employee_id,
                phone: data.phone,
              }
            : user,
        ),
      );

      setBalances(
        balances.map((balance) =>
          balance.user_id === editUser.id
            ? {
                ...balance,
                user_name: data.name,
                user_department: data.department,
                employee_id: data.employee_id,
                user_phone: data.phone,
              }
            : balance,
        ),
      );

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

      // Update the local state
      setUsers(users.filter((user) => user.id !== deleteUserId));

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
      <CardHeader>
        <CardTitle>Users' Balances</CardTitle>
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
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Employee ID</TableHead>
                  <TableHead>Phone</TableHead>
                  <TableHead>Balance (PKR)</TableHead>
                  <TableHead className="w-[120px]">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {balances.map((balance) => (
                  <TableRow key={balance.user_id}>
                    <TableCell className="font-medium">
                      {`${balance.user_name} (${balance.user_department})`}
                    </TableCell>
                    <TableCell>{balance.employee_id}</TableCell>
                    <TableCell>
                      <div className="flex space-x-2">
                        <span className="p-1 bg-zinc-700 w-28 text-center rounded">
                          {balance.user_phone}
                        </span>

                        <CopyButton value={balance.user_phone} />
                      </div>
                    </TableCell>
                    <TableCell>
                      {balance.balance && balance.balance.toFixed(2)}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center space-x-1">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() =>
                            showUserTransactions(balance.employee_id)
                          }
                          className="h-8 w-8 p-0"
                          title="View Transactions"
                        >
                          <CreditCard className="h-4 w-4" />
                          <span className="sr-only">View Transactions</span>
                        </Button>

                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" className="h-8 w-8 p-0">
                              <span className="sr-only">Open menu</span>
                              <MoreHorizontal className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem
                              onClick={() => handleEdit(balance)}
                            >
                              <Pencil className="mr-2 h-4 w-4" />
                              Edit
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              disabled={!admin}
                              onClick={() => confirmDelete(balance.user_id)}
                              className="text-red-600"
                            >
                              <Trash2 className="mr-2 h-4 w-4" />
                              Delete
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
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
    </Card>
  );
}
