import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import * as z from "zod";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { transactionService } from "@/services/transaction-service";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@components/ui/popover";
import { cn } from "@/lib/utils";
import { Check, ChevronsUpDown } from "lucide-react";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@components/ui/command";

const Departments = [
  "HR",
  "Design",
  "Sales",
  "PMO",
  "Development",
  "Operations",
  "Admin",
] as const;

// Form validation schema
const formSchema = z.object({
  name: z.string().min(2, "Name must be at least 2 characters"),
  employee_id: z.string().min(2, "Employee ID must be at least 2 characters"),
  department: z.enum(Departments),
  phone: z
    .string()
    .trim()
    .refine(
      (val) => /^0[3-9][0-9]{9}$/.test(val) || /^\+92[0-9]{10}$/.test(val),
      {
        message:
          "Invalid phone number format. Expected format: +923XXXXXXXXX or 03XXXXXXXXX",
      },
    )
    .transform((val) => (val.startsWith("0") ? `+92${val.slice(1)}` : val)),
});

type FormValues = z.infer<typeof formSchema>;

interface UserFormProps {
  onUserAdded: () => void;
}

export default function UserForm({ onUserAdded }: UserFormProps) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [userPopover, setUserPopover] = useState(false);

  // Initialize form
  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: "",
      employee_id: "",
      phone: "",
    },
  });

  const onSubmit = async (data: FormValues) => {
    setIsSubmitting(true);
    try {
      await transactionService.createUser({
        name: data.name,
        employee_id: data.employee_id,
        department: data.department,
        phone: data.phone,
      });

      toast.success("User added successfully");
      form.reset();
      onUserAdded();
    } catch (error) {
      console.error("Error adding user:", error);
      toast.error("Failed to add user");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>Add New User</CardTitle>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Name</FormLabel>
                  <FormControl>
                    <Input placeholder="Enter user name" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="employee_id"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Employee ID</FormLabel>
                  <FormControl>
                    <Input placeholder="Enter employee ID" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="department"
              render={({ field }) => (
                <FormItem className="flex flex-col items-stretch">
                  <FormLabel>Department</FormLabel>
                  <Popover open={userPopover} onOpenChange={setUserPopover}>
                    <PopoverTrigger asChild>
                      <FormControl>
                        <Button
                          variant="outline"
                          role="combobox"
                          className={cn(
                            "w-full justify-between",
                            !field.value && "text-muted-foreground",
                          )}
                        >
                          {field.value || "Select Department"}
                          <ChevronsUpDown className="opacity-50" />
                        </Button>
                      </FormControl>
                    </PopoverTrigger>
                    <PopoverContent className="w-full p-0">
                      <Command>
                        <CommandInput
                          placeholder="Search Employee..."
                          className="h-9"
                        />
                        <CommandList>
                          <CommandEmpty>No Employee found.</CommandEmpty>
                          <CommandGroup>
                            {Departments.map((department) => (
                              <CommandItem
                                key={department}
                                value={department}
                                onSelect={() => {
                                  form.setValue("department", department);
                                  setUserPopover(false);
                                }}
                              >
                                {department}
                                <Check
                                  className={cn(
                                    "ml-auto",
                                    department === field.value
                                      ? "opacity-100"
                                      : "opacity-0",
                                  )}
                                />
                              </CommandItem>
                            ))}
                          </CommandGroup>
                        </CommandList>
                      </Command>
                    </PopoverContent>
                  </Popover>
                  <FormDescription>Select The Department</FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="phone"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Phone Number format ( +923452324442 )</FormLabel>
                  <FormControl>
                    <Input placeholder="+92 345 2324442" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? "Adding..." : "Add User"}
            </Button>
          </form>
        </Form>
      </CardContent>
    </Card>
  );
}
