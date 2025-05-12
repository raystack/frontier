import { createContext } from "react";
import type {
  V1Beta1Role,
  V1Beta1Organization,
  V1Beta1BillingAccount,
  V1Beta1User,
  V1Beta1OrganizationKyc,
} from "~/api/frontier";

export interface SearchConfig {
  setVisibility: (isVisible: boolean) => void;
  isVisible: boolean;
  query: string;
  onChange: (query: string) => void;
}

interface OrganizationContextType {
  roles: V1Beta1Role[];
  organization?: V1Beta1Organization;
  updateOrganization: (organization: V1Beta1Organization) => Promise<void>;
  search: SearchConfig;
  billingAccount?: V1Beta1BillingAccount;
  tokenBalance: string;
  isTokenBalanceLoading: boolean;
  fetchTokenBalance: (orgId: string, billingAccountId: string) => Promise<void>;
  orgMembersMap: Record<string, V1Beta1User>;
  isOrgMembersMapLoading: boolean;
  updateKYCDetails: (kycDetails: V1Beta1OrganizationKyc) => void;
  kycDetails?: V1Beta1OrganizationKyc;
  isKYCLoading: boolean;
}

export const OrganizationContext = createContext<OrganizationContextType>({
  roles: [],
  organization: {},
  updateOrganization: async () => {},
  billingAccount: {},
  tokenBalance: "",
  isTokenBalanceLoading: false,
  fetchTokenBalance: async () => {},
  search: {
    setVisibility: () => {},
    isVisible: false,
    query: "",
    onChange: () => {},
  },
  orgMembersMap: {},
  isOrgMembersMapLoading: false,
  updateKYCDetails: (kycDetails: V1Beta1OrganizationKyc) => {},
  kycDetails: undefined,
  isKYCLoading: false,
});
