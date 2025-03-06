import { useEffect, useState } from "react";
import { toast } from "sonner";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  UserBalance,
  transactionService,
} from "@/services/transaction-service";

interface UserBalancesProps {
  refreshTrigger?: number;
}

export default function UserBalances({
  refreshTrigger = 0,
}: UserBalancesProps) {
  const [balances, setBalances] = useState<UserBalance[]>([]);
  const [loading, setLoading] = useState(true);

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
                  <TableHead>Balance (PKR)</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {balances.map((balance) => (
                  <TableRow key={balance.user_id}>
                    <TableCell className="font-medium">
                      {balance.user_name}
                    </TableCell>
                    <TableCell>{balance.employee_id}</TableCell>
                    <TableCell>
                      {balance.balance && balance.balance.toFixed(2)}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
