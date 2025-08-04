import { zodResolver } from "@hookform/resolvers/zod";
import { Menu } from "lucide-react";
import { useContext, useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { Navigate, NavLink, Route, Routes, useNavigate } from "react-router";
import { toast } from "sonner";
import { z } from "zod";
import { ModeToggle } from "@/components/mode-toggle";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
	Form,
	FormControl,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import {
	NavigationMenu,
	NavigationMenuItem,
	NavigationMenuList,
	navigationMenuTriggerStyle,
} from "@/components/ui/navigation-menu";
import {
	Sheet,
	SheetClose,
	SheetContent,
	SheetHeader,
	SheetTitle,
	SheetTrigger,
} from "@/components/ui/sheet";
import { AppContext } from "@/context";
import { cn } from "@/lib/utils";
import CanteenPage from "@/pages/canteen-page";
import DashboardPage from "@/pages/dashboard-page";
import ProductPage from "@/pages/product-page";
import ProductSalesPage from "@/pages/product-sales-page";
import TransactionsPage from "@/pages/transactions-page";
import UserPage from "@/pages/user-page";
import { transactionService } from "@/services/transaction-service";

// Form validation schema
const formSchema = z.object({
	employee_id: z.string().min(2, "Employee ID must be at least 2 characters"),
});

type FormValues = z.infer<typeof formSchema>;

const AuthGuard = ({ children }: { children: React.ReactElement }) => {
	const { currentUser } = useContext(AppContext);

	if (!currentUser?.id) {
		return <Navigate to="/login" replace />;
	}
	return children;
};

function App() {
	const { admin, currentUser, setCurrentUser, zkDeviceStatus } =
		useContext(AppContext);
	const [showLogin, setShowLogin] = useState(false);
	const [isSubmitting, setIsSubmitting] = useState(false);
	const navigate = useNavigate();

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
			const password = prompt("Enter password to login");
			if (password !== "6479") {
				toast.error("Invalid password");
				form.reset();
				return;
			}
			const user = await transactionService.getUser(data.employee_id);
			if (!user) {
				toast.error("User not found");
				form.reset();
				return;
			}

			setCurrentUser(user);
			toast.success("User logged in successfully");
			form.reset();
			navigate("/canteen");
		} catch (error) {
			console.error("Error getting User:", error);
			toast.error("Failed to login");
		} finally {
			setIsSubmitting(false);
		}
	};

	const handleInteraction = () => {
		setShowLogin(true);
	};

	return (
		<div className="min-h-screen ">
			<div
				className={cn(
					"fixed bottom-0 left-0 w-full h-1 z-50",
					zkDeviceStatus ? "bg-green-500" : "bg-pink-700",
				)}
			></div>
			{currentUser && (
				<nav className="shadow-sm border-b">
					<div className="container mx-auto p-4">
						<div className="flex items-center justify-between">
							<div className="flex items-center gap-2">
								{/* Mobile Navigation */}
								{admin && (
									<div className="lg:hidden">
										<Sheet>
											<SheetTrigger asChild>
												<Button
													variant="outline"
													size="icon"
													className="h-9 w-9 p-0"
												>
													<Menu className="h-5 w-5" />
													<span className="sr-only">Toggle menu</span>
												</Button>
											</SheetTrigger>
											<SheetContent side="left">
												<SheetHeader className="pb-4">
													<SheetTitle>Maya Canteen</SheetTitle>
												</SheetHeader>
												<div className="grid gap-2 py-4">
													<NavigationLink
														to="/dashboard"
														className={({ isActive }) =>
															cn(
																"flex items-center py-2 px-3 rounded-md hover:bg-accent",
																isActive && "bg-accent text-accent-foreground",
															)
														}
													>
														Dashboard
													</NavigationLink>
													<NavigationLink
														to="/product-sales"
														className={({ isActive }) =>
															cn(
																"flex items-center py-2 px-3 rounded-md hover:bg-accent",
																isActive && "bg-accent text-accent-foreground",
															)
														}
													>
														Product Sales
													</NavigationLink>
													<NavigationLink
														to="/canteen"
														className={({ isActive }) =>
															cn(
																"flex items-center py-2 px-3 rounded-md hover:bg-accent",
																isActive && "bg-accent text-accent-foreground",
															)
														}
													>
														Canteen
													</NavigationLink>
													<NavigationLink
														to="/products"
														className={({ isActive }) =>
															cn(
																"flex items-center py-2 px-3 rounded-md hover:bg-accent",
																isActive && "bg-accent text-accent-foreground",
															)
														}
													>
														Products
													</NavigationLink>
													<NavigationLink
														to="/users"
														className={({ isActive }) =>
															cn(
																"flex items-center py-2 px-3 rounded-md hover:bg-accent",
																isActive && "bg-accent text-accent-foreground",
															)
														}
													>
														Users
													</NavigationLink>
													<NavigationLink
														to="/transactions"
														className={({ isActive }) =>
															cn(
																"flex items-center py-2 px-3 rounded-md hover:bg-accent",
																isActive && "bg-accent text-accent-foreground",
															)
														}
													>
														Transactions
													</NavigationLink>
												</div>
											</SheetContent>
										</Sheet>
									</div>
								)}
								<h2 className="text-lg font-bold">Maya Canteen</h2>

								{/* Desktop Navigation */}
								{admin && (
									<div className="hidden lg:flex ml-6">
										<NavigationMenu>
											<NavigationMenuList>
												<NavigationMenuItem>
													<NavLink
														to="/dashboard"
														className={({ isActive }) =>
															cn(
																navigationMenuTriggerStyle(),
																isActive && "bg-accent text-accent-foreground",
															)
														}
													>
														Dashboard
													</NavLink>
												</NavigationMenuItem>
												<NavigationMenuItem>
													<NavLink
														to="/product-sales"
														className={({ isActive }) =>
															cn(
																navigationMenuTriggerStyle(),
																isActive && "bg-accent text-accent-foreground",
															)
														}
													>
														Product Sales
													</NavLink>
												</NavigationMenuItem>
												<NavigationMenuItem>
													<NavLink
														to="/canteen"
														className={({ isActive }) =>
															cn(
																navigationMenuTriggerStyle(),
																isActive && "bg-accent text-accent-foreground",
															)
														}
													>
														Canteen
													</NavLink>
												</NavigationMenuItem>
												<NavigationMenuItem>
													<NavLink
														to="/products"
														className={({ isActive }) =>
															cn(
																navigationMenuTriggerStyle(),
																isActive && "bg-accent text-accent-foreground",
															)
														}
													>
														Products
													</NavLink>
												</NavigationMenuItem>
												<NavigationMenuItem>
													<NavLink
														to="/users"
														className={({ isActive }) =>
															cn(
																navigationMenuTriggerStyle(),
																isActive && "bg-accent text-accent-foreground",
															)
														}
													>
														Users
													</NavLink>
												</NavigationMenuItem>
												<NavigationMenuItem>
													<NavLink
														to="/transactions"
														className={({ isActive }) =>
															cn(
																navigationMenuTriggerStyle(),
																isActive && "bg-accent text-accent-foreground",
															)
														}
													>
														Transactions
													</NavLink>
												</NavigationMenuItem>
											</NavigationMenuList>
										</NavigationMenu>
									</div>
								)}
							</div>

							<div className="flex items-center gap-2">
								<Button
									variant="destructive"
									onClick={() => {
										setCurrentUser(null);
										navigate("/login");
									}}
								>
									Logout {currentUser?.name || "Guest"}
								</Button>
								<ModeToggle />
							</div>
						</div>
					</div>
				</nav>
			)}
			<main>
				<Routes>
					<Route
						path="/canteen"
						element={
							<AuthGuard>
								<CanteenPage />
							</AuthGuard>
						}
					/>
					<Route
						path="/products"
						element={
							<AuthGuard>
								<ProductPage />
							</AuthGuard>
						}
					/>
					<Route
						path="/users"
						element={
							<AuthGuard>
								<UserPage />
							</AuthGuard>
						}
					/>
					<Route
						path="/transactions"
						element={
							<AuthGuard>
								<TransactionsPage />
							</AuthGuard>
						}
					/>
					<Route
						path="/dashboard"
						element={
							<AuthGuard>
								<DashboardPage />
							</AuthGuard>
						}
					/>
					<Route
						path="/product-sales"
						element={
							<AuthGuard>
								<ProductSalesPage />
							</AuthGuard>
						}
					/>
					<Route
						path="/login"
						element={
							showLogin && !currentUser?.id ? (
								<div className="flex h-screen w-full overflow items-center justify-center bg-background">
									<Card className="w-[350px]">
										<CardHeader>
											<CardTitle className="text-center">Admin Login</CardTitle>
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
																	<Input
																		placeholder="Enter employee ID"
																		{...field}
																	/>
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
							) : (
								<button
									type="button"
									className="flex h-screen w-full overflow-hidden relative items-center justify-center bg-background"
									onClick={handleInteraction}
									onTouchStart={handleInteraction}
									onMouseMove={handleInteraction}
									onKeyDown={(e) => e.key === "Enter" && handleInteraction()}
									onKeyUp={(e) => e.key === "Enter" && handleInteraction()}
								>
									<Screensaver />
								</button>
							)
						}
					/>
					<Route path="*" element={<Navigate to="/screensaver" />} />
				</Routes>
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
		"|         ᴥ  |",
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

const NavigationLink = ({
	to,
	children,
	className,
}: {
	to: string;
	children: React.ReactNode;
	className?: (props: { isActive: boolean }) => string;
}) => {
	return (
		<SheetClose asChild>
			<NavLink
				to={to}
				className={cn("block py-2 px-4 rounded-md hover:bg-accent", className)}
			>
				{children}
			</NavLink>
		</SheetClose>
	);
};
