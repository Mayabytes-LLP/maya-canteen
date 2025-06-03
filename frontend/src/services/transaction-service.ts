import { z } from "zod";

// Using Vite's proxy instead of hardcoded URL
const API_BASE = import.meta.env.VITE_API_BASE_URL || "/api";

export interface Transaction {
  id: number;
  user_id: number;
  user_name: string;
  amount: number;
  description: string;
  transaction_type: "deposit" | "purchase";
  created_at: string;
  updated_at: string;
  products?: TransactionProduct[];
}

export interface TransactionProduct {
  id: number;
  transaction_id: number;
  product_id: number;
  product_name: string;
  quantity: number;
  unit_price: number;
  is_single_unit: boolean;
  created_at: string;
  updated_at: string;
}

export interface ProductSalesSummary {
  product_id: number;
  product_name: string;
  product_type: string;
  total_quantity: number;
  total_sales: number;
  single_unit_sold: number;
  full_unit_sold: number;
}

export interface TransactionProductDetail {
  id: number;
  transaction_id: number;
  product_id: number;
  product_name: string;
  product_type: string;
  quantity: number;
  unit_price: number;
  total_price: number;
  is_single_unit: boolean;
  created_at: string;
  updated_at: string;
}

export interface EmployeeTransaction extends Transaction {
  employee_id: string;
  department: string;
}

export interface User {
  id: number;
  name: string;
  employee_id: string;
  department: string;
  phone?: string;
  active: boolean;
  last_notification?: string;
  created_at: string;
  updated_at: string;
}

export interface Product {
  id: number;
  name: string;
  description: string;
  price: number;
  type: "regular" | "cigarette";
  active: boolean;
  is_single_unit: boolean;
  single_unit_price: number;
  created_at: string;
  updated_at: string;
}

export interface DateRangeRequest {
  startDate: string;
  endDate: string;
}

export interface UserBalance {
  user_id: number;
  user_name: string;
  employee_id: string;
  user_department: string;
  user_phone: string;
  user_active: boolean;
  last_notification: string;
  balance: number;
}

export interface UserCSVResponse {
  success: number;
  failed: number;
  errors: string[];
}

export const Departments = [
  "Creative Design Dept",
  "Operations Dept",
  "Development Dept",
  "Business Development Dept",
  "Digital Marketing Dept",
  "Admin Department",
  "Sales Department Code Coffee",
  "Human Resources",
  "IT Support",
] as const;

export const zodUserSchema = z.object({
  name: z.string().min(2, "Name must be at least 2 characters"),
  // employee_id must be a string of numbers
  employee_id: z.string().refine((val) => /^\d+$/.test(val), {
    message: "Employee ID must be a number",
  }),
  department: z.string().min(2, "Department must be at least 2 characters"),
  phone: z
    .string()
    .optional()
    .transform((val) => {
      if (!val) return "";

      // Format Pakistan numbers
      if (val.startsWith("0")) {
        return `+92${val.slice(1)}`;
      }

      // Keep existing country codes
      if (val.startsWith("+")) {
        return val;
      }

      return val;
    }),
  active: z.boolean().default(true),
});

// Form validation schema
export const zodProductSchema = z.object({
  name: z.string().min(2, "Name must be at least 2 characters"),
  description: z.string(),
  price: z.coerce.number().positive("Price must be a positive number"),
  type: z.enum(["regular", "cigarette"]),
  active: z.boolean().default(true),
  is_single_unit: z.boolean().default(false),
  single_unit_price: z.coerce.number().min(0),
});

