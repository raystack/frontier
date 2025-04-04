import { createContext } from "react";
import {
  V1Beta1Role,
  V1Beta1Organization,
  V1Beta1BillingAccount,
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
  fetchTokenBalance: (orgId: string, billingAccountId: string) => Promise<void>;
}

export const OrganizationContext = createContext<OrganizationContextType>({
  roles: [],
  organization: {},
  billingAccount: {},
  tokenBalance: "",
  fetchTokenBalance: async () => {},
  search: {
    setVisibility: () => {},
    isVisible: false,
    query: "",
    onChange: () => {},
  },
});
