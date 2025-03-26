import {
  DateRangeRequest,
  Product,
  Transaction,
  transactionService,
  User,
  UserBalance,
} from "@/services/transaction-service";
import {
  useMutation,
  useQuery,
  useQueryClient,
  UseQueryOptions,
} from "@tanstack/react-query";

// Query keys
export const queryKeys = {
  transactions: ["transactions"],
  transaction: (id: number) => ["transaction", id],
  latestTransactions: (limit: number) => ["transactions", "latest", limit],
  transactionsByDateRange: (dateRange: DateRangeRequest) => [
    "transactions",
    "date-range",
    dateRange,
  ],
  transactionsByUser: (userId: string, limit: number) => [
    "transactions",
    "user",
    userId,
    limit,
  ],
  users: ["users"],
  user: (id: string) => ["user", id],
  usersBalances: ["users", "balances"],
  products: ["products"],
};

// Transaction queries
export const useAllTransactions = (
  options?: UseQueryOptions<Transaction[]>,
) => {
  return useQuery({
    queryKey: queryKeys.transactions,
    queryFn: () => transactionService.getAllTransactions(),
    ...options,
  });
};

export const useLatestTransactions = (
  limit: number = 10,
  options?: UseQueryOptions<Transaction[]>,
) => {
  return useQuery({
    queryKey: queryKeys.latestTransactions(limit),
    queryFn: () => transactionService.getLatestTransactions(limit),
    ...options,
  });
};

export const useTransaction = (
  id: number,
  options?: UseQueryOptions<Transaction>,
) => {
  return useQuery({
    queryKey: queryKeys.transaction(id),
    queryFn: () => transactionService.getTransaction(id),
    enabled: !!id,
    ...options,
  });
};

export const useTransactionsByDateRange = (
  dateRange: DateRangeRequest,
  options?: UseQueryOptions<Transaction[]>,
) => {
  return useQuery({
    queryKey: queryKeys.transactionsByDateRange(dateRange),
    queryFn: () => transactionService.getTransactionsByDateRange(dateRange),
    enabled: !!dateRange.startDate && !!dateRange.endDate,
    ...options,
  });
};

export const useTransactionsByUser = (
  userId: string,
  limit: number = 10,
  options?: UseQueryOptions<Transaction[]>,
) => {
  return useQuery({
    queryKey: queryKeys.transactionsByUser(userId, limit),
    queryFn: () => transactionService.getTransactionsByUserId(userId, limit),
    enabled: !!userId,
    ...options,
  });
};

// User queries
export const useAllUsers = (options?: UseQueryOptions<User[]>) => {
  return useQuery({
    queryKey: queryKeys.users,
    queryFn: () => transactionService.getAllUsers(),
    ...options,
  });
};

export const useUser = (id: string, options?: UseQueryOptions<User>) => {
  return useQuery({
    queryKey: queryKeys.user(id),
    queryFn: () => transactionService.getUser(id),
    enabled: !!id,
    ...options,
  });
};

export const useUsersBalances = (options?: UseQueryOptions<UserBalance[]>) => {
  return useQuery({
    queryKey: queryKeys.usersBalances,
    queryFn: () => transactionService.getUsersBalances(),
    ...options,
  });
};

// Product queries
export const useAllProducts = (options?: UseQueryOptions<Product[]>) => {
  return useQuery({
    queryKey: queryKeys.products,
    queryFn: () => transactionService.getAllProducts(),
    ...options,
  });
};

// Mutations
export const useCreateTransaction = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (
      transaction: Omit<
        Transaction,
        "id" | "user_name" | "created_at" | "updated_at"
      >,
    ) => transactionService.createTransaction(transaction),
    onSuccess: () => {
      // Invalidate and refetch transactions lists
      queryClient.invalidateQueries({ queryKey: queryKeys.transactions });
      // You might want to be more specific about which queries to invalidate
      queryClient.invalidateQueries({ queryKey: ["transactions", "latest"] });
      queryClient.invalidateQueries({ queryKey: queryKeys.usersBalances });
    },
  });
};

export const useUpdateTransaction = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (transaction: Omit<Transaction, "created_at" | "updated_at">) =>
      transactionService.updateTransaction(transaction),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.transactions });
      queryClient.invalidateQueries({
        queryKey: queryKeys.transaction(data.id),
      });
      queryClient.invalidateQueries({ queryKey: queryKeys.usersBalances });
      // Invalidate user transactions
      queryClient.invalidateQueries({
        queryKey: ["transactions", "user", data.user_id.toString()],
      });
    },
  });
};

export const useDeleteTransaction = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => transactionService.deleteTransaction(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.transactions });
      queryClient.removeQueries({ queryKey: queryKeys.transaction(id) });
      queryClient.invalidateQueries({ queryKey: queryKeys.usersBalances });
    },
  });
};

export const useCreateProduct = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (product: Omit<Product, "id" | "created_at" | "updated_at">) =>
      transactionService.createProduct(product),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.products });
    },
  });
};

export const useCreateUser = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (user: Omit<User, "id" | "created_at" | "updated_at">) =>
      transactionService.createUser(user),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users });
    },
  });
};

export const useUpdateUser = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (
      user: Pick<
        User,
        "id" | "name" | "employee_id" | "department" | "phone" | "active"
      >,
    ) => transactionService.updateUser(user),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users });
      queryClient.invalidateQueries({
        queryKey: queryKeys.user(data.employee_id),
      });
      queryClient.invalidateQueries({ queryKey: queryKeys.usersBalances });
    },
  });
};

export const useDeleteUser = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => transactionService.deleteUser(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users });
      queryClient.invalidateQueries({ queryKey: queryKeys.usersBalances });
    },
  });
};
