import { useContext, useEffect, useState } from "react";
import { toast } from "sonner";

import { ProductTable } from "@/components/product/product-data-table";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { AppContext } from "@/context";
import { Product, transactionService } from "@/services/transaction-service";
import { LoaderCircle } from "lucide-react";

interface ProductListProps {
  refreshTrigger?: number;
  onEditProduct: (product: Product) => void;
  onDeleteProduct: (product: Product) => void;
}

export default function ProductList({
  refreshTrigger = 0,
  onEditProduct,
  onDeleteProduct,
}: ProductListProps) {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);

  const { admin } = useContext(AppContext);

  useEffect(() => {
    const fetchProducts = async () => {
      setLoading(true);
      try {
        const productsData = await transactionService.getAllProducts();
        setProducts(productsData ?? []);
      } catch (error) {
        toast.error("Failed to load products");
        console.error(error);
      } finally {
        setLoading(false);
      }
    };

    fetchProducts();
  }, [refreshTrigger]);

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>Products</CardTitle>
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className="flex items-center justify-center h-32">
            <LoaderCircle className="w-8 h-8 text-primary animate-spin" />
          </div>
        ) : (
          <ProductTable
            data={products}
            admin={admin}
            onEdit={onEditProduct}
            onDelete={onDeleteProduct}
          />
        )}
      </CardContent>
    </Card>
  );
}
