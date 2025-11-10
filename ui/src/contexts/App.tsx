import * as R from "ramda";
import React, {
  PropsWithChildren,
  createContext,
} from "react";
import { useQuery as useConnectQuery } from "@connectrpc/connect-query";
import { useQuery } from "@tanstack/react-query";
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
  } = useConnectQuery(FrontierServiceQueries.getCurrentUser, {});

  const user = currentUserResponse?.user;

  const { data: config = defaultConfig } = useQuery({
    queryKey: ["config"],
    queryFn: async () => {
      const resp = await fetch("/configs");
      return (await resp.json()) as Config;
    },
  });

  const { error: adminUserError } = useConnectQuery(
    AdminServiceQueries.getCurrentAdminUser,
    {},
  );

  const isAdmin = Boolean(user?.id) && !adminUserError;

  return (
    <AppContext.Provider
      value={{
        isLoading: isUserLoading,
        isAdmin,
        config,
        user,
      }}>
      {children}
    </AppContext.Provider>
  );
};
