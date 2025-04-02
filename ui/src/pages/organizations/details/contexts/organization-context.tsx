import { createContext } from "react";
import { V1Beta1Role, V1Beta1Organization } from "~/api/frontier";

export interface SearchConfig {
  setVisibility?: (isVisible: boolean) => void;
  isVisible?: boolean;
  query?: string;
  onChange?: (query: string) => void;
}

interface OrganizationContextType {
  roles: V1Beta1Role[];
  organization?: V1Beta1Organization;
  search: SearchConfig;
}

export const OrganizationContext = createContext<OrganizationContextType>({
  roles: [],
  organization: {},
  search: {},
});
