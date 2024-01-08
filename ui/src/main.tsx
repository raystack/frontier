import { ThemeProvider } from "@raystack/apsara";
import { FrontierProvider } from "@raystack/frontier/react";
import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import Routes from "./routes";

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <BrowserRouter>
      <ThemeProvider defaultTheme="light">
        <FrontierProvider
          config={{
            endpoint: import.meta.env.NEXT_PUBLIC_SHIELD_URL,
            redirectLogin: import.meta.env.NEXT_PUBLIC_WEBSITE_URL,
            redirectSignup: `${import.meta.env.NEXT_PUBLIC_WEBSITE_URL}/signup`,
            redirectMagicLinkVerify: `${
              import.meta.env.NEXT_PUBLIC_WEBSITE_URL
            }/magiclink-verify`,
            callbackUrl: import.meta.env.NEXT_PUBLIC_CALLBACK_URL,
          }}
        >
          <Routes />
        </FrontierProvider>
      </ThemeProvider>
    </BrowserRouter>
  </React.StrictMode>
);
