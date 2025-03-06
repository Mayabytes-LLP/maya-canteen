import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import App from "./App.tsx";
import "./index.css";

import { AppProvider } from "@/components/canteen-provider.tsx";
import { ThemeProvider } from "@/components/theme-provider";
import { Toaster } from "sonner";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
      <AppProvider>
        <App />
        <Toaster />
      </AppProvider>
    </ThemeProvider>
  </StrictMode>,
);
