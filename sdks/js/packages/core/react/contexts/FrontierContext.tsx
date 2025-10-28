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
import { V1Beta1Organization } from '../../api-client/data-contracts';
import {
  User,
  Group,
  Organization,
  OrganizationKyc,
  GetOrganizationKycRequestSchema,
  GetBillingAccountRequestSchema,
  ListBillingAccountsRequestSchema,
  ListSubscriptionsRequestSchema,
  ListPlansRequestSchema,
  BillingAccount,
  BillingAccountDetails,
  PaymentMethod,
  Subscription,
  Plan
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
import { useLastActiveTracker } from '../hooks/useLastActiveTracker';

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

  subscriptions: Subscription[];

  isActiveSubscriptionLoading: boolean;

  trialPlan: Plan | undefined;
  activePlan: Plan | undefined;

  allPlans: Plan[];
  isAllPlansLoading: boolean;

  fetchActiveSubsciption: () => Promise<Subscription | undefined>;

  paymentMethod: PaymentMethod | undefined;

  basePlan?: Plan;

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

  subscriptions: [],

  isActiveSubscriptionLoading: false,

  trialPlan: undefined,
  activePlan: undefined,

  allPlans: [],
  isAllPlansLoading: false,

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
      {
        enabled: !!activeOrganization?.id,
        retry: false
      }
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

  const {
    data: subscriptionsData,
    isLoading: isActiveSubscriptionLoading,
    refetch: refetchActiveSubscription
  } = useConnectQuery(
    FrontierServiceQueries.listSubscriptions,
    create(ListSubscriptionsRequestSchema, {
      orgId: activeOrganization?.id ?? '',
      billingId: billingAccount?.id ?? ''
    }),
    { enabled: !!activeOrganization?.id && !!billingAccount?.id }
  );

  const subscriptions = (subscriptionsData?.subscriptions ||
    []) as Subscription[];

  const { data: plansData, isLoading: isAllPlansLoading } = useConnectQuery(
    FrontierServiceQueries.listPlans,
    create(ListPlansRequestSchema, {}),
    { enabled: !!activeOrganization?.id }
  );

  const allPlans = (plansData?.plans || []) as Plan[];

  const { activeSubscription, trialSubscription } = useMemo(() => {
    const activeSubscription =
      billingAccountId && subscriptions.length
        ? getActiveSubscription(subscriptions)
        : undefined;
    const trialSubscription =
      billingAccountId && subscriptions.length
        ? getTrialingSubscription(subscriptions)
        : undefined;
    return { activeSubscription, trialSubscription };
  }, [subscriptions, billingAccountId]);

  const activePlan = useMemo(() => {
    return allPlans.find(p => p.id === activeSubscription?.planId);
  }, [allPlans, activeSubscription?.planId]);

  const trialPlan = useMemo(() => {
    return allPlans.find(p => p.id === trialSubscription?.planId);
  }, [allPlans, trialSubscription?.planId]);

  const fetchActiveSubsciption = useCallback(async () => {
    const refetchedData = await refetchActiveSubscription();
    const refetchedSubscriptions = refetchedData?.data?.subscriptions || [];
    return getActiveSubscription(refetchedSubscriptions);
  }, [refetchActiveSubscription]);

  const basePlan = useMemo(() => {
    return config?.billing?.basePlan
      ? enrichBasePlan(config.billing.basePlan)
      : undefined;
  }, [config?.billing?.basePlan]);

  // Track last user activity for session management
  useLastActiveTracker();

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
        trialSubscription,
        activeSubscription,
        subscriptions,
        trialPlan,
        activePlan,
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
