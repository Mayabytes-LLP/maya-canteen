import { useContext, useEffect, useState } from "react";
import { toast } from "sonner";

import { AppContext } from "@/components/canteen-provider";
import { Button } from "@/components/ui/button";
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
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { formatDate, formatPrice } from "@/lib/utils";
import {
  Transaction,
  transactionService,
  User,
} from "@/services/transaction-service";
import { Info, MoreHorizontal, Pencil, Trash2 } from "lucide-react";

import DateRangeFilter from "@/components/date-range-filter";

interface TransactionListProps {
  limit?: number;
  refreshTrigger?: number;
}

export default function TransactionList({
  limit = 10,
  refreshTrigger = 0,
}: TransactionListProps) {
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [isFiltered, setIsFiltered] = useState(false);

  const { admin } = useContext(AppContext);

  // CRUD state management
  const [selectedTransaction, setSelectedTransaction] =
    useState<Transaction | null>(null);
  const [editTransaction, setEditTransaction] = useState<Transaction | null>(
    null,
  );
  const [deleteTransactionId, setDeleteTransactionId] = useState<number | null>(
    null,
  );
  const [editAmount, setEditAmount] = useState("");
  const [editDescription, setEditDescription] = useState("");
  const [editType, setEditType] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [openViewDialog, setOpenViewDialog] = useState(false);
  const [openEditDialog, setOpenEditDialog] = useState(false);
  const [openDeleteDialog, setOpenDeleteDialog] = useState(false);

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      try {
        const [transactionsData, usersData] = await Promise.all([
          transactionService.getLatestTransactions(limit),
          transactionService.getAllUsers(),
        ]);
        setTransactions(transactionsData ?? []);
        setUsers(usersData ?? []);
        setIsFiltered(false);
      } catch (error) {
        toast.error("Failed to load transactions");
        console.error(error);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [limit, refreshTrigger]);

  // Handler for when date range filter loads transactions
  const handleFilteredTransactions = (filteredTransactions: Transaction[]) => {
    setTransactions(filteredTransactions);
    setIsFiltered(true);
  };

  // Reset filter and load latest transactions
  const handleResetFilter = async () => {
    setLoading(true);
    try {
      const transactionsData =
        await transactionService.getLatestTransactions(limit);
      setTransactions(transactionsData);
      setIsFiltered(false);
      toast.success("Showing latest transactions");
    } catch (error) {
      toast.error("Failed to reset transactions");
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  // Helper function to get user name by ID
  const getUserName = (userId: number): string => {
    const user = users.find((u) => u.id === userId);
    return user ? user.name : "Unknown User";
  };

  // Format transaction amount with currency symbol
  const formatAmount = (amount: number, type: string): string => {
    return `${type === "deposit" ? "+" : "-"}${formatPrice(amount)}`;
  };

  // View transaction details
  const handleViewDetails = (transaction: Transaction) => {
    setSelectedTransaction(transaction);
    setOpenViewDialog(true);
  };

  // Edit transaction
  const handleEdit = (transaction: Transaction) => {
    setEditTransaction(transaction);
    setEditAmount(transaction.amount.toString());
    setEditDescription(transaction.description);
    setEditType(transaction.transaction_type);
    setOpenEditDialog(true);
  };

  // Update transaction
  const handleUpdate = async () => {
    if (!editTransaction) return;

    setIsSubmitting(true);
    try {
      const updatedTransaction = {
        ...editTransaction,
        amount: parseFloat(editAmount),
        description: editDescription,
        transaction_type: editType,
      };

      await transactionService.updateTransaction(updatedTransaction);

      // Update local state
      setTransactions(
        transactions.map((t) =>
          t.id === editTransaction.id ? updatedTransaction : t,
        ),
      );

      toast.success("Transaction updated successfully");
      setEditTransaction(null);
      setOpenEditDialog(false);
    } catch (error) {
      console.error("Error updating transaction:", error);
      toast.error("Failed to update transaction");
    } finally {
      setIsSubmitting(false);
    }
  };

  // Confirm delete
  const confirmDelete = (transactionId: number) => {
    setDeleteTransactionId(transactionId);
    setOpenDeleteDialog(true);
  };

  // Delete transaction
  const handleDelete = async () => {
    if (!deleteTransactionId) return;

    setIsSubmitting(true);
    try {
      await transactionService.deleteTransaction(deleteTransactionId);

      // Update local state
      setTransactions(transactions.filter((t) => t.id !== deleteTransactionId));

      toast.success("Transaction deleted successfully");
      setDeleteTransactionId(null);
      setOpenDeleteDialog(false);
    } catch (error) {
      console.error("Error deleting transaction:", error);
      toast.error("Failed to delete transaction");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Card>
      <CardHeader className="flex flex-col space-y-4">
        <CardTitle>
          {isFiltered
            ? "Filtered Transactions"
            : `Latest ${limit} Transactions`}
        </CardTitle>
        <div className="flex flex-col space-y-2">
          <DateRangeFilter onTransactionsLoaded={handleFilteredTransactions} />
          {isFiltered && (
            <button
              onClick={handleResetFilter}
              className="text-sm text-blue-600 hover:underline"
            >
              Reset filter and show latest transactions
            </button>
          )}
        </div>
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className="flex justify-center py-6">
            Loading transactions...
          </div>
        ) : transactions.length === 0 ? (
          <div className="text-center py-6 text-muted-foreground">
            No transactions found
          </div>
        ) : (
          <div className="space-y-4">
            {transactions.map((transaction) => (
              <div
                key={transaction.id}
                className="flex items-center justify-between border-b pb-2"
              >
                <div className="space-y-1">
                  <p className="font-medium">
                    {getUserName(transaction.user_id)}
                  </p>
                  <p className="text-sm text-muted-foreground">
                    {transaction.description}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    {formatDate(transaction.created_at)}
                  </p>
                </div>
                <div className="flex items-center space-x-2">
                  <div
                    className={`font-medium ${
                      transaction.transaction_type === "deposit"
                        ? "text-green-600"
                        : "text-red-600"
                    }`}
                  >
                    {formatAmount(
                      transaction.amount,
                      transaction.transaction_type,
                    )}
                  </div>
                  {admin && (
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" className="h-8 w-8 p-0">
                          <span className="sr-only">Open menu</span>
                          <MoreHorizontal className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem
                          onClick={() => handleViewDetails(transaction)}
                        >
                          <Info className="mr-2 h-4 w-4" />
                          View Details
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          onClick={() => handleEdit(transaction)}
                        >
                          <Pencil className="mr-2 h-4 w-4" />
                          Edit
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          onClick={() => confirmDelete(transaction.id)}
                          className="text-red-600"
                        >
                          <Trash2 className="mr-2 h-4 w-4" />
                          Delete
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>

      {/* View Transaction Details Dialog */}
      <Dialog open={openViewDialog} onOpenChange={setOpenViewDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Transaction Details</DialogTitle>
          </DialogHeader>
          {selectedTransaction && (
            <div className="space-y-4">
              <div className="grid grid-cols-3 gap-4">
                <div>
                  <p className="text-sm font-medium text-muted-foreground">
                    User
                  </p>
                  <p>{getUserName(selectedTransaction.user_id)}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">
                    Amount
                  </p>
                  <p
                    className={
                      selectedTransaction.transaction_type === "deposit"
                        ? "text-green-600"
                        : "text-red-600"
                    }
                  >
                    {formatAmount(
                      selectedTransaction.amount,
                      selectedTransaction.transaction_type,
                    )}
                  </p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">
                    Type
                  </p>
                  <p className="capitalize">
                    {selectedTransaction.transaction_type}
                  </p>
                </div>
              </div>

              <div>
                <p className="text-sm font-medium text-muted-foreground">
                  Description
                </p>
                <p>{selectedTransaction.description}</p>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <p className="text-sm font-medium text-muted-foreground">
                    Created
                  </p>
                  <p>{formatDate(selectedTransaction.created_at)}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">
                    Last Updated
                  </p>
                  <p>{formatDate(selectedTransaction.updated_at)}</p>
                </div>
              </div>
            </div>
          )}
          <DialogFooter>
            <DialogClose asChild>
              <Button>Close</Button>
            </DialogClose>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit Transaction Dialog */}
      <Dialog open={openEditDialog} onOpenChange={setOpenEditDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Transaction</DialogTitle>
            <DialogDescription>
              Update the transaction information below.
            </DialogDescription>
          </DialogHeader>

          {editTransaction && (
            <div className="grid gap-4 py-4">
              <div className="flex items-center gap-4">
                <Label htmlFor="user" className="w-[120px]">
                  User
                </Label>
                <p>{getUserName(editTransaction.user_id)}</p>
              </div>

              <div className="flex items-center gap-4">
                <Label htmlFor="amount" className="w-[120px]">
                  Amount (PKR.)
                </Label>
                <Input
                  id="amount"
                  type="number"
                  step="1"
                  value={editAmount}
                  onChange={(e) => setEditAmount(e.target.value)}
                  className="flex-1"
                />
              </div>

              <div className="flex items-center gap-4">
                <Label htmlFor="transaction_type" className="w-[120px]">
                  Type
                </Label>
                <Select value={editType} onValueChange={setEditType}>
                  <SelectTrigger className="flex-1">
                    <SelectValue placeholder="Select type" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="purchase">Purchase</SelectItem>
                    <SelectItem value="deposit">Deposit</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="flex items-start gap-4">
                <Label htmlFor="description" className="w-[120px] pt-2">
                  Description
                </Label>
                <Textarea
                  id="description"
                  value={editDescription}
                  onChange={(e) => setEditDescription(e.target.value)}
                  className="flex-1"
                />
              </div>
            </div>
          )}

          <DialogFooter>
            <DialogClose asChild>
              <Button variant="outline">Cancel</Button>
            </DialogClose>
            <Button onClick={handleUpdate} disabled={isSubmitting}>
              {isSubmitting ? "Updating..." : "Save Changes"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Transaction Confirmation Dialog */}
      <Dialog open={openDeleteDialog} onOpenChange={setOpenDeleteDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Confirm Deletion</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete this transaction? This action
              cannot be undone.
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
              {isSubmitting ? "Deleting..." : "Delete Transaction"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Card>
  );
}
