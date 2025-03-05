import { ThemeProvider } from "@raystack/apsara/v1";
import { FrontierProvider } from "@raystack/frontier/react";
import { SkeletonTheme } from "react-loading-skeleton";
import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import { Toaster } from "sonner";
import Routes from "./routes";
import { AppContextProvider } from "./contexts/App";

const getFrontierConfig = () => {
  const frontierEndpoint =
    process.env.NEXT_PUBLIC_FRONTIER_URL || "/frontier-api";
  const currentHost = window?.location?.origin || "http://localhost:3000";
  return {
    endpoint: frontierEndpoint,
    redirectLogin: `${currentHost}/login`,
    redirectSignup: `${currentHost}/signup`,
    redirectMagicLinkVerify: `${currentHost}/magiclink-verify`,
    callbackUrl: `${currentHost}/callback`,
  };
};

export const frontierConfig = getFrontierConfig();

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <BrowserRouter>
      <ThemeProvider defaultTheme="light">
        <SkeletonTheme
          highlightColor="var(--background-base)"
          baseColor="var(--background-base-hover)"
        >
          <FrontierProvider config={frontierConfig}>
            <AppContextProvider>
              <Routes />
            </AppContextProvider>
          </FrontierProvider>
          <Toaster richColors />
        </SkeletonTheme>
      </ThemeProvider>
    </BrowserRouter>
  </React.StrictMode>
);
