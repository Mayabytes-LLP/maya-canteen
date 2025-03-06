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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { transactionService } from "@/services/transaction-service";
import { Checkbox } from "../ui/checkbox";

// Form validation schema
const formSchema = z.object({
  name: z.string().min(2, "Name must be at least 2 characters"),
  description: z.string().min(5, "Description must be at least 5 characters"),
  price: z.coerce.number().positive("Price must be a positive number"),
  type: z.enum(["regular", "cigarette"]),
  is_single_unit: z.boolean().default(false),
  single_unit_price: z.coerce.number().min(0),
});

type FormValues = z.infer<typeof formSchema>;

interface ProductFormProps {
  onProductAdded: () => void;
}

export default function ProductForm({ onProductAdded }: ProductFormProps) {
  const [isSubmitting, setIsSubmitting] = useState(false);

  const defaultValues: FormValues = {
    name: "",
    description: "",
    price: 5,
    type: "regular",
    is_single_unit: false,
    single_unit_price: 0,
  };

  // Initialize form
  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues,
  });

  const onSubmit = async (data: FormValues) => {
    setIsSubmitting(true);
    try {
      await transactionService.createProduct({
        name: data.name,
        description: data.description,
        price: data.price,
        type: data.type,
        is_single_unit: data.is_single_unit,
        single_unit_price: data.single_unit_price,
      });

      toast.success("Product added successfully");
      form.reset(defaultValues);
      onProductAdded();
    } catch (error) {
      console.error("Error adding product:", error);
      toast.error("Failed to add product");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>Add New Product</CardTitle>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Product Name</FormLabel>
                  <FormControl>
                    <Input placeholder="Enter product name" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Description</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder="Enter product description"
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="type"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Product Type</FormLabel>
                  <Select
                    onValueChange={field.onChange}
                    defaultValue={field.value}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select product type" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="regular">Regular</SelectItem>
                      <SelectItem value="cigarette">Cigarette</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />

            {form.watch("type") === "cigarette" && (
              <>
                <FormField
                  control={form.control}
                  name="is_single_unit"
                  render={({ field }) => (
                    <FormItem className="flex flex-row items-start space-x-3">
                      <FormControl>
                        <Checkbox
                          checked={field.value}
                          onCheckedChange={field.onChange}
                        />
                      </FormControl>
                      <FormLabel>Single Unit</FormLabel>
                    </FormItem>
                  )}
                />
                {form.watch("is_single_unit") && (
                  <FormField
                    control={form.control}
                    name="single_unit_price"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Single unit Price</FormLabel>
                        <FormControl>
                          <Input
                            step="1"
                            placeholder="0"
                            type="number"
                            {...field}
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                )}
              </>
            )}

            <FormField
              control={form.control}
              name="price"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Price</FormLabel>
                  <FormControl>
                    <Input type="number" step="1" placeholder="5" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? "Adding..." : "Add Product"}
            </Button>
          </form>
        </Form>
      </CardContent>
    </Card>
  );
}
