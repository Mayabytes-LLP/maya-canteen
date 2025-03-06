import ProductForm from "@/components/product/product-form";
import ProductList from "@/components/product/product-list";
import { useState } from "react";

export default function ProductPage() {
  const [refreshTrigger, setRefreshTrigger] = useState(0);

  const handleProductAdded = () => {
    setRefreshTrigger((prev) => prev + 1);
  };

  return (
    <div className="container mx-auto py-6 space-y-6">
      <h1 className="text-3xl font-bold">Product Management</h1>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div>
          <ProductForm onProductAdded={handleProductAdded} />
        </div>
        <div>
          <ProductList refreshTrigger={refreshTrigger} />
        </div>
      </div>
    </div>
  );
}