export const transactionService = {
  async getAllTransactions(): Promise<Transaction[]> {
    const response = await fetch(`${API_BASE}/transactions`);
    if (!response.ok) {
      throw new Error("Failed to fetch transactions");
    }
    const res = await response.json();
    return res.data;
  },

  async getLatestTransactions(limit: number = 10): Promise<Transaction[]> {
    const response = await fetch(
      `${API_BASE}/transactions/latest?limit=${limit}`
    );
    if (!response.ok) {
      throw new Error("Failed to fetch latest transactions");
    }

    const res = await response.json();

    console.log(res.data);
    return res.data;
  },

  async getTransaction(id: number): Promise<Transaction> {
    const response = await fetch(`${API_BASE}/transactions/${id}`);
    if (!response.ok) {
      throw new Error("Failed to fetch transaction");
    }
    const res = await response.json();
    return res.data;
  },

  async getTransactionProducts(id: number): Promise<TransactionProduct[]> {
    const response = await fetch(`${API_BASE}/transactions/${id}/products`);
    if (!response.ok) {
      throw new Error("Failed to fetch transaction products");
    }
    const res = await response.json();
    return res.data;
  },

  async createTransaction(
    transaction: Omit<
      Transaction,
      "id" | "user_name" | "created_at" | "updated_at"
    > & {
      products?: Omit<
        TransactionProduct,
        "id" | "transaction_id" | "created_at" | "updated_at"
      >[];
    }
  ): Promise<Transaction> {
    const response = await fetch(`${API_BASE}/transactions`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(transaction),
    });
    if (!response.ok) {
      throw new Error("Failed to create transaction");
    }
    const res = await response.json();
    return res.data;
  },

  async getTransactionsByDateRange(
    dateRange: DateRangeRequest
  ): Promise<Transaction[]> {
    const response = await fetch(`${API_BASE}/transactions/date-range`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(dateRange),
    });
    if (!response.ok) {
      throw new Error("Failed to fetch transactions by date range");
    }
    const res = await response.json();
    return res.data;
  },

  async getProductSalesSummary(
    dateRange: DateRangeRequest
  ): Promise<ProductSalesSummary[]> {
    const response = await fetch(`${API_BASE}/reports/product-sales`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(dateRange),
    });
    if (!response.ok) {
      throw new Error("Failed to fetch product sales summary");
    }
    const res = await response.json();
    return res.data;
  },

  async getTransactionProductDetails(
    dateRange: DateRangeRequest
  ): Promise<TransactionProductDetail[]> {
    const response = await fetch(`${API_BASE}/reports/transaction-products`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(dateRange),
    });
    if (!response.ok) {
      throw new Error("Failed to fetch transaction product details");
    }
    const res = await response.json();
    return res.data;
  },

  async getAllUsers(): Promise<User[]> {
    const response = await fetch(`${API_BASE}/users`);
    if (!response.ok) {
      throw new Error("Failed to fetch users");
    }
    const res = await response.json();
    return res.data;
  },

  async getAllProducts(): Promise<Product[]> {
    const response = await fetch(`${API_BASE}/products`);
    if (!response.ok) {
      throw new Error("Failed to fetch products");
    }
    const res = await response.json();
    return res.data;
  },

  async createProduct(
    product: Omit<Product, "id" | "created_at" | "updated_at">
  ): Promise<Product> {
    const response = await fetch(`${API_BASE}/products`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(product),
    });
    if (!response.ok) {
      console.error(await response.text());
      throw new Error("Failed to create product");
    }
    const res = await response.json();
    return res.data;
  },

  async updateProduct(
    product: Omit<Product, "created_at" | "updated_at">
  ): Promise<Product> {
    console.log(product);
    const response = await fetch(`${API_BASE}/products/${product.id}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(product),
    });
    if (!response.ok) {
      throw new Error("Failed to update user");
    }
    const res = await response.json();
    return res.data;
  },

  async deleteProduct(id: number): Promise<void> {
    const response = await fetch(`${API_BASE}/products/${id}`, {
      method: "DELETE",
    });
    if (!response.ok) {
      throw new Error("Failed to delete product");
    }
  },

  async createUser(
    user: Omit<User, "id" | "created_at" | "updated_at">
  ): Promise<User> {
    const response = await fetch(`${API_BASE}/users`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(user),
    });
    if (!response.ok) {
      throw new Error("Failed to create user");
    }
    const res = await response.json();
    return res.data;
  },

  async importUsers(file: File): Promise<UserCSVResponse> {
    const formData = new FormData();
    formData.append("csv_file", file);

    const response = await fetch(`${API_BASE}/users/import`, {
      method: "POST",
      body: formData,
    });

    if (!response.ok) {
      throw new Error("Failed to import users");
    }

    const res = await response.json();
    return res.data;
  },

  async getUser(id: string): Promise<User> {
    // make sure number is 5 digits
    const paddedId = id.padStart(5, "0");
    const response = await fetch(`${API_BASE}/users/${paddedId}`);
    if (!response.ok) {
      throw new Error("Failed to fetch user");
    }
    const res = await response.json();
    return res.data;
  },

  async updateUser(
    user: Pick<
      User,
      "id" | "name" | "employee_id" | "department" | "phone" | "active"
    >
  ): Promise<User> {
    const response = await fetch(`${API_BASE}/users/${user.id}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(user),
    });
    if (!response.ok) {
      throw new Error("Failed to update user");
    }
    const res = await response.json();
    return res.data;
  },

  async deleteUser(id: number): Promise<void> {
    const response = await fetch(`${API_BASE}/users/${id}`, {
      method: "DELETE",
    });
    if (!response.ok) {
      throw new Error("Failed to delete user");
    }
  },

  async updateTransaction(
    transaction: Omit<Transaction, "created_at" | "updated_at">
  ): Promise<Transaction> {
    const response = await fetch(`${API_BASE}/transactions/${transaction.id}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(transaction),
    });
    if (!response.ok) {
      throw new Error("Failed to update transaction");
    }
    const res = await response.json();
    return res.data;
  },

  async deleteTransaction(id: number): Promise<void> {
    const response = await fetch(`${API_BASE}/transactions/${id}`, {
      method: "DELETE",
    });
    if (!response.ok) {
      throw new Error("Failed to delete transaction");
    }
  },

  // default limit 50
  async getTransactionsByUserId(
    userId: string,
    limit = 50
  ): Promise<EmployeeTransaction[]> {
    // make sure number is 5 digits
    const paddedId = userId.padStart(5, "0");
    const response = await fetch(
      `${API_BASE}/users/${paddedId}/transactions?limit=${limit}`
    );
    if (!response.ok) {
      throw new Error("Failed to fetch user transactions");
    }
    const res = await response.json();
    return res.data;
  },

  async getUsersBalances(): Promise<UserBalance[]> {
    const response = await fetch(`${API_BASE}/users/balances`);
    if (!response.ok) {
      throw new Error("Failed to fetch user balances");
    }
    const res = await response.json();
    return res.data;
  },

  async getBalanceByUserId(userId: number): Promise<UserBalance> {
    const response = await fetch(`${API_BASE}/users/${userId}/balance`);
    if (!response.ok) {
      throw new Error("Failed to fetch user balance");
    }
    const res = await response.json();
    return res.data;
  },

  async uploadUsersCsv(file: File): Promise<UserCSVResponse> {
    const formData = new FormData();
    formData.append("file", file);

    const response = await fetch(`${API_BASE}/users/csv`, {
      method: "POST",
      body: formData,
    });

    if (!response.ok) {
      throw new Error("Failed to upload CSV");
    }

    const res = await response.json();
    return res.data;
  },

  // Add new functions for WhatsApp notification
  async sendBalanceNotification(
    employeeId: string,
    messageTemplate?: string,
    month?: string,
    year?: number
  ): Promise<{ success: boolean; message?: string }> {
    const response = await fetch(`${API_BASE}/whatsapp/notify/${employeeId}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        message_template: messageTemplate,
        month: month,
        year: year,
      }),
    });

    if (!response.ok) {
      const error = await response.json();
      return { success: false, message: error.message };
    }

    return { success: true };
  },

  async sendAllBalanceNotifications(
    messageTemplate?: string,
    month?: string,
    year?: number
  ): Promise<{
    success: boolean;
    data?: {
      details: {
        fail_count: number;
        failed_users: string[];
        success_count: number;
      };
      message: string;
      success: boolean;
    };
    message?: string;
  }> {
    const response = await fetch(`${API_BASE}/whatsapp/notify-all`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(
        messageTemplate
          ? { message_template: messageTemplate, month: month, year: year }
          : {}
      ),
    });

    if (!response.ok) {
      const error = await response.json();
      return { success: false, message: error.message };
    }

    const data = await response.json();
    return { success: true, data: data.data };
  },
};
