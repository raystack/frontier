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

import { V1Beta1 } from '../../api-client/V1Beta1';
import {
  V1Beta1AuthStrategy,
  V1Beta1BillingAccount,
  V1Beta1Group,
  V1Beta1Organization,
  V1Beta1PaymentMethod,
  V1Beta1Plan,
  V1Beta1Subscription,
  V1Beta1User
} from '../../api-client/data-contracts';
import Frontier from '../frontier';
import {
  getActiveSubscription,
  getDefaultPaymentMethod,
  enrichBasePlan
} from '../utils';
import {
  DEFAULT_DATE_FORMAT,
  DEFAULT_DATE_SHORT_FORMAT
} from '../utils/constants';

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

  subscriptions: V1Beta1Subscription[];

  isActiveSubscriptionLoading: boolean;
  setIsActiveSubscriptionLoading: Dispatch<SetStateAction<boolean>>;

  activePlan: V1Beta1Plan | undefined;
  setActivePlan: Dispatch<SetStateAction<V1Beta1Plan | undefined>>;

  allPlans: V1Beta1Plan[];
  isAllPlansLoading: boolean;

  isActivePlanLoading: boolean;
  setIsActivePlanLoading: Dispatch<SetStateAction<boolean>>;

  fetchActiveSubsciption: () => Promise<V1Beta1Subscription | undefined>;

  paymentMethod: V1Beta1PaymentMethod | undefined;

  basePlan?: V1Beta1Plan;
}

const defaultConfig: FrontierClientOptions = {
  endpoint: 'http://localhost:8080',
  redirectLogin: 'http://localhost:3000',
  redirectSignup: 'http://localhost:3000/signup',
  redirectMagicLinkVerify: 'http://localhost:3000/magiclink-verify',
  callbackUrl: 'http://localhost:3000/callback',
  dateFormat: DEFAULT_DATE_FORMAT,
  shortDateFormat: DEFAULT_DATE_SHORT_FORMAT,
  billing: {
    successUrl: 'http://localhost:3000/success',
    cancelUrl: 'http://localhost:3000/cancel',
    cancelAfterTrial: true,
    showPerMonthPrice: false
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

  subscriptions: [],

  isActiveSubscriptionLoading: false,
  setIsActiveSubscriptionLoading: () => false,

  activePlan: undefined,
  setActivePlan: () => undefined,

  allPlans: [],
  isAllPlansLoading: false,

  isActivePlanLoading: false,
  setIsActivePlanLoading: () => false,

  fetchActiveSubsciption: async () => undefined,

  paymentMethod: undefined,

  basePlan: undefined
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
  const [paymentMethod, setPaymentMethod] = useState<V1Beta1PaymentMethod>();
  const [isBillingAccountLoading, setIsBillingAccountLoading] = useState(false);

  const [isActiveSubscriptionLoading, setIsActiveSubscriptionLoading] =
    useState(false);
  const [activeSubscription, setActiveSubscription] =
    useState<V1Beta1Subscription>();
  const [subscriptions, setSubscriptions] = useState<V1Beta1Subscription[]>([]);

  const [allPlans, setAllPlans] = useState<V1Beta1Plan[]>([]);
  const [isAllPlansLoading, setIsAllPlansLoading] = useState(false);

  const [activePlan, setActivePlan] = useState<V1Beta1Plan>();
  const [isActivePlanLoading, setIsActivePlanLoading] = useState(false);

  const [basePlan, setBasePlan] = useState<V1Beta1Plan>();

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

  const getPlan = useCallback(
    async (planId: string) => {
      setIsActivePlanLoading(true);

      try {
        const resp = await frontierClient?.frontierServiceGetPlan(planId);
        const plan = resp?.data?.plan;
        if (plan) {
          setActivePlan(plan);
        }
      } catch (err) {
        console.error(
          'frontier:sdk:: There is problem with fetching active plan'
        );
        console.error(err);
      } finally {
        setIsActivePlanLoading(false);
      }
    },
    [frontierClient]
  );

  const getSubscription = useCallback(
    async (orgId: string, billingId: string) => {
      setIsActiveSubscriptionLoading(true);
      try {
        const resp = await frontierClient?.frontierServiceListSubscriptions(
          orgId,
          billingId
        );
        const subscriptionsList = resp?.data?.subscriptions || [];
        setSubscriptions(subscriptionsList);
        if (subscriptionsList.length) {
          const activeSub = getActiveSubscription(subscriptionsList);
          setActiveSubscription(activeSub);
          if (activeSub?.plan_id) {
            getPlan(activeSub?.plan_id);
          }
          return activeSub;
        } else {
          setActiveSubscription(undefined);
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
    [frontierClient, getPlan]
  );

  const getBillingAccount = useCallback(
    async (orgId: string) => {
      setIsBillingAccountLoading(true);
      try {
        const {
          data: { billing_accounts = [] }
        } = await frontierClient.frontierServiceListBillingAccounts(orgId);
        const billingAccountId = billing_accounts[0]?.id || '';
        if (billingAccountId) {
          const [resp] = await Promise.all([
            frontierClient?.frontierServiceGetBillingAccount(
              orgId,
              billingAccountId,
              { with_payment_methods: true }
            ),
            getSubscription(orgId, billingAccountId)
          ]);

          if (resp?.data) {
            const paymentMethods = resp?.data?.payment_methods || [];
            setBillingAccount(resp.data.billing_account);
            const defaultPaymentMethod =
              getDefaultPaymentMethod(paymentMethods);
            setPaymentMethod(defaultPaymentMethod);
          }
        } else {
          setBillingAccount(undefined);
          setActiveSubscription(undefined);
        }
      } catch (error) {
        console.error(
          'frontier:sdk:: There is problem with fetching org billing accounts'
        );
        console.error(error);
      } finally {
        setIsBillingAccountLoading(false);
      }
    },
    [frontierClient, getSubscription]
  );

  const fetchActiveSubsciption = useCallback(async () => {
    if (activeOrganization?.id && billingAccount?.id) {
      return getSubscription(activeOrganization?.id, billingAccount?.id);
    }
  }, [activeOrganization?.id, billingAccount?.id, getSubscription]);

  const fetchAllPlans = useCallback(async () => {
    try {
      setIsAllPlansLoading(true);
      const resp = await frontierClient.frontierServiceListPlans();
      const plans = resp?.data?.plans || [];
      setAllPlans(plans);
    } catch (err) {
      console.error('frontier:sdk:: There is problem with fetching plans');
      console.error(err);
    } finally {
      setIsAllPlansLoading(false);
    }
  }, [frontierClient]);

  useEffect(() => {
    if (activeOrganization?.id) {
      getBillingAccount(activeOrganization.id);
      fetchAllPlans();
    }
  }, [activeOrganization?.id, getBillingAccount, fetchAllPlans]);

  useEffect(() => {
    if (config?.billing?.basePlan) {
      setBasePlan(enrichBasePlan(config.billing.basePlan));
    }
  }, [config?.billing?.basePlan]);

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
        paymentMethod,
        isBillingAccountLoading,
        setIsBillingAccountLoading,
        isActiveSubscriptionLoading,
        setIsActiveSubscriptionLoading,
        activeSubscription,
        setActiveSubscription,
        subscriptions,
        activePlan,
        setActivePlan,
        isActivePlanLoading,
        setIsActivePlanLoading,
        fetchActiveSubsciption,
        allPlans,
        isAllPlansLoading,
        basePlan
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
