import { ThemeProvider, ToastContainer } from "@raystack/apsara";
import { SkeletonTheme } from "react-loading-skeleton";
import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter } from "react-router-dom";

import Routes from "./routes";
import { AppContextProvider } from "./contexts/App";
import { ConnectProvider } from "./contexts/ConnectProvider";
import { themeConfig } from "~/configs/theme";

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <BrowserRouter>
      <ThemeProvider {...themeConfig}>
        <SkeletonTheme
          highlightColor="var(--rs-color-background-base-primary)"
          baseColor="var(--rs-color-background-base-primary-hover)"
        >
          <ConnectProvider>
            <AppContextProvider>
              <Routes />
            </AppContextProvider>
            <ToastContainer richColors />
          </ConnectProvider>
        </SkeletonTheme>
      </ThemeProvider>
    </BrowserRouter>
  </React.StrictMode>,
);
