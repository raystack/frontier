import { ThemeProvider, Toast } from "@raystack/apsara";
import { SkeletonTheme } from "react-loading-skeleton";
import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter } from "react-router-dom";

import Routes from "./routes";
import { AppContextProvider } from "./contexts/App";
import { ConnectProvider } from "./contexts/ConnectProvider";
import { themeConfig } from "~/configs/theme";
import { ErrorBoundary } from "./components/error-boundary";

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <ErrorBoundary>
      <BrowserRouter>
        <ThemeProvider {...themeConfig}>
          <SkeletonTheme
            highlightColor="var(--rs-color-background-base-primary)"
            baseColor="var(--rs-color-background-base-primary-hover)"
          >
            <ConnectProvider>
              <Toast.Provider>
                <AppContextProvider>
                  <Routes />
                </AppContextProvider>
              </Toast.Provider>
            </ConnectProvider>
          </SkeletonTheme>
        </ThemeProvider>
      </BrowserRouter>
    </ErrorBoundary>
  </React.StrictMode>,
);
