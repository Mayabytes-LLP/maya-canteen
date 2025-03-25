"use client";

import { useState, useEffect } from "react";
import {
  type ColumnDef,
  flexRender,
  getCoreRowModel,
  useReactTable,
  getPaginationRowModel,
  type SortingState,
  getSortedRowModel,
  getFilteredRowModel,
  type ColumnFiltersState,
  type VisibilityState,
} from "@tanstack/react-table";
import {
  CreditCard,
  MoreHorizontal,
  Pencil,
  Search,
  Send,
  SendToBack,
  Trash2,
  X,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Input } from "@/components/ui/input";

import type { UserBalance as Balance } from "@/services/transaction-service";

import { CopyButton } from "@/components/ui/copy-button";

interface BalanceTableProps {
  data: Balance[];
  admin?: boolean;
  whatsappStatus: { connected: boolean };
  onViewTransactions: (employeeId: string) => void;
  onEdit: (balance: Balance) => void;
  onDelete: (userId: number) => void;
  onSendBalanceNotification: (employeeId: string) => void;
  sendingNotification: boolean;
}

export function BalanceTable({
  data,
  admin = false,
  whatsappStatus,
  onViewTransactions,
  onEdit,
  onDelete,
  onSendBalanceNotification,
  sendingNotification,
}: BalanceTableProps) {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({});

  // Update search filter when searchQuery changes
  useEffect(() => {
    if (searchQuery) {
      setColumnFilters([
        {
          id: "user_name",
          value: searchQuery,
        },
      ]);
    } else {
      setColumnFilters([]);
    }
  }, [searchQuery]);

  // Generate WhatsApp URL function
  const generateWhatsappUrl = (
    phone: string,
    name: string,
    balance: number,
  ) => {
    const message = `Hello ${name}, your current balance is PKR ${balance?.toFixed(2)}`;
    return `https://wa.me/${phone}?text=${encodeURIComponent(message)}`;
  };

  // Define columns
  const columns: ColumnDef<Balance>[] = [
    {
      accessorKey: "employee_id",
      header: "ID",
    },
    {
      accessorKey: "user_name",
      header: "Name",
      cell: ({ row }) => (
        <div className="font-medium">{`${row.original.user_name} (${row.original.user_department})`}</div>
      ),
      filterFn: (row, id, value) => {
        return (
          row
            .getValue(id)
            ?.toString()
            .toLowerCase()
            .includes(value.toLowerCase()) ?? false
        );
      },
    },
    {
      accessorKey: "user_phone",
      header: "Phone",
      cell: ({ row }) => (
        <div className="flex space-x-2">
          <span className="p-1 bg-zinc-700 w-28 text-center rounded">
            {row.original.user_phone}
          </span>
          <CopyButton value={row.original.user_phone} />
        </div>
      ),
    },
    {
      accessorKey: "balance",
      header: "Balance (PKR)",
      cell: ({ row }) => row.original.balance?.toFixed(2),
    },
    {
      id: "actions",
      header: "Actions",
      cell: ({ row }) => {
        const balance = row.original;

        return (
          <div className="flex items-center space-x-1">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onViewTransactions(balance.employee_id)}
              className="h-8 w-8 p-0"
              title="View Transactions"
            >
              <CreditCard className="h-4 w-4" />
              <span className="sr-only">View Transactions</span>
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() =>
                window.open(
                  generateWhatsappUrl(
                    balance.user_phone,
                    balance.user_name,
                    balance.balance,
                  ),
                  "_blank",
                )
              }
              disabled={sendingNotification || !balance.user_phone}
              className="h-8 w-8 p-0"
              title="Send Balance Notification"
            >
              <SendToBack className="h-4 w-4" />
              <span className="sr-only">Send Balance Notification</span>
            </Button>
            {admin && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onSendBalanceNotification(balance.employee_id)}
                disabled={
                  sendingNotification ||
                  !whatsappStatus.connected ||
                  !balance.user_phone
                }
                className="h-8 w-8 p-0"
                title="Send Balance Notification"
              >
                <Send className="h-4 w-4" />
                <span className="sr-only">Send Balance Notification</span>
              </Button>
            )}
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" className="h-8 w-8 p-0">
                  <span className="sr-only">Open menu</span>
                  <MoreHorizontal className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onClick={() => onEdit(balance)}>
                  <Pencil className="mr-2 h-4 w-4" />
                  Edit
                </DropdownMenuItem>
                <DropdownMenuItem
                  disabled={!admin}
                  onClick={() => onDelete(balance.user_id)}
                  className="text-red-600"
                >
                  <Trash2 className="mr-2 h-4 w-4" />
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        );
      },
    },
  ];

  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    onColumnFiltersChange: setColumnFilters,
    getFilteredRowModel: getFilteredRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    state: {
      sorting,
      columnFilters,
      columnVisibility,
    },
  });

  return (
    <div>
      {/* Search input */}
      <div className="flex items-end py-4">
        <div className="relative w-full max-w-sm">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search by Name..."
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

      <div className="overflow-x-auto">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead
                    key={header.id}
                    className={header.id === "actions" ? "w-[150px]" : ""}
                  >
                    {header.isPlaceholder
                      ? null
                      : flexRender(
                          header.column.columnDef.header,
                          header.getContext(),
                        )}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && "selected"}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext(),
                      )}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell
                  colSpan={columns.length}
                  className="h-24 text-center"
                >
                  No results found.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>

        {/* Pagination controls */}
        <div className="flex items-center justify-end space-x-2 py-4">
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.previousPage()}
            disabled={!table.getCanPreviousPage()}
          >
            Previous
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.nextPage()}
            disabled={!table.getCanNextPage()}
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  );
}
