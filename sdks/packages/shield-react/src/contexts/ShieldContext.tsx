import {
  Group,
  Organization,
  ShieldClientOptions,
  ShieldProviderProps,
  Strategy,
  User,
} from "@raystack/shield-js";
import React, {
  createContext,
  Dispatch,
  SetStateAction,
  useContext,
  useEffect,
  useState,
} from "react";

import Shield from "../shield";
interface ShieldContextProviderProps {
  client: Shield;

  organizations: Organization[];
  setOrganizations: Dispatch<SetStateAction<Organization[]>>;

  groups: Group[];
  setGroups: Dispatch<SetStateAction<Group[]>>;

  strategies: Strategy[];
  setStrategies: Dispatch<SetStateAction<Strategy[]>>;

  user: User | null;
  setUser: Dispatch<SetStateAction<User | null>>;
}

const initialValues: ShieldContextProviderProps = {
  client: Shield.getOrCreateInstance({ endpoint: "http://localhost:8080" }),

  organizations: [],
  setOrganizations: () => undefined,

  groups: [],
  setGroups: () => undefined,

  strategies: [],
  setStrategies: () => undefined,

  user: null,
  setUser: () => undefined,
};

export const ShieldContext =
  createContext<ShieldContextProviderProps>(initialValues);
ShieldContext.displayName = "ShieldContext ";

export const ShieldContextProvider = (props: ShieldProviderProps) => {
  const { children, initialState, ...options } = props;
  const { shieldClient } = useShieldClient(options);

  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [groups, setGroups] = useState<Group[]>([]);
  const [strategies, setStrategies] = useState<Strategy[]>([]);
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    async function getShieldInformation() {
      try {
        const {
          data: { strategies },
        } = await shieldClient.getAuthAtrategies();
        setStrategies(strategies);
      } catch (error) {
        console.error(
          "shield:sdk:: There is problem with fetching auth strategies"
        );
      }
    }
    getShieldInformation();
  }, []);

  useEffect(() => {
    async function getShieldCurrentUser() {
      try {
        const {
          data: { user },
        } = await shieldClient.getCurrentUser();
        setUser(user);
      } catch (error) {
        console.error(
          "shield:sdk:: There is problem with fetching current user information"
        );
      }
    }
    getShieldCurrentUser();
  }, []);

  useEffect(() => {
    async function getShieldCurrentUserGroups(userId: string) {
      try {
        const {
          data: { groups },
        } = await shieldClient.getUserGroups(userId);
        setGroups(groups);
      } catch (error) {
        console.error(
          "shield:sdk:: There is problem with fetching user groups information"
        );
      }
    }

    if (user) {
      getShieldCurrentUserGroups(user.id);
    }
  }, [user]);

  useEffect(() => {
    async function getShieldCurrentUserOrganizations(userId: string) {
      try {
        const {
          data: { organizations },
        } = await shieldClient.getUserOrganisations(userId);
        setOrganizations(organizations);
      } catch (error) {
        console.error(
          "shield:sdk:: There is problem with fetching user current organizations"
        );
      }
    }

    if (user) {
      getShieldCurrentUserOrganizations(user.id);
    }
  }, [user]);

  return (
    <ShieldContext.Provider
      value={{
        client: shieldClient,
        organizations,
        setOrganizations,
        groups,
        setGroups,
        strategies,
        setStrategies,
        user,
        setUser,
      }}
    >
      {children}
    </ShieldContext.Provider>
  );
};

export const useShieldClient = (options: ShieldClientOptions) => {
  const shieldClient = React.useMemo(
    () => Shield.getOrCreateInstance(options),
    []
  );

  return { shieldClient };
};

export function useShield() {
  const context = useContext<ShieldContextProviderProps>(ShieldContext);
  return context ? context : (initialValues as ShieldContextProviderProps);
}
