import { zodResolver } from "@hookform/resolvers/zod";
import { useContext, useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import * as z from "zod";

import { Badge } from "@/components/ui/badge";
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { Check, ChevronsUpDown, Trash2 } from "lucide-react";

import {
  Product,
  transactionService,
  User,
} from "@/services/transaction-service";

import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { AppContext } from "@/context";
import { cn } from "@/lib/utils";

import { Checkbox } from "@/components/ui/checkbox";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";

// Define cart item to represent a product and its quantity
interface CartItem {
  productId: number;
  productName: string;
  single?: boolean;
  quantity: number;
  price: number;
}

// Define form schema with Zod
const formSchema = z.object({
  user_id: z.string({
    required_error: "Please select an employee",
  }),
  product_id: z.string().optional(),
  quantity: z.string().default("1"),
  description: z.string().optional(),
  transaction_type: z.enum(["purchase", "deposit"]),
  amount: z.string().optional(),
});

type FormValues = z.infer<typeof formSchema>;

interface TransactionFormProps {
  onTransactionAdded: () => void;
}

export default function DepositForm({
  onTransactionAdded,
}: TransactionFormProps) {
  const [users, setUsers] = useState<User[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [cartItems, setCartItems] = useState<CartItem[]>([]);
  const [totalAmount, setTotalAmount] = useState<number>(0);
  const [isSingleUnit, setIsSingleUnit] = useState<boolean>(false);

  const [userPopover, setUserPopover] = useState<boolean>(false);
  const [productPopover, setProductPopover] = useState<boolean>(false);

  const { admin } = useContext(AppContext);

  const defaultValues: FormValues = {
    user_id: "",
    product_id: "",
    quantity: "1",
    description: "",
    transaction_type: "purchase", // Default transaction type
    amount: "",
  };

  // Initialize form
  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: defaultValues,
  });

  // Fetch users and products on component mount
  useEffect(() => {
    const fetchData = async () => {
      try {
        const [usersData, productsData] = await Promise.all([
          transactionService.getAllUsers(),
          transactionService.getAllProducts(),
        ]);
        setUsers(usersData ?? []);
        setProducts(productsData ?? []);
      } catch (error) {
        toast.error("Failed to load data");
        console.error(error);
      }
    };

    fetchData();
  }, []);

  // Calculate total amount whenever cart changes
  useEffect(() => {
    const total = cartItems.reduce(
      (sum, item) => sum + item.price * item.quantity,
      0
    );
    setTotalAmount(total);
  }, [cartItems]);

  // Handle adding product to cart
  const handleAddToCart = () => {
    const productId = parseInt(form.getValues("product_id") || "");
    if (!productId) {
      toast.error("Please select a product");
      return;
    }

    const quantity = parseInt(form.getValues("quantity") || "1");
    if (isNaN(quantity) || quantity <= 0) {
      toast.error("Quantity must be a positive number");
      return;
    }

    const selectedProduct = products.find(
      (product) => product.id === productId
    );

    if (!selectedProduct) {
      toast.error("Product not found");
      return;
    }

    // Determine the price based on whether the user selected a single unit or a full packet
    const price = isSingleUnit
      ? selectedProduct.single_unit_price
      : selectedProduct.price;

    // Check if product is already in cart
    const existingItemIndex = cartItems.findIndex(
      (item) => item.productId === productId && item.single === isSingleUnit
    );

    if (existingItemIndex >= 0) {
      // Update quantity if product already exists in cart
      const updatedCart = [...cartItems];
      updatedCart[existingItemIndex].quantity += quantity;
      setCartItems(updatedCart);
    } else {
      // Add new product to cart
      setCartItems([
        ...cartItems,
        {
          productId,
          productName: selectedProduct.name,
          single: isSingleUnit,
          quantity,
          price,
        },
      ]);
    }

    // Reset product selection and quantity
    form.setValue("product_id", "");
    form.setValue("quantity", "1");
    setIsSingleUnit(false);
    toast.success(`Added ${quantity} ${selectedProduct.name} to cart`);
  };

  // Remove item from cart
  const removeFromCart = (index: number) => {
    const newCart = [...cartItems];
    newCart.splice(index, 1);
    setCartItems(newCart);
  };

  // Handle form submission
  const onSubmit = async (data: FormValues) => {
    if (!data.user_id) {
      toast.error("Please select an employee");
      return;
    }
    if (data.transaction_type === "purchase" && cartItems.length === 0) {
      toast.error("Please add at least one product to the cart");
      return;
    }

    if (data.transaction_type === "deposit" && !data.amount) {
      toast.error("Please enter deposit amount");
      return;
    }

    setIsSubmitting(true);
    try {
      let finalAmount = 0;
      let description = "";

      if (data.transaction_type === "purchase") {
        const productsDescription = cartItems
          .map(
            (item) =>
              `${item.quantity}x at PKR.${item.price} ${item.productName} ${
                item.single ? "(Single Unit)" : ""
              }`
          )
          .filter(Boolean)
          .join(", ");

        description = data.description
          ? `${data.description} (${productsDescription})`
          : productsDescription;
        finalAmount = totalAmount;
      } else {
        description = data.description || "Cash Deposit";
        finalAmount = parseFloat(data.amount || "0");
      }

      const products =
        data.transaction_type === "purchase"
          ? cartItems.map((item) => ({
              product_id: item.productId,
              product_name: item.productName,
              quantity: item.quantity,
              unit_price: item.price,
              is_single_unit: !!item.single,
              transaction_id: 0, // Placeholder, will be set by the backend
              id: 0, // Placeholder, will be set by the backend
              created_at: new Date().toISOString(), // Placeholder
              updated_at: new Date().toISOString(), // Placeholder
            }))
          : undefined;

      const transaction = {
        user_id: Number(data.user_id),
        amount: finalAmount,
        description: description,
        transaction_type: data.transaction_type,
        products: products,
      };

      await transactionService.createTransaction(transaction);
      toast.success("Transaction added successfully");
      form.reset(defaultValues);
      setCartItems([]);
      onTransactionAdded();
    } catch (error) {
      toast.error("Failed to add transaction");
      console.error(error);
    } finally {
      setIsSubmitting(false);
    }
  };

  if ((!users || !products) && (users.length === 0 || products.length === 0)) {
    return <div>Loading...</div>;
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>New Canteen Transaction</CardTitle>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="transaction_type"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Transaction Type</FormLabel>
                  <Select
                    onValueChange={field.onChange}
                    defaultValue={field.value}
                  >
                    <FormControl>
                      <SelectTrigger className="w-full">
                        <SelectValue placeholder="Select type" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="purchase">Purchase</SelectItem>
                      <SelectItem disabled={!admin} value="deposit">
                        Deposit
                      </SelectItem>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="user_id"
              render={({ field }) => (
                <FormItem className="flex flex-col items-stretch">
                  <FormLabel>Employee</FormLabel>
                  <Popover open={userPopover} onOpenChange={setUserPopover}>
                    <PopoverTrigger asChild>
                      <FormControl>
                        <Button
                          variant="outline"
                          role="combobox"
                          className={cn(
                            "w-full justify-between",
                            !field.value && "text-muted-foreground"
                          )}
                        >
                          {field.value
                            ? (() => {
                                const cu = users.find(
                                  (user) => user.id.toString() === field.value
                                );

                                if (!cu) {
                                  return ``;
                                }

                                return `${cu.name} (${cu.department})`;
                              })()
                            : "Select User"}
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
                            {users.map((user) => (
                              <CommandItem
                                key={user.id.toString()}
                                value={user.name}
                                onSelect={() => {
                                  form.setValue("user_id", user.id.toString());
                                  setUserPopover(false);
                                }}
                              >
                                {user.name} {user.employee_id} (
                                {user.department})
                                <Check
                                  className={cn(
                                    "ml-auto",
                                    user.id.toString() === field.value
                                      ? "opacity-100"
                                      : "opacity-0"
                                  )}
                                />
                              </CommandItem>
                            ))}
                          </CommandGroup>
                        </CommandList>
                      </Command>
                    </PopoverContent>
                  </Popover>
                  <FormDescription>Select The Employee</FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            {form.watch("transaction_type") === "deposit" ? (
              <FormField
                control={form.control}
                name="amount"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Deposit Amount (PKR)</FormLabel>
                    <FormControl>
                      <Input
                        type="number"
                        min="0"
                        step="0.01"
                        placeholder="Enter amount"
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            ) : (
              <>
                <div className="flex align-top space-x-4">
                  <FormField
                    control={form.control}
                    name="product_id"
                    render={({ field }) => (
                      <FormItem className="w-1/3">
                        <FormLabel>Products</FormLabel>
                        <Popover
                          open={productPopover}
                          onOpenChange={setProductPopover}
                        >
                          <PopoverTrigger asChild>
                            <FormControl>
                              <Button
                                variant="outline"
                                role="combobox"
                                className={cn(
                                  "w-full justify-between",
                                  !field.value && "text-muted-foreground"
                                )}
                              >
                                {field.value
                                  ? (() => {
                                      const cu = products.find(
                                        (product) =>
                                          product.id.toString() === field.value
                                      );

                                      if (!cu) {
                                        return ``;
                                      }

                                      return `${cu.name}`;
                                    })()
                                  : "Select Product"}
                                <ChevronsUpDown className="opacity-50" />
                              </Button>
                            </FormControl>
                          </PopoverTrigger>
                          <PopoverContent className="w-full p-0">
                            <Command>
                              <CommandInput
                                placeholder="Search Product..."
                                className="h-9"
                              />
                              <CommandList>
                                <CommandEmpty>No Products found.</CommandEmpty>
                                <CommandGroup>
                                  {products.map((product) => (
                                    <CommandItem
                                      key={product.id.toString()}
                                      value={product.name}
                                      onSelect={() => {
                                        form.setValue(
                                          "product_id",
                                          product.id.toString()
                                        );
                                        setProductPopover(false);
                                      }}
                                    >
                                      {product.name}
                                      <Check
                                        className={cn(
                                          "ml-auto",
                                          product.id.toString() === field.value
                                            ? "opacity-100"
                                            : "opacity-0"
                                        )}
                                      />
                                    </CommandItem>
                                  ))}
                                </CommandGroup>
                              </CommandList>
                            </Command>
                          </PopoverContent>
                        </Popover>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  {form.watch("product_id") && (
                    <>
                      <FormField
                        control={form.control}
                        name="quantity"
                        render={({ field }) => (
                          <FormItem className="w-28">
                            <FormLabel>Quantity</FormLabel>
                            <FormControl>
                              <Input
                                type="number"
                                min="1"
                                step="1"
                                {...field}
                              />
                            </FormControl>
                            <FormMessage />
                          </FormItem>
                        )}
                      />
                      {products.map((product) => {
                        if (
                          product.id.toString() === form.watch("product_id")
                        ) {
                          if (product.type === "cigarette") {
                            return (
                              <FormItem
                                key={product.id}
                                className="flex flex-row items-start space-x-3 space-y-0 rounded-md border p-4 shadow flex-1"
                              >
                                <FormControl>
                                  <Checkbox
                                    className="h-6 w-6"
                                    checked={isSingleUnit}
                                    onCheckedChange={(val) => {
                                      if (typeof val === "string") {
                                        setIsSingleUnit(false);
                                      } else {
                                        setIsSingleUnit(val);
                                      }
                                    }}
                                  />
                                </FormControl>
                                <div className="space-y-1 leading-none whitespace-nowrap">
                                  <FormLabel>Single Unit</FormLabel>
                                  <FormDescription>
                                    <span className="text-xs text-muted-foreground">
                                      Price: PKR.
                                    </span>
                                    {isSingleUnit
                                      ? product.single_unit_price
                                      : product.price}
                                  </FormDescription>
                                </div>
                              </FormItem>
                            );
                          }
                          return (
                            <div
                              key={product.id}
                              className="flex h-14 items-center flex-col justify-end"
                            >
                              <span className="font-semibold">
                                Price: PKR.{product.price}
                              </span>
                            </div>
                          );
                        }
                      })}
                      <div className="ml-auto flex flex-col pt-6">
                        <Button
                          type="button"
                          onClick={handleAddToCart}
                          size="lg"
                          variant="secondary"
                        >
                          Add Rs.
                          {parseInt(form.watch("quantity") || "1") *
                            (isSingleUnit &&
                            products.length > 0 &&
                            form.watch("product_id")
                              ? products.find(
                                  (product) =>
                                    product.id.toString() ===
                                    form.watch("product_id")
                                )?.single_unit_price ?? 0
                              : products.find(
                                  (product) =>
                                    product.id.toString() ===
                                    form.watch("product_id")
                                )?.price ?? 0)}{" "}
                          to Cart
                        </Button>
                      </div>
                    </>
                  )}
                </div>

                {/* Cart summary */}
                {cartItems.length > 0 && (
                  <div className="border rounded-md p-4 space-y-2">
                    <h3 className="font-medium">Cart Items</h3>
                    <div className="space-y-2">
                      {cartItems.map((item, index) => (
                        <div
                          key={index}
                          className="flex justify-between items-center"
                        >
                          <div>
                            <Badge variant="outline" className="mr-2">
                              {item.quantity}x
                            </Badge>
                            {item.productName} {item.single && "(Single Unit)"}
                          </div>
                          <div className="flex items-center space-x-2">
                            <span>
                              PKR.{(item.price * item.quantity).toFixed(2)}
                            </span>
                            <Button
                              type="button"
                              variant="ghost"
                              size="sm"
                              className="h-8 w-8 p-0"
                              onClick={() => removeFromCart(index)}
                            >
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </div>
                        </div>
                      ))}
                    </div>
                    <div className="flex justify-between font-bold pt-2 border-t">
                      <span>Total:</span>
                      <span>PKR.{totalAmount.toFixed(2)}</span>
                    </div>
                  </div>
                )}
              </>
            )}

            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Additional Notes (Optional)</FormLabel>
                  <FormControl>
                    <Textarea {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <Button
              type="submit"
              className="w-full"
              disabled={
                isSubmitting ||
                (form.watch("transaction_type") === "purchase" &&
                  cartItems.length === 0) ||
                (form.watch("transaction_type") === "deposit" &&
                  !form.watch("amount"))
              }
            >
              {isSubmitting ? "Processing..." : "Submit Transaction"}
            </Button>
          </form>
        </Form>
      </CardContent>
    </Card>
  );
}
