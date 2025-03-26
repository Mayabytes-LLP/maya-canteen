import { format } from "date-fns";
import { useEffect, useState } from "react";
import { toast } from "sonner";

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
import { formatTransaction } from "@/lib/utils";
import {
  Transaction,
  User,
  transactionService,
} from "@/services/transaction-service";
import { Info } from "lucide-react";
import { ScrollArea } from "../ui/scroll-area";

interface UserTransactionsProps {
  userId: string;
  onClose: () => void;
}

export default function UserTransactions({
  userId,
  onClose,
}: UserTransactionsProps) {
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [selectedTransaction, setSelectedTransaction] =
    useState<Transaction | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      try {
        const [transactionsData, userData] = await Promise.all([
          transactionService.getTransactionsByUserId(userId),
          transactionService.getUser(userId),
        ]);
        setTransactions(transactionsData);
        setUser(userData);
      } catch (error) {
        toast.error("Failed to load user transactions");
        console.error(error);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [userId]);

  // Calculate balance
  const calculateBalance = (): number => {
    if (!Array.isArray(transactions)) {
      return 0;
    }
    return transactions.reduce((balance, transaction) => {
      if (transaction.transaction_type === "deposit") {
        return balance + transaction.amount;
      } else {
        return balance - transaction.amount;
      }
    }, 0);
  };

  return (
    <Card className="w-full">
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle>
          {user ? `${user.name}'s Transactions` : "User Transactions"}
        </CardTitle>
        <Button variant="outline" onClick={onClose}>
          Close
        </Button>
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className="flex justify-center py-6">
            Loading transactions...
          </div>
        ) : !transactions || transactions.length === 0 ? (
          <div className="text-center py-6 text-muted-foreground">
            No transactions found for this user
          </div>
        ) : (
          <>
            <div className="mb-6 p-4 bg-muted rounded-md">
              <div className="flex justify-between items-center">
                <div>
                  <p className="text-sm font-medium text-muted-foreground">
                    User
                  </p>
                  <p className="font-medium">{user?.name}</p>
                  <p className="text-sm text-muted-foreground">
                    Employee ID: {user?.employee_id}
                  </p>
                </div>
                <div className="text-right">
                  <p className="text-sm font-medium text-muted-foreground">
                    Current Balance
                  </p>
                  <p
                    className={`text-xl font-bold ${
                      calculateBalance() >= 0
                        ? "text-green-600"
                        : "text-red-600"
                    }`}
                  >
                    PKR.{calculateBalance().toFixed(2)}
                  </p>
                </div>
              </div>
            </div>

            <ScrollArea className="h-96">
              <div className="space-y-4">
                {transactions.map((transaction) => (
                  <div
                    key={transaction.id}
                    className="flex items-center justify-between border-b pb-2"
                  >
                    <div className="space-y-1">
                      <div className="flex items-center space-x-2">
                        <p className="font-medium">
                          {transaction.transaction_type === "deposit"
                            ? "Deposit"
                            : "Purchase"}
                        </p>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-6 w-6 p-0"
                          onClick={() => setSelectedTransaction(transaction)}
                        >
                          <Info className="h-3 w-3" />
                          <span className="sr-only">Details</span>
                        </Button>
                      </div>
                      <p className="text-sm text-muted-foreground">
                        {transaction.description}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {format(new Date(transaction.created_at), "PPp")}
                      </p>
                    </div>
                    <div
                      className={`font-medium ${
                        transaction.transaction_type === "deposit"
                          ? "text-green-600"
                          : "text-red-600"
                      }`}
                    >
                      {formatTransaction(
                        transaction.amount,
                        transaction.transaction_type
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </ScrollArea>
          </>
        )}
      </CardContent>

      {/* Transaction Details Dialog */}
      <Dialog
        open={selectedTransaction !== null}
        onOpenChange={(open) => !open && setSelectedTransaction(null)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Transaction Details</DialogTitle>
            <DialogDescription>
              Details of the selected transaction
            </DialogDescription>
          </DialogHeader>
          {selectedTransaction && (
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
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
                    {formatTransaction(
                      selectedTransaction.amount,
                      selectedTransaction.transaction_type
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
                  <p>
                    {format(new Date(selectedTransaction.created_at), "PPpp")}
                  </p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">
                    Last Updated
                  </p>
                  <p>
                    {format(new Date(selectedTransaction.updated_at), "PPpp")}
                  </p>
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
    </Card>
  );
}
