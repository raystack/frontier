import {
  Organization,
  OrganizationSchema,
  type Role,
  type BillingAccount,
  type User,
  type OrganizationKyc,
  type BillingAccountDetails,
} from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
import { createContext } from "react";

export interface SearchConfig {
  setVisibility: (isVisible: boolean) => void;
  isVisible: boolean;
  query: string;
  onChange: (query: string) => void;
}

interface OrganizationContextType {
  roles: Role[];
  organization?: Organization;
  updateOrganization: (organization: Organization) => Promise<void>;
  search: SearchConfig;
  billingAccount?: BillingAccount;
  billingAccountDetails?: BillingAccountDetails;
  isBillingAccountLoading: boolean;
  fetchBillingAccountDetails: () => void;
  tokenBalance: string;
  isTokenBalanceLoading: boolean;
  fetchTokenBalance: () => void;
  orgMembersMap: Record<string, User>;
  isOrgMembersMapLoading: boolean;
  updateKYCDetails: (kycDetails: OrganizationKyc | undefined) => void;
  kycDetails?: OrganizationKyc;
  isKYCLoading: boolean;
}

const defaultOrganiztionContextValue = {
  roles: [],
  organization: create(OrganizationSchema),
  updateOrganization: async () => {},
  isBillingAccountLoading: false,
  fetchBillingAccountDetails: () => {},
  tokenBalance: "",
  isTokenBalanceLoading: false,
  fetchTokenBalance: () => {},
  search: {
    setVisibility: () => {},
    isVisible: false,
    query: "",
    onChange: () => {},
  },
  orgMembersMap: {},
  isOrgMembersMapLoading: false,
  updateKYCDetails: () => {},
  kycDetails: undefined,
  isKYCLoading: false,
};

export const OrganizationContext = createContext<OrganizationContextType>(defaultOrganiztionContextValue);
