import { V1Beta1Organization, V1Beta1Role } from "~/api/frontier";

export interface OutletContext {
  organizationId: string;
  organization: V1Beta1Organization;
  fetchOrganization: (id: string) => Promise<void>;
  roles: V1Beta1Role[];
}

export const OrganizationStatus = {
  enabled: "enabled",
  disabled: "disabled",
};

export type OrganizationStatusType = keyof typeof OrganizationStatus;

export const ORG_NAMESPACE = "app/organization";

export const DEFAULT_INVITE_ROLE = "app_organization_viewer";
