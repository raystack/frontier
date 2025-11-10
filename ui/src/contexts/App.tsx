import * as R from "ramda";
import React, {
  PropsWithChildren,
  createContext,
  useCallback,
  useEffect,
  useState,
} from "react";
import { useQuery } from "@connectrpc/connect-query";
import {
  AdminServiceQueries,
  FrontierServiceQueries,
  type User,
} from "@raystack/proton/frontier";
import { Config, defaultConfig } from "~/utils/constants";

interface AppContextValue {
  isAdmin: boolean;
  isLoading: boolean;
  config: Config;
  user?: User;
}

const AppContextDefaultValue = {
  isAdmin: false,
  isLoading: false,
  config: defaultConfig,
};

export const AppContext = createContext<AppContextValue>(
  AppContextDefaultValue,
);

export const AppContextProvider: React.FC<PropsWithChildren> = function ({
  children,
}) {
  const {
    data: currentUserResponse,
    isLoading: isUserLoading,
  } = useQuery(FrontierServiceQueries.getCurrentUser, {});

  const user = currentUserResponse?.user;

  const [config, setConfig] = useState<Config>(defaultConfig);

  const { error: adminUserError } = useQuery(
    AdminServiceQueries.getCurrentAdminUser,
    {},
  );

  const isAdmin = Boolean(user?.id) && !adminUserError;

  const fetchConfig = useCallback(async () => {
    try {
      const resp = await fetch("/configs");
      const data = (await resp?.json()) as Config;
      setConfig(data);
    } catch (err) {
      console.error(err);
    }
  }, []);

  useEffect(() => {
    fetchConfig();
  }, [fetchConfig]);

  const isLoading = isUserLoading;

  return (
    <AppContext.Provider
      value={{
        isLoading,
        isAdmin,
        config,
        user,
      }}>
      {children}
    </AppContext.Provider>
  );
};
