import { createContext } from "react";
import { V1Beta1Role, V1Beta1Organization } from "~/api/frontier";

interface OrganizationContextType {
  roles: V1Beta1Role[];
  organization?: V1Beta1Organization;
}

export const OrganizationContext = createContext<OrganizationContextType>({
  roles: [],
  organization: {},
});
