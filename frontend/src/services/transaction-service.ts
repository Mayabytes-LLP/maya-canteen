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
  user_phone: string;
  balance: number;
}

// Using Vite's proxy instead of hardcoded URL
const API_BASE = "/api";

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
    >
  ): Promise<Transaction> {
    console.log(transaction);
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
    user: Pick<User, "id" | "name" | "employee_id" | "phone">
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

  async getTransactionsByUserId(userId: string): Promise<Transaction[]> {
    const response = await fetch(`${API_BASE}/users/${userId}/transactions`);
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
};
