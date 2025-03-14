import CanteenPage from "@/components/canteen-page";
import { AppContext } from "@/components/canteen-provider";
import { ModeToggle } from "@/components/mode-toggle";
import ProductPage from "@/components/product/product-page";
import UserPage from "@/components/user/user-page";
import { useContext, useEffect, useState } from "react";
import { navigationMenuTriggerStyle } from "./components/ui/navigation-menu";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "./components/ui/form";
import { Input } from "./components/ui/input";
import { transactionService } from "./services/transaction-service";

// Form validation schema
const formSchema = z.object({
  employee_id: z.string().min(2, "Employee ID must be at least 2 characters"),
});

type FormValues = z.infer<typeof formSchema>;

function App() {
  const { admin, currentPage, currentUser, setCurrentPage, setCurrentUser } =
    useContext(AppContext);
  const [showLogin, setShowLogin] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Initialize form
  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      employee_id: "",
    },
  });

  const onSubmit = async (data: FormValues) => {
    setIsSubmitting(true);
    try {
      setShowLogin(true);
      console.log(data);
      // number need to be 5 digits
      const user = await transactionService.getUser(data.employee_id);
      if (!user) {
        toast.error("User not found");
        form.reset();
        return;
      }

      setCurrentUser(user);
      toast.success("User logged in successfully");
      form.reset();
    } catch (error) {
      console.error("Error adding user:", error);
      toast.error("Failed to add user");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleInteraction = () => {
    setShowLogin(true);
  };
  return (
    <div className="min-h-screen ">
      {currentPage != "screenSaver" && (
        <nav className="shadow-sm">
          <div className="container mx-auto p-4">
            <div className="flex">
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
      )}
      <main>
        {currentPage === "canteen" && <CanteenPage />}
        {currentPage === "products" && <ProductPage />}
        {currentPage === "users" && <UserPage />}

        {currentPage === "screenSaver" && !showLogin && (
          <div
            className="flex h-screen w-full overflow-hidden relative items-center justify-center bg-background"
            onClick={handleInteraction}
            onTouchStart={handleInteraction}
            onMouseMove={handleInteraction}
          >
            <Screensaver />
          </div>
        )}
        {showLogin && !currentUser?.id && (
          <div className="flex h-screen w-full overflow items-center justify-center bg-background">
            <Card className="w-[350px]">
              <CardHeader>
                <CardTitle className="text-center">
                  Mayabytes Canteen Login
                </CardTitle>
              </CardHeader>
              <CardContent>
                <Form {...form}>
                  <form
                    onSubmit={form.handleSubmit(onSubmit)}
                    className="space-y-4"
                  >
                    <FormField
                      control={form.control}
                      name="employee_id"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Employee ID</FormLabel>
                          <FormControl>
                            <Input placeholder="Enter employee ID" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />

                    <Button
                      className="w-full"
                      type="submit"
                      disabled={isSubmitting}
                    >
                      {isSubmitting ? "Validating..." : "Login"}
                    </Button>
                  </form>
                </Form>
              </CardContent>

              <div className="flex justify-center p-4">
                <Button
                  variant="outline"
                  className="text-sm w-full"
                  onClick={() => setShowLogin(false)}
                  type="button"
                >
                  back to screensaver
                </Button>
              </div>
            </Card>
          </div>
        )}
      </main>
    </div>
  );
}

function Screensaver() {
  const [position, setPosition] = useState({ x: 50, y: 50 });
  const [direction, setDirection] = useState({ x: 1, y: 1 });

  // ASCII art for canteen theme
  const asciiArt = [
    "  __  __                _           _            ",
    " |  \\/  | __ _ _   _  | |__  _   _| |_ ___  ___ ",
    " | |\\/| |/ _` | | | | | '_ \\| | | | __/ _ \\/ __|",
    " | |  | | (_| | |_| | | |_) | |_| | ||  __/\\__ \\",
    " |_|  |_|\\__,_|\\__, | |_.__/ \\__, |\\__\\___||___/",
    "               |___/         |___/              ",
    "",
    "  ____            _                    ",
    " / ___|__ _ _ __ | |_ ___  ___ _ __   ",
    "| |   / _` | '_ \\| __/ _ \\/ _ \\ '_ \\  ",
    "| |__| (_| | | | | ||  __/  __/ | | | ",
    " \\____\\__,_|_| |_|\\__\\___|\\___|_| |_| ",
  ];

  // Food-related ASCII art
  const foodArt = [
    "   (    )",
    "  (    )",
    " (    )",
    " |    |",
    " |    |",
    " |____|",
    "  \\  /",
    "   \\/",
    "",
    "  _____",
    " /     \\",
    "|  o o  |",
    "|   ᴥ   |",
    " \\_____/",
    "",
    "   ____",
    "  /    \\",
    " | ><   |",
    "  \\____/",
  ];

  // Handle animation
  useEffect(() => {
    const interval = setInterval(() => {
      setPosition((prev) => {
        // Calculate new position
        const newX = prev.x + direction.x * 0.5;
        const newY = prev.y + direction.y * 0.5;

        // Check boundaries and reverse direction if needed
        let newDirX = direction.x;
        let newDirY = direction.y;

        if (newX <= 0 || newX >= 90) {
          newDirX = -direction.x;
        }

        if (newY <= 0 || newY >= 90) {
          newDirY = -direction.y;
        }

        // Update direction if needed
        if (newDirX !== direction.x || newDirY !== direction.y) {
          setDirection({ x: newDirX, y: newDirY });
        }

        return { x: newX, y: newY };
      });
    }, 50);

    return () => clearInterval(interval);
  }, [direction]);

  // Handle user interaction

  return (
    <div className="flex h-full w-full cursor-none flex-col items-center justify-center overflow-hidden bg-black text-primary">
      <div
        className="absolute transition-all duration-500 ease-linear"
        style={{
          left: `${position.x}%`,
          top: `${position.y}%`,
          transform: "translate(-50%, -50%)",
        }}
      >
        <pre className="animate-pulse text-xs sm:text-sm md:text-base">
          {asciiArt.join("\n")}
        </pre>
      </div>

      <div className="absolute left-10 top-10 animate-pulse">
        <pre className="text-xs text-orange-400 sm:text-sm">
          {foodArt.slice(8, 14).join("\n")}
        </pre>
      </div>

      <div className="absolute bottom-10 right-10 animate-bounce">
        <pre className="text-xs text-blue-300 sm:text-sm">
          {foodArt.slice(14).join("\n")}
        </pre>
      </div>

      <div className="absolute bottom-5 w-full text-center">
        <p className="animate-pulse text-lg text-primary-foreground">
          Touch screen to login
        </p>
      </div>
    </div>
  );
}
export default App;
