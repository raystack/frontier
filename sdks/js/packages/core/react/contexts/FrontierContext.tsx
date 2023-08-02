import React, {
  Dispatch,
  SetStateAction,
  createContext,
  useContext,
  useEffect,
  useState
} from 'react';

import {
  FrontierClientOptions,
  FrontierProviderProps
} from '../../shared/types';

import { V1Beta1 } from '../../client/V1Beta1';
import {
  V1Beta1AuthStrategy,
  V1Beta1Group,
  V1Beta1Organization,
  V1Beta1User
} from '../../client/data-contracts';
import Frontier from '../frontier';
interface FrontierContextProviderProps {
  config: FrontierClientOptions;
  client: V1Beta1<unknown> | undefined;

  organizations: V1Beta1Organization[];
  setOrganizations: Dispatch<SetStateAction<V1Beta1Organization[]>>;

  groups: V1Beta1Group[];
  setGroups: Dispatch<SetStateAction<V1Beta1Group[]>>;

  strategies: V1Beta1AuthStrategy[];
  setStrategies: Dispatch<SetStateAction<V1Beta1AuthStrategy[]>>;

  user: V1Beta1User | undefined;
  setUser: Dispatch<SetStateAction<V1Beta1User | undefined>>;
}

const defaultConfig = {
  endpoint: 'http://localhost:8080',
  redirectLogin: 'http://localhost:3000',
  redirectSignup: 'http://localhost:3000/signup',
  redirectMagicLinkVerify: 'http://localhost:3000/magiclink-verify'
};

const initialValues: FrontierContextProviderProps = {
  config: defaultConfig,
  client: undefined,

  organizations: [],
  setOrganizations: () => undefined,

  groups: [],
  setGroups: () => undefined,

  strategies: [],
  setStrategies: () => undefined,

  user: undefined,
  setUser: () => undefined
};

export const FrontierContext =
  createContext<FrontierContextProviderProps>(initialValues);
FrontierContext.displayName = 'FrontierContext ';

export const FrontierContextProvider = ({
  children,
  config,
  initialState,
  ...options
}: FrontierProviderProps) => {
  const { frontierClient } = useFrontierClient(config);

  const [organizations, setOrganizations] = useState<V1Beta1Organization[]>([]);
  const [groups, setGroups] = useState<V1Beta1Group[]>([]);
  const [strategies, setStrategies] = useState<V1Beta1AuthStrategy[]>([]);
  const [user, setUser] = useState<V1Beta1User>();

  useEffect(() => {
    async function getFrontierInformation() {
      try {
        const {
          data: { strategies = [] }
        } = await frontierClient.frontierServiceListAuthStrategies();
        setStrategies(strategies);
      } catch (error) {
        console.error(
          'frontier:sdk:: There is problem with fetching auth strategies'
        );
      }
    }
    getFrontierInformation();
  }, []);

  useEffect(() => {
    async function getFrontierCurrentUser() {
      try {
        const {
          data: { user }
        } = await frontierClient.frontierServiceGetCurrentUser();
        setUser(user);
      } catch (error) {
        console.error(
          'frontier:sdk:: There is problem with fetching current user information'
        );
      }
    }
    getFrontierCurrentUser();
  }, []);

  useEffect(() => {
    async function getFrontierCurrentUserGroups(userId: string) {
      try {
        const {
          data: { groups = [] }
        } = await frontierClient.frontierServiceListUserGroups(userId);
        setGroups(groups);
      } catch (error) {
        console.error(
          'frontier:sdk:: There is problem with fetching user groups information'
        );
      }
    }

    if (user?.id) {
      getFrontierCurrentUserGroups(user.id);
    }
  }, [user]);

  useEffect(() => {
    async function getFrontierCurrentUserOrganizations(userId: string) {
      try {
        const {
          data: { organizations = [] }
        } = await frontierClient.frontierServiceGetOrganizationsByUser(userId);
        setOrganizations(organizations);
      } catch (error) {
        console.error(
          'frontier:sdk:: There is problem with fetching user current organizations'
        );
      }
    }

    if (user?.id) {
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
        setUser
      }}
    >
      {children}
    </FrontierContext.Provider>
  );
};

export const useFrontierClient = (options: FrontierClientOptions) => {
  const frontierClient = React.useMemo(
    () => Frontier.getInstance({ endpoint: options.endpoint }),
    []
  );

  return { frontierClient };
};

export function useFrontier() {
  const context = useContext<FrontierContextProviderProps>(FrontierContext);
  return context ? context : (initialValues as FrontierContextProviderProps);
}
