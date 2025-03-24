import { useState } from "react";

import DepositForm from "@/components/deposit-form";
import TransactionList from "@/components/transaction-list";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import ErrorBoundary from "@/components/error-boundary";

export default function TransactionsPage() {
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  const [transactionLimit, setTransactionLimit] = useState(50);
  const [inputLimit, setInputLimit] = useState("10");

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
    <div className="container mx-auto py-6 space-y-6">
      <h1 className="text-3xl font-bold">Transactions Management</h1>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div>
          <DepositForm onTransactionAdded={handleTransactionAdded} />
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
