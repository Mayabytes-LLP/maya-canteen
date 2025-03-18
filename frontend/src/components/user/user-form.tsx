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
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { transactionService } from "@/services/transaction-service";

// Form validation schema
const formSchema = z.object({
  name: z.string().min(2, "Name must be at least 2 characters"),
  employee_id: z.string().min(2, "Employee ID must be at least 2 characters"),
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
