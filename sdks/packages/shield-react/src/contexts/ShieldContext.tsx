import {
  Group,
  Organization,
  FrontierClientOptions,
  FrontierProviderProps,
  Strategy,
  User,
} from "@raystack/frontier";
import React, {
  createContext,
  Dispatch,
  SetStateAction,
  useContext,
  useEffect,
  useState,
} from "react";

import Frontier from "../frontier";
interface FrontierContextProviderProps {
  config: FrontierClientOptions;
  client: Frontier;

  organizations: Organization[];
  setOrganizations: Dispatch<SetStateAction<Organization[]>>;

  groups: Group[];
  setGroups: Dispatch<SetStateAction<Group[]>>;

  strategies: Strategy[];
  setStrategies: Dispatch<SetStateAction<Strategy[]>>;

  user: User | null;
  setUser: Dispatch<SetStateAction<User | null>>;
}

const defaultConfig = {
  endpoint: "http://localhost:8080",
  redirectLogin: "http://localhost:3000",
  redirectSignup: "http://localhost:3000/signup",
  redirectMagicLinkVerify: "http://localhost:3000/magiclink-verify",
};

const initialValues: FrontierContextProviderProps = {
  config: defaultConfig,
  client: Frontier.getOrCreateInstance(defaultConfig),

  organizations: [],
  setOrganizations: () => undefined,

  groups: [],
  setGroups: () => undefined,

  strategies: [],
  setStrategies: () => undefined,

  user: null,
  setUser: () => undefined,
};

export const FrontierContext =
  createContext<FrontierContextProviderProps>(initialValues);
FrontierContext.displayName = "FrontierContext ";

export const FrontierContextProvider = ({
  children,
  config,
  initialState,
  ...options
}: FrontierProviderProps) => {
  const { frontierClient } = useFrontierClient(config);

  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [groups, setGroups] = useState<Group[]>([]);
  const [strategies, setStrategies] = useState<Strategy[]>([]);
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    async function getFrontierInformation() {
      try {
        const {
          data: { strategies },
        } = await frontierClient.getAuthAtrategies();
        setStrategies(strategies);
      } catch (error) {
        console.error(
          "frontier:sdk:: There is problem with fetching auth strategies"
        );
      }
    }
    getFrontierInformation();
  }, []);

  useEffect(() => {
    async function getFrontierCurrentUser() {
      try {
        const {
          data: { user },
        } = await frontierClient.getCurrentUser();
        setUser(user);
      } catch (error) {
        console.error(
          "frontier:sdk:: There is problem with fetching current user information"
        );
      }
    }
    getFrontierCurrentUser();
  }, []);

  useEffect(() => {
    async function getFrontierCurrentUserGroups(userId: string) {
      try {
        const {
          data: { groups },
        } = await frontierClient.getUserGroups(userId);
        setGroups(groups);
      } catch (error) {
        console.error(
          "frontier:sdk:: There is problem with fetching user groups information"
        );
      }
    }

    if (user) {
      getFrontierCurrentUserGroups(user.id);
    }
  }, [user]);

  useEffect(() => {
    async function getFrontierCurrentUserOrganizations(userId: string) {
      try {
        const {
          data: { organizations },
        } = await frontierClient.getUserOrganisations(userId);
        setOrganizations(organizations);
      } catch (error) {
        console.error(
          "frontier:sdk:: There is problem with fetching user current organizations"
        );
      }
    }

    if (user) {
      getFrontierCurrentUserOrganizations(user.id);
    }
  }, [user]);

  return (
    <FrontierContext.Provider
      value={{
        config: { ...defaultConfig, ...config },
        client: frontierClient,
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
    </FrontierContext.Provider>
  );
};

export const useFrontierClient = (options: FrontierClientOptions) => {
  const frontierClient = React.useMemo(
    () => Frontier.getOrCreateInstance(options),
    []
  );

  return { frontierClient };
};

export function useFrontier() {
  const context = useContext<FrontierContextProviderProps>(FrontierContext);
  return context ? context : (initialValues as FrontierContextProviderProps);
}
