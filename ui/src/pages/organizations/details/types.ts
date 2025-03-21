import { V1Beta1Organization } from "@raystack/frontier";

export interface OutletContext {
  organizationId: string;
  organization: V1Beta1Organization;
  fetchOrganization: (id: string) => Promise<void>;
}

export const OrganizationStatus = {
  enabled: "enabled",
  disabled: "disabled",
};

export type OrganizationStatusType = keyof typeof OrganizationStatus;
