import { type Organization } from "@raystack/proton/frontier";

export interface OutletContext {
  organization: Organization;
}

export const OrganizationStatus = {
  enabled: "enabled",
  disabled: "disabled",
};

export type OrganizationStatusType = keyof typeof OrganizationStatus;

export const ORG_NAMESPACE = "app/organization";
export const PROJECT_NAMESPACE = "app/project";

export const DEFAULT_INVITE_ROLE = "app_organization_viewer";
