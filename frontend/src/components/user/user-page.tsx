import { useState } from "react";

import DepositForm from "@/components/deposit-form";
import ErrorBoundary from "../error-boundary";
import UserBalances from "./user-balances";
import UserForm from "./user-form";
// import UserList from "./user-list";

export default function UserPage() {
  const [refreshTrigger, setRefreshTrigger] = useState(0);

  // Function to trigger a refresh of the user list
  const handleUserAdded = () => {
    setRefreshTrigger((prev) => prev + 1);
  };

  return (
    <div className="container mx-auto py-6 space-y-6">
      <h1 className="text-3xl font-bold">User Management</h1>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div>
          <UserForm onUserAdded={handleUserAdded} />
        </div>
        <div>
          <DepositForm onTransactionAdded={handleUserAdded} />
        </div>
        {/* <div> */}
        {/*   <ErrorBoundary> */}
        {/*     <UserList refreshTrigger={refreshTrigger} /> */}
        {/*   </ErrorBoundary> */}
        {/* </div> */}
        {/**/}
      </div>

      <div>
        <ErrorBoundary>
          <UserBalances refreshTrigger={refreshTrigger} />
        </ErrorBoundary>
      </div>
    </div>
  );
}
