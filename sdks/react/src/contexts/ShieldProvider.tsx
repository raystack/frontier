import React from "react";
import { ShieldProviderProps } from "../types";
import { ShieldContextProvider } from "./ShieldContext";
import { withMaxAllowedInstancesGuard } from "./useMaxAllowedInstancesGuard";

export const multipleShieldProvidersError =
  "Clerk: You've added multiple <ClerkProvider> components in your React component tree. Wrap your components in a single <ClerkProvider>.";

const ShieldProviderBase = (props: ShieldProviderProps) => {
  const { children, initialState, ...options } = props;
  return (
    <ShieldContextProvider initialState={initialState} {...options}>
      {children}
    </ShieldContextProvider>
  );
};

export const ShieldProvider = withMaxAllowedInstancesGuard(
  ShieldProviderBase,
  "ShieldProvider",
  multipleShieldProvidersError
);
ShieldProvider.displayName = "ShieldProvider";
