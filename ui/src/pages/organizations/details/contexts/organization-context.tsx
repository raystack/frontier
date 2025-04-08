import { createContext } from "react";
import {
  V1Beta1Role,
  V1Beta1Organization,
  V1Beta1BillingAccount,
  V1Beta1User,
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
  search: SearchConfig;
  billingAccount?: V1Beta1BillingAccount;
  tokenBalance: string;
  isTokenBalanceLoading: boolean;
  fetchTokenBalance: (orgId: string, billingAccountId: string) => Promise<void>;
  orgMembersMap: Record<string, V1Beta1User>;
  isOrgMembersMapLoading: boolean;
}

export const OrganizationContext = createContext<OrganizationContextType>({
  roles: [],
  organization: {},
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
});
