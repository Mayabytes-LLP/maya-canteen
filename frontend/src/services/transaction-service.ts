import { z } from "zod";

// Using Vite's proxy instead of hardcoded URL
const API_BASE = import.meta.env.VITE_API_BASE_URL || "/api";

export interface Transaction {
  id: number;
  user_id: number;
  user_name: string;
  amount: number;
  description: string;
  transaction_type: string;
  created_at: string;
  updated_at: string;
}

export interface User {
  id: number;
  name: string;
  employee_id: string;
  department: string;
  phone: string;
  created_at: string;
  updated_at: string;
}

export interface Product {
  id: number;
  name: string;
  description: string;
  price: number;
  type: "regular" | "cigarette";
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
  employee_id: z.string().min(2, "Employee ID must be at least 2 characters"),
  department: z.enum(Departments),
  phone: z
    .string()
    .trim()
    .refine(
      (val) =>
        /^0[3-9][0-9]{9}$/.test(val) || // 03XX-XXXXXXX
        /^\+92[0-9]{10}$/.test(val) || // +92XXXXXXXXXX
        /^\+1[0-9]{10}$/.test(val) || // +1XXXXXXXXXX
        /^03[0-9]{2}[-\s]?[0-9]{7}$/.test(val), // 0311-5410355 or 0332 2723005
      {
        message:
          "Invalid phone number format. Expected: 03XXXXXXXXX, +92XXXXXXXXXX, +1XXXXXXXXXX or 0311-XXXXXXX",
      },
    )
    .transform((val) => {
      // Remove spaces and dashes
      val = val.replace(/[\s-]/g, "");

      // Handle Pakistan numbers
      if (val.startsWith("0")) {
        return `+92${val.slice(1)}`;
      }

      // Keep existing country codes
      if (val.startsWith("+")) {
        return val;
      }

      return val;
    }),
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
      `${API_BASE}/transactions/latest?limit=${limit}`,
    );
    if (!response.ok) {
      throw new Error("Failed to fetch latest transactions");
    }

    const res = await response.json();

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

  async createTransaction(
    transaction: Omit<
      Transaction,
      "id" | "user_name" | "created_at" | "updated_at"
    >,
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
    dateRange: DateRangeRequest,
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

  async getAllUsers(): Promise<User[]> {
    const response = await fetch(`${API_BASE}/users`);
    if (!response.ok) {
      throw new Error("Failed to fetch users");
    }

    const result = await response.json();
    return result.data;
  },

  async getAllProducts(): Promise<Product[]> {
    const response = await fetch(`${API_BASE}/products`);
    if (!response.ok) {
      throw new Error("Failed to fetch products");
    }
    const result = await response.json();
    return result.data;
  },

  async createProduct(
    product: Omit<Product, "id" | "created_at" | "updated_at">,
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
    product: Omit<Product, "created_at" | "updated_at">,
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
      throw new Error("Failed to delete user");
    }
  },

  async createUser(
    user: Omit<User, "id" | "created_at" | "updated_at">,
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
    user: Pick<User, "id" | "name" | "employee_id" | "department" | "phone">,
  ): Promise<User> {
    console.log(user);
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
    transaction: Omit<Transaction, "created_at" | "updated_at">,
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

  // default limit 10
  async getTransactionsByUserId(
    userId: string,
    limit = 10,
  ): Promise<Transaction[]> {
    const response = await fetch(
      `${API_BASE}/users/${userId}/transactions?limit=${limit}`,
    );
    if (!response.ok) {
      throw new Error("Failed to fetch user transactions");
    }
    const res = await response.json();
    console.log(userId, res.data);
    return res.data;
  },

  async getUsersBalances(): Promise<UserBalance[]> {
    try {
      const response = await fetch(`${API_BASE}/users/balances`);
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`Failed to fetch users' balances: ${errorText}`);
      }
      const res = await response.json();
      return res.data;
    } catch (error) {
      console.error("Error fetching users' balances:", error);
      throw error;
    }
  },

  async getBalanceByUserId(userId: number): Promise<UserBalance> {
    try {
      const response = await fetch(`${API_BASE}/users/${userId}/balance`);
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`Failed to fetch user's balance: ${errorText}`);
      }
      const res = await response.json();
      return res.data;
    } catch (error) {
      console.error("Error fetching user's balance:", error);
      throw error;
    }
  },

  async uploadUsersCsv(file: File): Promise<UserCSVResponse> {
    const formData = new FormData();
    formData.append("file", file);

    const response = await fetch(`${API_BASE}/users/upload-csv`, {
      method: "POST",
      body: formData,
    });

    if (!response.ok) {
      throw new Error("Failed to upload users CSV");
    }

    const res = await response.json();
    return res.data;
  },
};
