import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import UserTransactions from "@/components/user/user-transactions";
import { AppContext } from "@/context";
import { cn } from "@/lib/utils";
import {
  DateRangeRequest,
  Transaction,
  UserBalance,
  transactionService,
} from "@/services/transaction-service";
import { format } from "date-fns";
import {
  ArrowUpIcon,
  CalendarIcon,
  CreditCardIcon,
  DownloadIcon,
  Search,
  Send,
  WalletIcon,
  X,
} from "lucide-react";
import { useCallback, useContext, useEffect, useState } from "react";
import { DateRange } from "react-day-picker";
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Legend,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { toast } from "sonner";

const COLORS = ["#0088FE", "#00C49F", "#FFBB28", "#FF8042", "#8884d8"];

const DashboardPage = () => {
  const [dateRange, setDateRange] = useState<DateRange>({
    from: new Date(new Date().setDate(1)), // First day of current month
    to: new Date(),
  });
  const [isLoading, setIsLoading] = useState(false);
  const [userBalances, setUserBalances] = useState<UserBalance[]>([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedUserId, setSelectedUserId] = useState<string | null>(null);
  const [openTransactionsDialog, setOpenTransactionsDialog] = useState(false);
  const [sendingNotification, setSendingNotification] = useState(false);
  const { admin, whatsappStatus } = useContext(AppContext);

  // Filtered balances based on search query
  const filteredBalances = userBalances.filter(
    (user) =>
      user.balance !== 0 &&
      (searchQuery === "" ||
        user.user_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        user.user_department
          .toLowerCase()
          .includes(searchQuery.toLowerCase()) ||
        user.employee_id.toLowerCase().includes(searchQuery.toLowerCase())),
  );

  // Calculated summary values
  const [summaryData, setSummaryData] = useState({
    totalTransactions: 0,
    totalPurchase: 0,
    totalReceived: 0, // Total amount received in payments
    totalOwedByUsers: 0, // Total amount users owe to the canteen
    totalCreditToUsers: 0, // Total amount canteen owes to users
    recentTransactions: [] as Transaction[],
    transactionsByType: [] as { name: string; value: number }[],
    transactionTrend: [] as {
      date: string;
      deposit: number;
      purchase: number;
    }[],
  });

  // Use useCallback to memoize the fetchDashboardData function
  const fetchDashboardData = useCallback(async () => {
    if (!dateRange.from || !dateRange.to) return;

    setIsLoading(true);
    try {
      // Create a request with the date range - format dates to YYYY-MM-DD
      const dateRangeRequest: DateRangeRequest = {
        startDate: format(dateRange.from, "yyyy-MM-dd"),
        endDate: format(dateRange.to, "yyyy-MM-dd"),
      };

      // Fetch transactions for the date range
      const transactionsData =
        await transactionService.getTransactionsByDateRange(dateRangeRequest);
      console.log("Transactions data:", transactionsData);

      // Fetch user balances
      const balancesData = await transactionService.getUsersBalances();
      setUserBalances(balancesData);
      console.log("User balances:", balancesData);

      // Get latest transactions
      const latestTransactions =
        await transactionService.getLatestTransactions(5);
      console.log("Latest transactions:", latestTransactions);

      // Process the data for summaries
      processData(
        transactionsData ?? [],
        balancesData ?? [],
        latestTransactions ?? [],
      );
    } catch (error) {
      console.error("Error fetching dashboard data:", error);
      toast.error("Failed to load dashboard data");
    } finally {
      setIsLoading(false);
    }
  }, [dateRange]);

  // Fetch data when date range changes
  useEffect(() => {
    if (dateRange.from && dateRange.to) {
      fetchDashboardData();
    }
  }, [fetchDashboardData, dateRange.from, dateRange.to]);

  const processData = (
    transactionsData: Transaction[],
    balancesData: UserBalance[],
    latestTransactions: Transaction[],
  ) => {
    // Calculate total received (deposits) and owed (negative balances)
    const totalReceived = transactionsData
      .filter((transaction) => transaction.transaction_type === "deposit")
      .reduce((sum, transaction) => sum + transaction.amount, 0);

    const totalPurchase = transactionsData
      .filter((transaction) => transaction.transaction_type === "purchase")
      .reduce((sum, transaction) => sum + transaction.amount, 0);

    // Total amount users owe to the canteen (negative balances)
    const totalOwedByUsers = balancesData
      .filter((user) => user.balance < 0)
      .reduce((sum, user) => sum + Math.abs(user.balance), 0);

    // Total amount canteen owes to users (positive balances)
    const totalCreditToUsers = balancesData
      .filter((user) => user.balance > 0)
      .reduce((sum, user) => sum + user.balance, 0);

    // Group transactions by type
    const typeGroups = transactionsData.reduce(
      (groups, transaction) => {
        const type = transaction.transaction_type;
        if (!groups[type]) {
          groups[type] = 0;
        }
        groups[type] += Math.abs(transaction.amount);
        return groups;
      },
      {} as Record<string, number>,
    );

    const transactionsByType = Object.entries(typeGroups).map(
      ([name, value]) => ({
        name,
        value,
      }),
    );

    // Create data for transaction trend over time
    const dateMap = new Map<string, { deposit: number; purchase: number }>();

    transactionsData.forEach((transaction) => {
      const date = transaction.created_at.split("T")[0];
      if (!dateMap.has(date)) {
        dateMap.set(date, { deposit: 0, purchase: 0 });
      }

      const entry = dateMap.get(date)!;
      if (transaction.transaction_type === "deposit") {
        entry.deposit += transaction.amount;
      } else {
        entry.purchase += Math.abs(transaction.amount);
      }
    });

    const transactionTrend: {
      date: string;
      deposit: number;
      purchase: number;
    }[] = [];

    // Sort dates and create the trend data
    [...dateMap.entries()]
      .sort((a, b) => a[0].localeCompare(b[0]))
      .forEach(([date, values]) => {
        transactionTrend.push({
          date: format(new Date(date), "MMM dd"),
          deposit: values.deposit,
          purchase: values.purchase,
        });
      });

    setSummaryData({
      totalPurchase,
      totalTransactions: transactionsData.length,
      totalReceived,
      totalOwedByUsers,
      totalCreditToUsers,
      recentTransactions: latestTransactions,
      transactionsByType,
      transactionTrend,
    });
  };

  const handleExportData = () => {
    // TODO: Implement export functionality
    toast.info("Export functionality will be implemented soon");
  };

  // Function to send balance notification to a single user
  const sendBalanceNotification = async (employeeId: string) => {
    setSendingNotification(true);
    try {
      const response =
        await transactionService.sendBalanceNotification(employeeId);
      if (response.success) {
        toast.success(
          `Balance notification sent to user with ID ${employeeId}`,
        );
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

  // Function to send balance notifications to all users
  const sendAllBalanceNotifications = async () => {
    setSendingNotification(true);
    try {
      const response = await transactionService.sendAllBalanceNotifications();
      if (response.success) {
        toast.success("Balance notifications sent to all users");
      } else {
        toast.error("Failed to send balance notifications to all users");
      }
    } catch (error) {
      console.error("Error sending all balance notifications:", error);
      toast.error("Failed to send balance notifications to all users");
    } finally {
      setSendingNotification(false);
    }
  };

  const showUserTransactions = (userId: string) => {
    setSelectedUserId(userId);
    setOpenTransactionsDialog(true);
  };

  return (
    <div className="container mx-auto p-4 space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold">Canteen Dashboard</h1>
        <div className="flex space-x-4 items-center">
          <Popover>
            <PopoverTrigger asChild>
              <Button
                variant={"outline"}
                className={cn(
                  "w-[240px] pl-3 text-left font-normal",
                  !dateRange && "text-muted-foreground",
                )}
              >
                {dateRange.from ? (
                  dateRange.to ? (
                    <>
                      {format(dateRange.from, "MMM dd, yyyy")} -{" "}
                      {format(dateRange.to, "MMM dd, yyyy")}
                    </>
                  ) : (
                    format(dateRange.from, "MMM dd, yyyy")
                  )
                ) : (
                  <span>Pick a date range</span>
                )}
                <CalendarIcon className="ml-auto h-4 w-4 opacity-50" />
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-auto p-0" align="end">
              <Calendar
                mode="range"
                defaultMonth={dateRange.from}
                selected={dateRange}
                onSelect={(range) => {
                  if (range) {
                    setDateRange(range);
                  }
                }}
                numberOfMonths={2}
              />
            </PopoverContent>
          </Popover>

          <Button onClick={handleExportData} variant="outline">
            <DownloadIcon className="mr-2 h-4 w-4" />
            Export Data
          </Button>

          <Button onClick={fetchDashboardData} disabled={isLoading}>
            {isLoading ? "Loading..." : "Refresh"}
          </Button>
        </div>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Payments Received
            </CardTitle>
            <ArrowUpIcon className="h-4 w-4 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              ₨ {summaryData.totalReceived.toFixed(2)}
            </div>
            <p className="text-xs text-muted-foreground">
              Total cash received from employees
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Purchases</CardTitle>
            <CreditCardIcon className="h-4 w-4 text-blue-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              ₨ {summaryData.totalPurchase.toFixed(2)}
            </div>
            <p className="text-xs text-muted-foreground">
              Total amount spent on purchases
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Employees Owe</CardTitle>
            <CreditCardIcon className="h-4 w-4 text-red-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              ₨ {summaryData.totalOwedByUsers.toFixed(2)}
            </div>
            <p className="text-xs text-muted-foreground">
              Amount employees need to pay
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Credit to Employees
            </CardTitle>
            <WalletIcon className="h-4 w-4 text-blue-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              ₨ {summaryData.totalCreditToUsers.toFixed(2)}
            </div>
            <p className="text-xs text-muted-foreground">
              Amount canteen owes to employees
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Transactions</CardTitle>
            <ArrowUpIcon className="h-4 w-4 text-purple-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {summaryData.totalTransactions}
            </div>
            <p className="text-xs text-muted-foreground">
              For the selected period
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card className="col-span-1">
          <CardHeader>
            <CardTitle>Transaction Trend</CardTitle>
          </CardHeader>
          <CardContent className="h-80">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart
                data={summaryData.transactionTrend}
                margin={{
                  top: 20,
                  right: 30,
                  left: 20,
                  bottom: 5,
                }}
              >
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="date" />
                <YAxis />
                <Tooltip />
                <Legend />
                <Bar
                  dataKey="deposit"
                  fill="#00C49F"
                  name="Payments Received"
                />
                <Bar dataKey="purchase" fill="#FF8042" name="Items Purchased" />
              </BarChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        <Card className="col-span-1">
          <CardHeader>
            <CardTitle>Transaction Types</CardTitle>
          </CardHeader>
          <CardContent className="h-80">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={summaryData.transactionsByType}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={({ name, percent }) =>
                    `${name === "deposit" ? "Payments" : "Purchases"}: ${(
                      percent * 100
                    ).toFixed(0)}%`
                  }
                  outerRadius={80}
                  fill="#8884d8"
                  dataKey="value"
                >
                  {summaryData.transactionsByType.map((_, index) => (
                    <Cell
                      key={`cell-${index}`}
                      fill={COLORS[index % COLORS.length]}
                    />
                  ))}
                </Pie>
                <Tooltip
                  formatter={(value, name) => [
                    value,
                    name === "deposit" ? "Payments" : "Purchases",
                  ]}
                />
              </PieChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      </div>

      {/* Recent Transactions */}
      <Card>
        <CardHeader>
          <CardTitle>Recent Transactions</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Type</TableHead>
                <TableHead>Description</TableHead>
                <TableHead>Amount</TableHead>
                <TableHead>Date</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {summaryData.recentTransactions.map((transaction) => (
                <TableRow key={transaction.id}>
                  <TableCell>
                    <Badge
                      variant={
                        transaction.transaction_type === "deposit"
                          ? "default"
                          : "destructive"
                      }
                    >
                      {transaction.transaction_type === "deposit"
                        ? "Payment"
                        : "Purchase"}
                    </Badge>
                  </TableCell>
                  <TableCell>{transaction.description}</TableCell>
                  <TableCell
                    className={
                      transaction.amount >= 0
                        ? "text-green-600"
                        : "text-red-600"
                    }
                  >
                    ₨ {Math.abs(transaction.amount).toFixed(2)}
                  </TableCell>
                  <TableCell>
                    {format(new Date(transaction.created_at), "MMM dd, yyyy")}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* Users with Outstanding Balances */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle>Employee Account Balances</CardTitle>
          {admin && (
            <Button
              onClick={sendAllBalanceNotifications}
              disabled={sendingNotification || !whatsappStatus?.connected}
              className="flex items-center gap-2"
            >
              {sendingNotification ? (
                "Sending..."
              ) : (
                <>
                  <Send className="h-4 w-4" /> Send All Balances
                </>
              )}
            </Button>
          )}
        </CardHeader>
        <CardContent>
          {/* Search input */}
          <div className="flex items-end py-4">
            <div className="relative w-full max-w-sm">
              <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search by Name, Department or ID..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-8 pr-8"
              />
              {searchQuery && (
                <Button
                  variant="ghost"
                  onClick={() => setSearchQuery("")}
                  className="absolute right-0 top-0 h-full px-3 py-2"
                >
                  <X className="h-4 w-4" />
                  <span className="sr-only">Clear search</span>
                </Button>
              )}
            </div>
          </div>

          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Department</TableHead>
                <TableHead>Balance</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredBalances.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} className="h-24 text-center">
                    No matching balances found.
                  </TableCell>
                </TableRow>
              ) : (
                filteredBalances
                  .sort((a, b) => a.balance - b.balance)
                  .map((user) => (
                    <TableRow key={user.user_id}>
                      <TableCell>{user.user_name}</TableCell>
                      <TableCell>{user.user_department}</TableCell>
                      <TableCell
                        className={
                          user.balance >= 0 ? "text-green-600" : "text-red-600"
                        }
                      >
                        ₨ {Math.abs(user.balance).toFixed(2)}
                        <span className="ml-1 text-xs text-gray-500">
                          {user.balance < 0
                            ? "(owes canteen)"
                            : "(canteen owes)"}
                        </span>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center space-x-2">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() =>
                              showUserTransactions(user.employee_id)
                            }
                            className="h-8 w-8 p-0"
                            title="View Transactions"
                          >
                            <CreditCardIcon className="h-4 w-4" />
                            <span className="sr-only">View Transactions</span>
                          </Button>
                          {admin && (
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() =>
                                sendBalanceNotification(user.employee_id)
                              }
                              disabled={
                                sendingNotification ||
                                !whatsappStatus?.connected ||
                                !user.user_phone
                              }
                              className="h-8 w-8 p-0"
                              title="Send Balance Notification"
                            >
                              <Send className="h-4 w-4" />
                              <span className="sr-only">
                                Send Balance Notification
                              </span>
                            </Button>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

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
    </div>
  );
};

export default DashboardPage;
