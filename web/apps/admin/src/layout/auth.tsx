import { FrontierProvider } from "@raystack/frontier/client";
import { Outlet } from "react-router-dom";
import { frontierConfig } from "~/configs/frontier";
import { themeConfig } from "~/configs/theme";

// TODO: remove frontier client dependency from auth pages like login, signup etc in SDK
export default function AuthLayout() {
  return (
    <FrontierProvider config={frontierConfig} theme={themeConfig} renderThemeProvider={false} renderToastProvider={false}>
      <Outlet />
    </FrontierProvider>
  );
}
