import React, {
  Dispatch,
  SetStateAction,
  createContext,
  useCallback,
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
  V1Beta1BillingAccount,
  V1Beta1Group,
  V1Beta1Organization,
  V1Beta1Subscription,
  V1Beta1User
} from '../../client/data-contracts';
import Frontier from '../frontier';
import { getActiveSubscription } from '../utils';
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

  activeOrganization: V1Beta1Organization | undefined;
  setActiveOrganization: Dispatch<
    SetStateAction<V1Beta1Organization | undefined>
  >;

  isActiveOrganizationLoading: boolean;
  setIsActiveOrganizationLoading: Dispatch<SetStateAction<boolean>>;

  isUserLoading: boolean;
  setIsUserLoading: Dispatch<SetStateAction<boolean>>;

  billingAccount: V1Beta1BillingAccount | undefined;
  setBillingAccount: Dispatch<
    SetStateAction<V1Beta1BillingAccount | undefined>
  >;

  isBillingAccountLoading: boolean;
  setIsBillingAccountLoading: Dispatch<SetStateAction<boolean>>;

  activeSubscription: V1Beta1Subscription | undefined;
  setActiveSubscription: Dispatch<
    SetStateAction<V1Beta1Subscription | undefined>
  >;

  isActiveSubscriptionLoading: boolean;
  setIsActiveSubscriptionLoading: Dispatch<SetStateAction<boolean>>;
}

const defaultConfig: FrontierClientOptions = {
  endpoint: 'http://localhost:8080',
  redirectLogin: 'http://localhost:3000',
  redirectSignup: 'http://localhost:3000/signup',
  redirectMagicLinkVerify: 'http://localhost:3000/magiclink-verify',
  callbackUrl: 'http://localhost:3000/callback',
  billing: {
    successUrl: 'http://localhost:3000/success',
    cancelUrl: 'http://localhost:3000/cancel'
  }
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
  setUser: () => undefined,

  activeOrganization: undefined,
  setActiveOrganization: () => undefined,

  isUserLoading: false,
  setIsUserLoading: () => undefined,

  isActiveOrganizationLoading: false,
  setIsActiveOrganizationLoading: () => undefined,

  billingAccount: undefined,
  setBillingAccount: () => undefined,

  isBillingAccountLoading: false,
  setIsBillingAccountLoading: () => false,

  activeSubscription: undefined,
  setActiveSubscription: () => undefined,

  isActiveSubscriptionLoading: false,
  setIsActiveSubscriptionLoading: () => false
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
  const [activeOrganization, setActiveOrganization] =
    useState<V1Beta1Organization>();
  const [isActiveOrganizationLoading, setIsActiveOrganizationLoading] =
    useState(false);
  const [isUserLoading, setIsUserLoading] = useState(false);

  const [billingAccount, setBillingAccount] = useState<V1Beta1BillingAccount>();
  const [isBillingAccountLoading, setIsBillingAccountLoading] = useState(false);

  const [isActiveSubscriptionLoading, setIsActiveSubscriptionLoading] =
    useState(false);
  const [activeSubscription, setActiveSubscription] =
    useState<V1Beta1Subscription>();

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
        setIsUserLoading(true);
        const {
          data: { user }
        } = await frontierClient.frontierServiceGetCurrentUser();
        setUser(user);
      } catch (error) {
        console.error(
          'frontier:sdk:: There is problem with fetching current user information'
        );
      } finally {
        setIsUserLoading(false);
      }
    }
    getFrontierCurrentUser();
  }, [frontierClient]);

  const getFrontierCurrentUserGroups = useCallback(async () => {
    try {
      const {
        data: { groups = [] }
      } = await frontierClient.frontierServiceListCurrentUserGroups();
      setGroups(groups);
    } catch (error) {
      console.error(
        'frontier:sdk:: There is problem with fetching user groups information'
      );
    }
  }, [frontierClient]);

  const getFrontierCurrentUserOrganizations = useCallback(async () => {
    try {
      const {
        data: { organizations = [] }
      } = await frontierClient.frontierServiceListOrganizationsByCurrentUser();
      setOrganizations(organizations);
    } catch (error) {
      console.error(
        'frontier:sdk:: There is problem with fetching user current organizations'
      );
    }
  }, [frontierClient]);

  useEffect(() => {
    if (user?.id) {
      getFrontierCurrentUserGroups();
      getFrontierCurrentUserOrganizations();
    }
  }, [getFrontierCurrentUserGroups, getFrontierCurrentUserOrganizations, user]);

  const getSubscription = useCallback(
    async (orgId: string, billingId: string) => {
      setIsActiveSubscriptionLoading(true);
      try {
        const resp = await frontierClient?.frontierServiceListSubscriptions(
          orgId,
          billingId
        );
        if (resp?.data?.subscriptions?.length) {
          const activeSub = getActiveSubscription(resp?.data?.subscriptions);
          setActiveSubscription(activeSub);
        }
      } catch (err: any) {
        console.error(
          'frontier:sdk:: There is problem with fetching active subscriptions'
        );
        console.error(err);
      } finally {
        setIsActiveSubscriptionLoading(false);
      }
    },
    [frontierClient]
  );

  const getBillingAccount = useCallback(
    async (orgId: string) => {
      setIsBillingAccountLoading(true);
      try {
        const {
          data: { billing_accounts = [] }
        } = await frontierClient.frontierServiceListBillingAccounts(orgId);
        if (billing_accounts.length > 0) {
          const billing_account = billing_accounts[0];
          setBillingAccount(billing_account);
          await getSubscription(orgId, billing_account?.id || '');
        }
      } catch (error) {
        console.error(
          'frontier:sdk:: There is problem with fetching org billing accounts'
        );
      } finally {
        setIsBillingAccountLoading(false);
      }
    },
    [frontierClient, getSubscription]
  );

  useEffect(() => {
    if (activeOrganization?.id) {
      getBillingAccount(activeOrganization.id);
    }
  }, [activeOrganization?.id, getBillingAccount]);

  return (
    <FrontierContext.Provider
      value={{
        config: {
          ...defaultConfig,
          ...config,
          billing: { ...defaultConfig.billing, ...config.billing }
        },
        client: frontierClient,
        organizations,
        setOrganizations,
        groups,
        setGroups,
        strategies,
        setStrategies,
        user,
        setUser,
        activeOrganization,
        setActiveOrganization,
        isActiveOrganizationLoading,
        setIsActiveOrganizationLoading,
        isUserLoading,
        setIsUserLoading,
        billingAccount,
        setBillingAccount,
        isBillingAccountLoading,
        setIsBillingAccountLoading,
        isActiveSubscriptionLoading,
        setIsActiveSubscriptionLoading,
        activeSubscription,
        setActiveSubscription
      }}
    >
      {children}
    </FrontierContext.Provider>
  );
};

export const useFrontierClient = (options: FrontierClientOptions) => {
  const frontierClient = React.useMemo(() => Frontier.getInstance(options), []);

  return { frontierClient };
};

export function useFrontier() {
  const context = useContext<FrontierContextProviderProps>(FrontierContext);
  return context ? context : (initialValues as FrontierContextProviderProps);
}
