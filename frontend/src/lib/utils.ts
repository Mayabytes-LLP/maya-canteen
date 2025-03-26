import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export const formatTransaction = (
  amount: number,
  type?: "deposit" | "purchase"
): string => {
  const formattedAmount = new Intl.NumberFormat("ur-pk", {
    style: "currency",
    currency: "PKR",
  }).format(amount);

  return type
    ? `${type === "deposit" ? "+" : "-"}${formattedAmount}`
    : formattedAmount;
};

export const formatDate = (dateString: string): string => {
  return new Date(dateString).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
};
