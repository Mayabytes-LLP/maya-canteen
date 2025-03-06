import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Transaction,
  transactionService,
} from "@/services/transaction-service";
import { useState } from "react";
import { toast } from "sonner";

interface DateRangeFilterProps {
  onTransactionsLoaded: (transactions: Transaction[]) => void;
}

export default function DateRangeFilter({
  onTransactionsLoaded,
}: DateRangeFilterProps) {
  const [startDate, setStartDate] = useState("");
  const [endDate, setEndDate] = useState("");
  const [loading, setLoading] = useState(false);

  const handleFilter = async () => {
    if (!startDate || !endDate) {
      toast.error("Please select both start and end dates");
      return;
    }

    setLoading(true);
    try {
      const dateRange = { startDate, endDate };
      const transactions = await transactionService.getTransactionsByDateRange(
        dateRange
      );
      onTransactionsLoaded(transactions);
      toast.success("Transactions filtered successfully");
    } catch (error) {
      toast.error("Failed to filter transactions");
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex items-center space-x-4">
      <div>
        <Label htmlFor="start-date">Start Date</Label>
        <Input
          id="start-date"
          type="date"
          value={startDate}
          onChange={(e) => setStartDate(e.target.value)}
        />
      </div>
      <div>
        <Label htmlFor="end-date">End Date</Label>
        <Input
          id="end-date"
          type="date"
          value={endDate}
          onChange={(e) => setEndDate(e.target.value)}
        />
      </div>
      <Button onClick={handleFilter} disabled={loading}>
        {loading ? "Filtering..." : "Filter"}
      </Button>
    </div>
  );
}
