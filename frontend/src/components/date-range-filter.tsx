import { zodResolver } from "@hookform/resolvers/zod";
import { addDays, format } from "date-fns";
import { CalendarIcon } from "lucide-react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { toast } from "sonner";
import { useState } from "react";

import {
  Transaction,
  transactionService,
} from "@/services/transaction-service";

const FormSchema = z.object({
  data_range: z
    .object({
      from: z.date().optional(),
      to: z.date().optional(),
    })
    .superRefine((data, ctx) => {
      if (!data.from || !data.to) {
        return ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: "Please select both start and end dates",
        });
      }

      if (data.from > data.to) {
        return ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: "Start date should be before end date",
        });
      }
      return data;
    }),
});

interface DateRangeFilterProps {
  onTransactionsLoaded: (transactions: Transaction[]) => void;
}

export default function DateRangeFilter({
  onTransactionsLoaded,
}: DateRangeFilterProps) {
  const [loading, setLoading] = useState(false);

  const initialStartDate = {
    data_range: {
      from: new Date(),
      to: addDays(new Date(), 15),
    },
  };
  const form = useForm<z.infer<typeof FormSchema>>({
    resolver: zodResolver(FormSchema),
    defaultValues: initialStartDate,
  });

  async function onSubmit(data: z.infer<typeof FormSchema>) {
    if (!data.data_range?.from || !data.data_range?.to) {
      console.error("Please select both start and end dates");
      toast.error("Please select both start and end dates");
      return;
    }

    setLoading(true);

    try {
      const transactions = await transactionService.getTransactionsByDateRange({
        startDate: data.data_range.from.toISOString(),
        endDate: data.data_range.to.toISOString(),
      });
      onTransactionsLoaded(transactions);
      toast.success("Transactions filtered successfully");
    } catch (error) {
      toast.error("Failed to filter transactions");
      console.error(error);
    } finally {
      setLoading(false);
    }
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        <FormField
          control={form.control}
          name="data_range"
          render={({ field }) => (
            <FormItem className="flex flex-col">
              <FormLabel>Filtered Transactions </FormLabel>
              <Popover>
                <PopoverTrigger asChild>
                  <FormControl>
                    <Button
                      variant={"outline"}
                      className={cn(
                        "w-[240px] pl-3 text-left font-normal",
                        !field.value && "text-muted-foreground",
                      )}
                    >
                      {field.value.from ? (
                        field.value.to ? (
                          <>
                            {format(field.value.from, "LLL dd, y")} -{" "}
                            {format(field.value.to, "LLL dd, y")}
                          </>
                        ) : (
                          format(field.value.from, "LLL dd, y")
                        )
                      ) : (
                        <span>Pick a date</span>
                      )}
                      <CalendarIcon className="ml-auto h-4 w-4 opacity-50" />
                    </Button>
                  </FormControl>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0" align="start">
                  <Calendar
                    mode="range"
                    defaultMonth={field.value.from || new Date()}
                    selected={{ from: field.value.from!, to: field.value.to }}
                    onSelect={field.onChange}
                    disabled={(date) => date < new Date("2025-03-01")}
                    numberOfMonths={2}
                    initialFocus
                  />
                </PopoverContent>
              </Popover>
              <FormDescription>
                Your date of birth is used to calculate your age.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
        <Button type="submit">{loading ? "Fetching..." : "Filter"}</Button>
      </form>
    </Form>
  );
}
