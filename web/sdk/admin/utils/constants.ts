export const SCOPES = {
  ORG: "app/organization",
  PROJECT: "app/project",
  GROUP: "app/group",
} as const;

export const DEFAULT_ROLES = {
  ORG_MANAGER: "app_organization_manager",
  ORG_OWNER: "app_organization_owner",
  ORG_BILLING_MANAGER: "app_billing_manager",
  ORG_VIEWER: "app_organization_viewer",
  PROJECT_VIEWER: "app_project_viewer",
} as const;

export const NULL_DATE = "0001-01-01T00:00:00Z";

export interface Config {
  title?: string;
  app_url?: string;
  token_product_id?: string;
  organization_types?: string[];
}

export const defaultConfig: Config = {
  title: "Frontier Admin",
  app_url: "example.com",
  token_product_id: "token",
  organization_types: [],
};
