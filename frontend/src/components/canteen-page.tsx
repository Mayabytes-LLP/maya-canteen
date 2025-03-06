import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import TransactionForm from "./transaction-form";
import TransactionList from "./transaction-list";

export default function CanteenPage() {
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  const [transactionLimit, setTransactionLimit] = useState(10);
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
    <div className="container mx-auto py-8">
      <h1 className="text-3xl font-bold mb-8 text-center">
        Maya Canteen System
      </h1>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
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
          <TransactionList
            limit={transactionLimit}
            refreshTrigger={refreshTrigger}
          />
        </div>
      </div>
    </div>
  );
}
