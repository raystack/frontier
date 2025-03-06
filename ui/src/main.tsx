import { ThemeProvider } from "@raystack/apsara/v1";
import { SkeletonTheme } from "react-loading-skeleton";
import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import { Toaster } from "sonner";
import Routes from "./routes";
import { AppContextProvider } from "./contexts/App";
import { themeConfig } from "~/configs/theme";

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <BrowserRouter>
      <ThemeProvider {...themeConfig}>
        <SkeletonTheme
          highlightColor="var(--background-base)"
          baseColor="var(--background-base-hover)"
        >
          <AppContextProvider>
            <Routes />
          </AppContextProvider>
          <Toaster richColors />
        </SkeletonTheme>
      </ThemeProvider>
    </BrowserRouter>
  </React.StrictMode>
);
