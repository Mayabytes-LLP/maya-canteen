import Papa from "papaparse";
import { useState } from "react";
import { toast } from "sonner";
import * as z from "zod";

import ErrorBoundary from "@/components/error-boundary";
import { Button } from "@/components/ui/button";
import { transactionService } from "@/services/transaction-service";
import UserBalances from "../components/user/user-balances";
import UserForm from "../components/user/user-form";

// User CSV schema matching the form validation
const userCsvSchema = z.object({
  name: z.string().min(2, "Name must be at least 2 characters"),
  employee_id: z.string().min(2, "Employee ID must be at least 2 characters"),
  department: z.string(),
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
      }
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
  active: z
    .union([z.boolean(), z.string()])
    .optional()
    .transform((val) => {
      if (typeof val === "boolean") return val;
      if (typeof val === "string") {
        const lowercaseVal = val.toLowerCase();
        if (["true", "yes", "1", "active"].includes(lowercaseVal)) return true;
        if (["false", "no", "0", "inactive"].includes(lowercaseVal))
          return false;
      }
      // Default to true if not specified or invalid
      return true;
    }),
});

type UserCsvRow = z.infer<typeof userCsvSchema>;

export default function UserPage() {
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  const [isUploading, setIsUploading] = useState(false);

  const handleUserAdded = () => {
    setRefreshTrigger((prev) => prev + 1);
  };

  const validateAndFormatCsvData = async (
    file: File
  ): Promise<{ data: UserCsvRow[]; errors: string[] }> => {
    return new Promise((resolve, reject) => {
      Papa.parse(file, {
        header: true,
        skipEmptyLines: true,
        complete: (results) => {
          const errors: string[] = [];
          const validRows: UserCsvRow[] = [];

          results.data.forEach((row, index: number) => {
            try {
              // Validate and format the row
              const validRow = userCsvSchema.parse(row);
              validRows.push(validRow);
            } catch (error) {
              if (error instanceof z.ZodError) {
                error.errors.forEach((err) => {
                  errors.push(
                    `Row ${index + 1}: ${err.path.join(".")} - ${err.message}`
                  );
                });
              }
            }
          });

          resolve({ data: validRows, errors });
        },
        error: (error) => {
          reject(new Error(`Failed to parse CSV: ${error.message}`));
        },
      });
    });
  };

  const handleFileUpload = async (
    event: React.ChangeEvent<HTMLInputElement>
  ) => {
    const file = event.target.files?.[0];
    if (!file) return;

    // Check if it's a CSV file
    if (!file.name.endsWith(".csv")) {
      toast.error("Please upload a CSV file");
      return;
    }

    setIsUploading(true);
    try {
      // First validate and format the CSV data
      const { data: validRows, errors: validationErrors } =
        await validateAndFormatCsvData(file);

      if (validationErrors.length > 0) {
        toast.error("CSV validation failed");
        console.error("Validation errors:", validationErrors);
        validationErrors.forEach((error) => toast.error(error));
        return;
      }

      // Create a new CSV file with the validated and formatted data
      const csvContent = Papa.unparse(validRows);
      const validatedFile = new File([csvContent], file.name, {
        type: "text/csv",
      });

      // Upload the validated file
      const result = await transactionService.uploadUsersCsv(validatedFile);
      toast.success(`Successfully added ${result.success} users`);
      if (result.failed > 0) {
        toast.error(`Failed to add ${result.failed} users`);
        console.error("Upload errors:", result.errors);
      }
      handleUserAdded();
    } catch (error) {
      toast.error("Failed to upload users CSV");
      console.error(error);
    } finally {
      setIsUploading(false);
      // Reset the input
      event.target.value = "";
    }
  };

  return (
    <div className="container mx-auto py-6 space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold">User Management</h1>

        <div className="flex items-center gap-4">
          <input
            type="file"
            accept=".csv"
            onChange={handleFileUpload}
            disabled={isUploading}
            className="hidden"
            id="csv-upload"
          />
          <Button
            variant="outline"
            onClick={() => document.getElementById("csv-upload")?.click()}
            disabled={isUploading}
          >
            {isUploading ? "Uploading..." : "Upload Users CSV"}
          </Button>
          <Button
            variant="outline"
            onClick={() => {
              // Create template CSV with properly formatted example data
              const csvContent = Papa.unparse([
                {
                  name: "John Doe",
                  employee_id: "12345",
                  department: "Development",
                  phone: "+923001234567",
                  active: "true",
                },
                {
                  name: "Jane Smith",
                  employee_id: "12346",
                  department: "HR",
                  phone: "03001234568",
                  active: "false",
                },
              ]);
              const blob = new Blob([csvContent], { type: "text/csv" });
              const url = window.URL.createObjectURL(blob);
              const a = document.createElement("a");
              a.href = url;
              a.download = "users_template.csv";
              a.click();
              window.URL.revokeObjectURL(url);
            }}
          >
            Download Template
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="col-span-2">
          <UserForm onUserAdded={handleUserAdded} />
        </div>
      </div>

      <div>
        <ErrorBoundary>
          <UserBalances refreshTrigger={refreshTrigger} />
        </ErrorBoundary>
      </div>
    </div>
  );
}
