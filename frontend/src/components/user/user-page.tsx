import { useState } from "react";

import ErrorBoundary from "@/components/error-boundary";
import UserBalances from "./user-balances";
import UserForm from "./user-form";

export default function UserPage() {
  const [refreshTrigger, setRefreshTrigger] = useState(0);

  const handleUserAdded = () => {
    setRefreshTrigger((prev) => prev + 1);
  };

  return (
    <div className="container mx-auto py-6 space-y-6">
      <h1 className="text-3xl font-bold">User Management</h1>

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
