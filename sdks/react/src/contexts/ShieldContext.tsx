import React, { createContext, useEffect, useState } from "react";
import Shield from "../shield";
import { ShieldClientOptions, ShieldProviderProps, User } from "../types";

import type { Strategy } from "./StrategyContext";
import { StrategyContext } from "./StrategyContext";
import { UserContext } from "./UserContext";

type ShieldContextProviderProps = {};
const initialValues: ShieldContextProviderProps = {};
export const ShieldContext =
  createContext<ShieldContextProviderProps>(initialValues);
ShieldContext.displayName = "ShieldContext ";

export const ShieldContextProvider = (props: ShieldProviderProps) => {
  const { children, initialState, ...options } = props;
  const { shieldClient } = useShieldClient(options);
  const [strategies, setStrategies] = useState<Strategy[]>([]);
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    async function getShieldInformation() {
      const {
        data: { strategies },
      } = await shieldClient.getAuthAtrategies();
      setStrategies(strategies);
    }
    getShieldInformation();
  }, []);

  useEffect(() => {
    async function getShieldCurrentUser() {
      const {
        data: { user },
      } = await shieldClient.getCurrentUser();
      setUser(user);
    }
    getShieldCurrentUser();
  }, []);

  return (
    <ShieldContext.Provider value={{ client: shieldClient }}>
      <StrategyContext.Provider value={{ strategies }}>
        <UserContext.Provider value={{ user }}>{children}</UserContext.Provider>
      </StrategyContext.Provider>
    </ShieldContext.Provider>
  );
};

const useShieldClient = (options: ShieldClientOptions) => {
  const shieldClient = React.useMemo(
    () => Shield.getOrCreateInstance(options),
    []
  );

  return { shieldClient };
};
