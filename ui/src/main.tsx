import { ThemeProvider } from "@raystack/apsara";
import { FrontierProvider } from "@raystack/frontier/react";
import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import { Toaster } from "sonner";
import Routes from "./routes";

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <BrowserRouter>
      <ThemeProvider defaultTheme="light">
        <FrontierProvider
          config={{
            endpoint: process.env.NEXT_PUBLIC_SHIELD_URL,
            redirectLogin: process.env.NEXT_PUBLIC_WEBSITE_URL,
            redirectMagicLinkVerify: `${process.env.NEXT_PUBLIC_WEBSITE_URL}/console/magiclink-verify`,
            callbackUrl: process.env.NEXT_PUBLIC_CALLBACK_URL,
          }}
        >
          <Routes />
        </FrontierProvider>
        <Toaster richColors />
      </ThemeProvider>
    </BrowserRouter>
  </React.StrictMode>
);
