import { createRoot } from "react-dom/client";
import App from "./App.tsx";
import "./index.css";

import { ThemeProvider } from "@/components/theme-provider";
import { AppProvider } from "@/context";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
// import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { Toaster } from "sonner";

// Create a client
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000, // 5 minutes
      refetchOnWindowFocus: false,
    },
  },
});

createRoot(document.getElementById("root")!).render(
  <QueryClientProvider client={queryClient}>
    <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
      <AppProvider>
        <App />
        <Toaster richColors />
      </AppProvider>
      {/* <ReactQueryDevtools initialIsOpen={false} /> */}
    </ThemeProvider>
  </QueryClientProvider>
);
