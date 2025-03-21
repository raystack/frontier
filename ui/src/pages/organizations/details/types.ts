import { V1Beta1Organization } from "@raystack/frontier";

export interface OutletContext {
  organizationId: string;
  organization: V1Beta1Organization;
  fetchOrganization: (id: string) => Promise<void>;
}
