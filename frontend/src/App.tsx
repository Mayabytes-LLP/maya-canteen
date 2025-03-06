import CanteenPage from "@/components/canteen-page";
import ProductPage from "@/components/product/product-page";
import UserPage from "@/components/user/user-page";
import { useContext } from "react";
import { AppContext } from "@/components/canteen-provider";
import { navigationMenuTriggerStyle } from "./components/ui/navigation-menu";
import { ModeToggle } from "@/components/mode-toggle";

function App() {
  const { admin, currentPage, setCurrentPage } = useContext(AppContext);

  return (
    <div className="min-h-screen ">
      <nav className="shadow-sm">
        <div className="container mx-auto p-4">
          <div className="flex">
            <div className="flex-shrink-0 flex items-center">
              <h1 className="text-xl font-bold">Maya Canteen</h1>
            </div>
            {admin && (
              <div className="ml-6 flex space-x-8">
                <button
                  onClick={() => setCurrentPage("canteen")}
                  className={navigationMenuTriggerStyle()}
                >
                  Canteen
                </button>
                <button
                  onClick={() => setCurrentPage("products")}
                  className={navigationMenuTriggerStyle()}
                >
                  Products
                </button>
                <button
                  onClick={() => setCurrentPage("users")}
                  className={navigationMenuTriggerStyle()}
                >
                  Users
                </button>
              </div>
            )}
            <div className="ml-auto flex items-center">
              <ModeToggle />
            </div>
          </div>
        </div>
      </nav>

      <main>
        {currentPage === "canteen" && <CanteenPage />}
        {currentPage === "products" && <ProductPage />}
        {currentPage === "users" && <UserPage />}
      </main>
    </div>
  );
}

export default App;
