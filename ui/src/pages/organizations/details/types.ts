import { type Organization } from "@raystack/proton/frontier";
import { type V1Beta1Role } from "~/api/frontier";

export interface OutletContext {
  organizationId: string;
  organization: Organization;
  fetchOrganization: (id: string) => Promise<void>;
  roles: V1Beta1Role[];
}

export const OrganizationStatus = {
  enabled: "enabled",
  disabled: "disabled",
};

export type OrganizationStatusType = keyof typeof OrganizationStatus;

export const ORG_NAMESPACE = "app/organization";
export const PROJECT_NAMESPACE = "app/project";

export const DEFAULT_INVITE_ROLE = "app_organization_viewer";
