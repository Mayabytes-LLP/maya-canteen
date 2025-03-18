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
import { Trash2 } from "lucide-react";

import {
  Product,
  transactionService,
  User,
} from "@/services/transaction-service";

import { AppContext } from "./canteen-provider";

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

  const { admin, currentUser } = useContext(AppContext);

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
    if (!currentUser?.id) {
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
        // Handle purchase transaction
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
        // Handle deposit transaction
        description = data.description || "Cash Deposit";
        finalAmount = parseFloat(data.amount || "0");
      }

      const transaction = {
        user_id: currentUser.id, // Convert to number as backend expects
        amount: finalAmount,
        description: description,
        transaction_type: data.transaction_type,
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
                      <SelectTrigger>
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
                <FormItem>
                  <FormLabel>Employee</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select employee" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {users.map((user) => (
                        <SelectItem
                          key={user.id.toString()}
                          value={user.id.toString()}
                        >
                          {user.name} ({user.employee_id})
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
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
              // Show product selection and cart only for purchase
              <>
                <div className="flex space-x-4">
                  <FormField
                    control={form.control}
                    name="product_id"
                    render={({ field }) => (
                      <FormItem className="flex-grow">
                        <FormLabel>Product</FormLabel>
                        <Select
                          onValueChange={field.onChange}
                          value={field.value}
                        >
                          <FormControl>
                            <SelectTrigger>
                              <SelectValue placeholder="Select product" />
                            </SelectTrigger>
                          </FormControl>
                          <SelectContent>
                            {products.map((product) => (
                              <SelectItem
                                key={product.id}
                                value={product.id.toString()}
                              >
                                {product.name}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
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
                              <div key={product.id} className="flex flex-col">
                                <FormItem>
                                  <FormLabel>Single Unit</FormLabel>
                                  <FormControl>
                                    <input
                                      type="checkbox"
                                      checked={isSingleUnit}
                                      onChange={(e) =>
                                        setIsSingleUnit(e.target.checked)
                                      }
                                    />
                                  </FormControl>
                                </FormItem>
                                <span className="font-semibold">
                                  Price: PKR.
                                  {isSingleUnit
                                    ? product.single_unit_price
                                    : product.price}
                                </span>
                              </div>
                            );
                          }
                          return (
                            <div key={product.id} className="flex flex-col">
                              <span className="font-semibold">
                                Price: PKR.{product.price}
                              </span>
                            </div>
                          );
                        }
                      })}
                      <div className="pt-6">
                        <Button
                          type="button"
                          onClick={handleAddToCart}
                          variant="outline"
                        >
                          Add
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
