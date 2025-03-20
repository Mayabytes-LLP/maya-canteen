import ProductForm from "@/components/product/product-form";
import ProductList from "@/components/product/product-list";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Product, transactionService } from "@/services/transaction-service";
import { useState } from "react";
import { toast } from "sonner";

export default function ProductPage() {
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  const [editProduct, setEditProduct] = useState<Product | null>(null);
  const [deleteProduct, setDeleteProduct] = useState<Product | null>(null);
  const [openEditDialog, setOpenEditDialog] = useState(false);
  const [openDeleteDialog, setOpenDeleteDialog] = useState(false);

  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleProductEvent = () => {
    setRefreshTrigger((prev) => prev + 1);
  };

  const handleEditProduct = (product: Product) => {
    setEditProduct(product);
    setOpenEditDialog(true);
  };

  const handleDeleteProduct = async (product: Product) => {
    setDeleteProduct(product);
    setOpenDeleteDialog(true);
  };

  const submitDeleteProduct = async (product: Product) => {
    if (!product) return;
    try {
      await transactionService.deleteProduct(product.id);
      toast.success("Product deleted successfully");
      handleProductEvent();
    } catch (error) {
      console.error("Error saving product:", error);
      toast.error("Failed to save product");
    } finally {
      setIsSubmitting(false);
      setOpenDeleteDialog(false);
      setDeleteProduct(null);
    }
  };

  return (
    <div className="container mx-auto py-6 space-y-6">
      <h1 className="text-3xl font-bold">Product Management</h1>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div>
          <ProductForm onProductAdded={handleProductEvent} />
        </div>
        <div className="md:col-span-2">
          <ProductList
            refreshTrigger={refreshTrigger}
            onEditProduct={handleEditProduct}
            onDeleteProduct={handleDeleteProduct}
          />
        </div>
      </div>
      {editProduct && (
        <Dialog open={openEditDialog} onOpenChange={setOpenEditDialog}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Edit Product</DialogTitle>
              <DialogDescription>
                Update the product information below.
              </DialogDescription>
            </DialogHeader>
            <ProductForm
              onProductAdded={() => {
                handleProductEvent();
                setOpenEditDialog(false);
              }}
              initialData={editProduct}
            />
            <DialogFooter>
              <DialogClose asChild>
                <Button variant="outline">Cancel</Button>
              </DialogClose>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}

      {deleteProduct && (
        <Dialog open={openDeleteDialog} onOpenChange={setOpenDeleteDialog}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Confirm Deletion</DialogTitle>
              <DialogDescription>
                Are you sure you want to delete this user? This action cannot be
                undone. {deleteProduct.name}
              </DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <DialogClose asChild>
                <Button variant="outline">Cancel</Button>
              </DialogClose>
              <Button
                variant="destructive"
                onClick={() => submitDeleteProduct(deleteProduct)}
                disabled={isSubmitting}
              >
                {isSubmitting ? "Deleting..." : "Delete Product"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </div>
  );
}
