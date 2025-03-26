import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { cn, formatDate, formatTransaction } from "@/lib/utils";
import {
  ProductSalesSummary,
  TransactionProductDetail,
  transactionService,
} from "@/services/transaction-service";
import { format } from "date-fns";
import { CalendarIcon, DownloadIcon } from "lucide-react";
import { useState } from "react";
import { DateRange } from "react-day-picker";
import { toast } from "sonner";

export default function ProductSalesPage() {
  const [date, setDate] = useState<DateRange>({
    from: new Date(new Date().setDate(new Date().getDate() - 30)),
    to: new Date(),
  });
  const [loading, setLoading] = useState(false);
  const [productSummary, setProductSummary] = useState<ProductSalesSummary[]>(
    []
  );
  const [productDetails, setProductDetails] = useState<
    TransactionProductDetail[]
  >([]);

  const handleDateSelect = (range: DateRange | undefined) => {
    if (range) {
      setDate(range);
    }
  };

  const handleGenerateReport = async () => {
    if (!date?.from || !date?.to) {
      toast.error("Please select a date range");
      return;
    }

    setLoading(true);
    try {
      // Format dates to YYYY-MM-DD format as expected by backend
      const dateRange = {
        startDate: format(date.from, "yyyy-MM-dd"),
        endDate: format(date.to, "yyyy-MM-dd"),
      };

      // Get product sales summary
      const summary = await transactionService.getProductSalesSummary(
        dateRange
      );
      setProductSummary(summary);

      // Get product transaction details
      const details = await transactionService.getTransactionProductDetails(
        dateRange
      );
      setProductDetails(details);
      toast.success("Report generated successfully");
    } catch (error) {
      console.error("Error generating report:", error);
      toast.error("Failed to generate report");
    } finally {
      setLoading(false);
    }
  };

  const exportToCSV = <T extends object>(data: T[], filename: string) => {
    // Convert data to CSV format
    if (!data || data.length === 0) {
      toast.error("No data to export");
      return;
    }

    // Get headers from first object
    const headers = Object.keys(data[0]);

    // Create CSV content
    let csvContent = headers.join(",") + "\n";

    // Add rows
    data.forEach((item) => {
      const row = headers
        .map((header) => {
          // Format value properly, handle special cases
          const value = (item as Record<string, unknown>)[header];
          if (value === null || value === undefined) return "";
          if (typeof value === "string")
            return `"${value.replace(/"/g, '""')}"`;
          if (value instanceof Date) return format(value, "yyyy-MM-dd");
          return value;
        })
        .join(",");
      csvContent += row + "\n";
    });

    // Create and download the file
    const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8;" });
    const link = document.createElement("a");
    const url = URL.createObjectURL(blob);
    link.setAttribute("href", url);
    link.setAttribute("download", filename);
    link.style.visibility = "hidden";
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  return (
    <div className="container mx-auto py-8">
      <Card>
        <CardHeader>
          <CardTitle>Product Sales Report</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col md:flex-row items-start md:items-center gap-4 mb-6">
            <div className="flex-1">
              <div className="grid gap-2">
                <div className="flex items-center gap-2">
                  <Popover>
                    <PopoverTrigger asChild>
                      <Button
                        id="date"
                        variant={"outline"}
                        className={cn(
                          "w-full justify-start text-left font-normal",
                          !date && "text-muted-foreground"
                        )}
                      >
                        <CalendarIcon className="mr-2 h-4 w-4" />
                        {date?.from ? (
                          date.to ? (
                            <>
                              {format(date.from, "LLL dd, y")} -{" "}
                              {format(date.to, "LLL dd, y")}
                            </>
                          ) : (
                            format(date.from, "LLL dd, y")
                          )
                        ) : (
                          <span>Pick a date</span>
                        )}
                      </Button>
                    </PopoverTrigger>
                    <PopoverContent className="w-auto p-0" align="start">
                      <Calendar
                        initialFocus
                        mode="range"
                        defaultMonth={date?.from}
                        selected={date}
                        onSelect={handleDateSelect}
                        numberOfMonths={2}
                      />
                    </PopoverContent>
                  </Popover>
                </div>
              </div>
            </div>
            <Button onClick={handleGenerateReport} disabled={loading}>
              {loading ? "Generating..." : "Generate Report"}
            </Button>
          </div>

          <Tabs defaultValue="summary">
            <TabsList className="mb-4">
              <TabsTrigger value="summary">Sales Summary</TabsTrigger>
              <TabsTrigger value="details">Transaction Details</TabsTrigger>
            </TabsList>

            <TabsContent value="summary">
              <div className="flex justify-between mb-4">
                <h3 className="text-lg font-semibold">Product Sales Summary</h3>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() =>
                    exportToCSV(
                      productSummary,
                      `product-sales-summary-${format(
                        new Date(),
                        "yyyy-MM-dd"
                      )}.csv`
                    )
                  }
                  disabled={!productSummary.length}
                >
                  <DownloadIcon className="mr-2 h-4 w-4" /> Export CSV
                </Button>
              </div>

              {productSummary.length > 0 ? (
                <div className="rounded-md border">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Product</TableHead>
                        <TableHead>Type</TableHead>
                        <TableHead>Total Quantity</TableHead>
                        <TableHead>Full Units</TableHead>
                        <TableHead>Single Units</TableHead>
                        <TableHead className="text-right">
                          Total Sales (PKR)
                        </TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {productSummary.map((item) => (
                        <TableRow key={item.product_id}>
                          <TableCell className="font-medium">
                            {item.product_name}
                          </TableCell>
                          <TableCell>
                            <Badge variant="outline">{item.product_type}</Badge>
                          </TableCell>
                          <TableCell>{item.total_quantity}</TableCell>
                          <TableCell>{item.full_unit_sold}</TableCell>
                          <TableCell>{item.single_unit_sold}</TableCell>
                          <TableCell className="text-right">
                            {formatTransaction(item.total_sales)}
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              ) : (
                <div className="text-center p-8 text-muted-foreground">
                  {loading
                    ? "Loading data..."
                    : "No data available. Generate a report first."}
                </div>
              )}
            </TabsContent>

            <TabsContent value="details">
              <div className="flex justify-between mb-4">
                <h3 className="text-lg font-semibold">
                  Transaction Product Details
                </h3>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() =>
                    exportToCSV(
                      productDetails,
                      `product-sales-details-${format(
                        new Date(),
                        "yyyy-MM-dd"
                      )}.csv`
                    )
                  }
                  disabled={!productDetails.length}
                >
                  <DownloadIcon className="mr-2 h-4 w-4" /> Export CSV
                </Button>
              </div>

              {productDetails.length > 0 ? (
                <div className="rounded-md border">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Transaction ID</TableHead>
                        <TableHead>Product</TableHead>
                        <TableHead>Type</TableHead>
                        <TableHead>Quantity</TableHead>
                        <TableHead>Unit Price</TableHead>
                        <TableHead>Unit Type</TableHead>
                        <TableHead className="text-right">
                          Total Price
                        </TableHead>
                        <TableHead>Date</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {productDetails.map((item) => (
                        <TableRow key={item.id}>
                          <TableCell>{item.transaction_id}</TableCell>
                          <TableCell className="font-medium">
                            {item.product_name}
                          </TableCell>
                          <TableCell>
                            <Badge variant="outline">{item.product_type}</Badge>
                          </TableCell>
                          <TableCell>{item.quantity}</TableCell>
                          <TableCell>
                            {formatTransaction(item.unit_price)}
                          </TableCell>
                          <TableCell>
                            {item.is_single_unit ? "Single" : "Full"}
                          </TableCell>
                          <TableCell className="text-right">
                            {formatTransaction(item.total_price)}
                          </TableCell>
                          <TableCell>{formatDate(item.created_at)}</TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              ) : (
                <div className="text-center p-8 text-muted-foreground">
                  {loading
                    ? "Loading data..."
                    : "No data available. Generate a report first."}
                </div>
              )}
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  );
}
