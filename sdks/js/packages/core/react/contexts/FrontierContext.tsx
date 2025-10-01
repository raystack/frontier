import {
  Dispatch,
  SetStateAction,
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState
} from 'react';
import { useQuery as useConnectQuery } from '@connectrpc/connect-query';
import { FrontierServiceQueries } from '~hooks';

import {
  FrontierClientOptions,
  FrontierProviderProps
} from '../../shared/types';

import { V1Beta1 } from '../../api-client/V1Beta1';
import {
  V1Beta1Organization,
  V1Beta1Plan
} from '../../api-client/data-contracts';
import {
  User,
  Group,
  Organization,
  OrganizationKyc,
  GetOrganizationKycRequestSchema,
  GetBillingAccountRequestSchema,
  ListBillingAccountsRequestSchema,
  ListSubscriptionsRequestSchema,
  BillingAccount,
  BillingAccountDetails,
  PaymentMethod,
  Subscription
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import {
  getActiveSubscription,
  getDefaultPaymentMethod,
  enrichBasePlan,
  defaultFetch,
  getTrialingSubscription
} from '../utils';
import {
  DEFAULT_DATE_FORMAT,
  DEFAULT_DATE_SHORT_FORMAT
} from '../utils/constants';

interface FrontierContextProviderProps {
  config: FrontierClientOptions;
  client: V1Beta1<unknown> | undefined;

  organizations: Organization[];

  groups: Group[];

  user: User | undefined;

  activeOrganization: V1Beta1Organization | undefined;
  setActiveOrganization: Dispatch<
    SetStateAction<V1Beta1Organization | undefined>
  >;

  isActiveOrganizationLoading: boolean;
  setIsActiveOrganizationLoading: Dispatch<SetStateAction<boolean>>;

  isUserLoading: boolean;

  billingAccount: BillingAccount | undefined;

  isBillingAccountLoading: boolean;

  trialSubscription: Subscription | undefined;
  activeSubscription: Subscription | undefined;
  setActiveSubscription: Dispatch<
    SetStateAction<Subscription | undefined>
  >;

  subscriptions: Subscription[];

  isActiveSubscriptionLoading: boolean;
  setIsActiveSubscriptionLoading: Dispatch<SetStateAction<boolean>>;

  trialPlan: V1Beta1Plan | undefined;
  activePlan: V1Beta1Plan | undefined;
  setActivePlan: Dispatch<SetStateAction<V1Beta1Plan | undefined>>;

  allPlans: V1Beta1Plan[];
  isAllPlansLoading: boolean;

  isActivePlanLoading: boolean;
  setIsActivePlanLoading: Dispatch<SetStateAction<boolean>>;

  fetchActiveSubsciption: () => Promise<Subscription | undefined>;

  paymentMethod: PaymentMethod | undefined;

  basePlan?: V1Beta1Plan;

  organizationKyc: OrganizationKyc | undefined;

  isOrganizationKycLoading: boolean;

  billingDetails: BillingAccountDetails | undefined;
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

  groups: [],

  user: undefined,

  activeOrganization: undefined,
  setActiveOrganization: () => undefined,

  isUserLoading: false,

  isActiveOrganizationLoading: false,
  setIsActiveOrganizationLoading: () => undefined,

  billingAccount: undefined,

  isBillingAccountLoading: false,

  trialSubscription: undefined,
  activeSubscription: undefined,
  setActiveSubscription: () => undefined,

  subscriptions: [],

  isActiveSubscriptionLoading: false,
  setIsActiveSubscriptionLoading: () => false,

  trialPlan: undefined,
  activePlan: undefined,
  setActivePlan: () => undefined,

  allPlans: [],
  isAllPlansLoading: false,

  isActivePlanLoading: false,
  setIsActivePlanLoading: () => false,

  fetchActiveSubsciption: async () => undefined,

  paymentMethod: undefined,

  basePlan: undefined,

  organizationKyc: undefined,

  isOrganizationKycLoading: false,

  billingDetails: undefined
};

export const FrontierContext =
  createContext<FrontierContextProviderProps>(initialValues);
FrontierContext.displayName = 'FrontierContext ';

export const FrontierContextProvider = ({
  children,
  config,
  customFetch
}: FrontierProviderProps) => {
  const [activeOrganization, setActiveOrganization] =
    useState<V1Beta1Organization>();
  const [isActiveOrganizationLoading, setIsActiveOrganizationLoading] =
    useState(false);

  const frontierClient = useMemo(
    () =>
      new V1Beta1({
        customFetch: customFetch
          ? customFetch(activeOrganization)
          : defaultFetch,
        baseUrl: config.endpoint,
        baseApiParams: {
          credentials: 'include'
        }
      }),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [activeOrganization?.id, config.endpoint]
  );

  const [isActiveSubscriptionLoading, setIsActiveSubscriptionLoading] =
    useState(false);
  const [activeSubscription, setActiveSubscription] =
    useState<Subscription>();

  const [trialSubscription, setTrialSubscription] =
    useState<Subscription>();

  const [allPlans, setAllPlans] = useState<V1Beta1Plan[]>([]);
  const [isAllPlansLoading, setIsAllPlansLoading] = useState(false);

  const [activePlan, setActivePlan] = useState<V1Beta1Plan>();
  const [trialPlan, setTrialPlan] = useState<V1Beta1Plan>();
  const [isActivePlanLoading, setIsActivePlanLoading] = useState(false);

  const [basePlan, setBasePlan] = useState<V1Beta1Plan>();

  const { data: currentUserData, isLoading: isUserLoading } = useConnectQuery(
    FrontierServiceQueries.getCurrentUser
  );

  const user = currentUserData?.user;

  const { data: groupsData } = useConnectQuery(
    FrontierServiceQueries.listCurrentUserGroups,
    {},
    { enabled: !!user?.id }
  );

  const { data: organizationsData } = useConnectQuery(
    FrontierServiceQueries.listOrganizationsByCurrentUser,
    {},
    { enabled: !!user?.id }
  );

  const groups = groupsData?.groups || [];
  const organizations = organizationsData?.organizations || [];

  const { data: organizationKycData, isLoading: isOrganizationKycLoading } =
    useConnectQuery(
      FrontierServiceQueries.getOrganizationKyc,
      create(GetOrganizationKycRequestSchema, {
        orgId: activeOrganization?.id ?? ''
      }),
      { enabled: !!activeOrganization?.id }
    );

  const organizationKyc = organizationKycData?.organizationKyc;

  const { data: billingAccountsData } = useConnectQuery(
    FrontierServiceQueries.listBillingAccounts,
    create(ListBillingAccountsRequestSchema, {
      orgId: activeOrganization?.id ?? ''
    }),
    { enabled: !!activeOrganization?.id }
  );

  const billingAccountId = billingAccountsData?.billingAccounts?.[0]?.id || '';

  const { data: billingAccountData } = useConnectQuery(
    FrontierServiceQueries.getBillingAccount,
    create(GetBillingAccountRequestSchema, {
      id: billingAccountId,
      orgId: activeOrganization?.id ?? '',
      withPaymentMethods: true,
      withBillingDetails: true
    }),
    { enabled: !!activeOrganization?.id && !!billingAccountId }
  );

  const billingAccount = billingAccountData?.billingAccount;
  const billingDetails = billingAccountData?.billingDetails;
  const paymentMethod = useMemo(() => {
    if (billingAccountData?.paymentMethods) {
      return getDefaultPaymentMethod(billingAccountData.paymentMethods);
    }
    return undefined;
  }, [billingAccountData?.paymentMethods]);

  const { data: subscriptionsData } = useConnectQuery(
    FrontierServiceQueries.listSubscriptions,
    create(ListSubscriptionsRequestSchema, {
      orgId: activeOrganization?.id ?? '',
      billingId: billingAccount?.id ?? ''
    }),
    { enabled: !!activeOrganization?.id && !!billingAccount?.id }
  );

  const subscriptions = (subscriptionsData?.subscriptions || []) as Subscription[];

  const getPlan = useCallback(
    async (planId?: string) => {
      if (!planId) return;
      setIsActivePlanLoading(true);
      try {
        const resp = await frontierClient?.frontierServiceGetPlan(planId);
        return resp?.data?.plan;
      } catch (err) {
        console.error(
          'frontier:sdk:: There is problem with fetching active plan'
        );
        console.error(err);
        return;
      } finally {
        setIsActivePlanLoading(false);
      }
    },
    [frontierClient]
  );

  const setActiveAndTrialSubscriptions = useCallback(
    async (subscriptionsList: Subscription[] = []) => {
      const activeSub = getActiveSubscription(subscriptionsList);
      setActiveSubscription(activeSub);
      const activeSubPlan = await getPlan(activeSub?.planId);
      setActivePlan(activeSubPlan);

      const trialSub = getTrialingSubscription(subscriptionsList);
      setTrialSubscription(trialSub);
      const trialSubPlan = await getPlan(trialSub?.planId);
      setTrialPlan(trialSubPlan);

      return [activeSub, trialSub];
    },
    [getPlan]
  );

  const processSubscriptions = useCallback(
    async (subscriptionsList: Subscription[]) => {
      setIsActiveSubscriptionLoading(true);
      try {
        const [activeSub] = await setActiveAndTrialSubscriptions(
          subscriptionsList
        );
        return activeSub;
      } catch (err: any) {
        console.error(
          'frontier:sdk:: There is problem with processing subscriptions'
        );
        console.error(err);
      } finally {
        setIsActiveSubscriptionLoading(false);
      }
    },
    [setActiveAndTrialSubscriptions]
  );

  const fetchActiveSubsciption = useCallback(async () => {
    if (subscriptions.length > 0) {
      return processSubscriptions(subscriptions);
    }
  }, [subscriptions, processSubscriptions]);

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
      fetchAllPlans();
    }
  }, [activeOrganization?.id, fetchAllPlans]);

  useEffect(() => {
    // Process subscriptions when available
    if (subscriptions.length > 0) {
      processSubscriptions(subscriptions);
    } else if (
      billingAccountsData &&
      billingAccountsData.billingAccounts?.length === 0
    ) {
      setActiveSubscription(undefined);
    }
  }, [subscriptions, billingAccountsData, processSubscriptions]);

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
        groups,
        user,
        activeOrganization,
        setActiveOrganization,
        isActiveOrganizationLoading,
        setIsActiveOrganizationLoading,
        isUserLoading,
        billingAccount,
        paymentMethod,
        isBillingAccountLoading: false,
        isActiveSubscriptionLoading,
        setIsActiveSubscriptionLoading,
        trialSubscription,
        activeSubscription,
        setActiveSubscription,
        subscriptions,
        trialPlan,
        activePlan,
        setActivePlan,
        isActivePlanLoading,
        setIsActivePlanLoading,
        fetchActiveSubsciption,
        allPlans,
        isAllPlansLoading,
        basePlan,
        organizationKyc,
        isOrganizationKycLoading,
        billingDetails
      }}
    >
      {children}
    </FrontierContext.Provider>
  );
};

export function useFrontier() {
  const context = useContext<FrontierContextProviderProps>(FrontierContext);
  return context ? context : (initialValues as FrontierContextProviderProps);
}
