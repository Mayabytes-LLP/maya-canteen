import { useContext, useEffect, useState } from "react";

import TransactionForm from "@/components/transaction-form";
import TransactionUserList from "@/components/transaction-user-list";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { AppContext } from "@/context";
import { cn } from "@/lib/utils";
import {
  transactionService,
  type UserBalance,
} from "@/services/transaction-service";
import { toast } from "sonner";

export default function CanteenPage() {
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  const [transactionLimit, setTransactionLimit] = useState(20);
  const [inputLimit, setInputLimit] = useState("10");
  const [balance, setBalance] = useState<UserBalance | null>(null);

  const { currentUser } = useContext(AppContext);

  useEffect(() => {
    if (currentUser?.id && balance === null) {
      async function fetchUserBalance() {
        const balance = currentUser
          ? await transactionService.getBalanceByUserId(currentUser.id)
          : null;
        return balance;
      }
      fetchUserBalance()
        .then((balance) => {
          if (balance !== null) {
            setBalance(balance);
            toast("Your canteen balance is " + balance.balance);
          } else {
            setBalance(null);
            toast.error("Failed to fetch user balance");
            console.error("Failed to fetch user balance");
          }
        })
        .catch((err) => {
          console.error(err);
        });
    }
  }, [currentUser, balance]);

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

  return (
    <div className="container mx-auto py-8">
      {balance && (
        <div className="mb-4 text-lg text-center bg-accent p-4 rounded-lg">
          <p>
            Your canteen balance is:{" "}
            <span className={cn(balance.balance < 0 && "text-red-500")}>
              {balance.balance}
            </span>
          </p>
        </div>
      )}

      <div className="grid grid-cols-1 gap-8">
        <div>
          <TransactionForm onTransactionAdded={handleTransactionAdded} />
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
          <TransactionUserList
            limit={transactionLimit}
            refreshTrigger={refreshTrigger}
          />
        </div>
      </div>
    </div>
  );
}
