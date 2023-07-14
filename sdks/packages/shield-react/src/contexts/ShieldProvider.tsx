import { ShieldProviderProps } from "@raystack/shield-js";
import React from "react";

import { ShieldContextProvider } from "./ShieldContext";
import { withMaxAllowedInstancesGuard } from "./useMaxAllowedInstancesGuard";

export const multipleShieldProvidersError =
  "Shield: You've added multiple <ShieldProvider> components in your React component tree. Wrap your components in a single <ShieldProvider>.";

const ShieldProviderBase = (props: ShieldProviderProps) => {
  const { children, initialState, ...options } = props;
  return (
    <ShieldContextProvider initialState={initialState} {...options}>
      {children}
    </ShieldContextProvider>
  );
};

export const ShieldProvider = withMaxAllowedInstancesGuard<ShieldProviderProps>(
  ShieldProviderBase,
  "ShieldProvider",
  multipleShieldProvidersError
);
ShieldProvider.displayName = "ShieldProvider";
