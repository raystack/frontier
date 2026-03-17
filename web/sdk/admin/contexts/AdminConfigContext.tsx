import React, { createContext, ReactNode, useContext } from "react";
import { merge } from "lodash";
import { Config, defaultConfig, defaultTerminology } from "../utils/constants";
import { TerminologyProvider } from "../../shared/terminology";

const AdminConfigContext = createContext<Config>(defaultConfig);

export interface AdminConfigProviderProps {
  children: ReactNode;
  config?: Config;
}

export const AdminConfigProvider: React.FC<AdminConfigProviderProps> = ({
  children,
  config = {},
}) => {
  const mergedConfig: Config = merge({}, defaultConfig, config);

  // Ensure terminology is always present with defaults
  mergedConfig.terminology = merge({}, defaultTerminology, config.terminology);

  return (
    <AdminConfigContext.Provider value={mergedConfig}>
      <TerminologyProvider
        terminology={mergedConfig.terminology}
        defaults={defaultTerminology}
      >
        {children}
      </TerminologyProvider>
    </AdminConfigContext.Provider>
  );
};

export const useAdminConfig = () => {
  const context = useContext(AdminConfigContext);
  return context || defaultConfig;
};

export { AdminConfigContext };
