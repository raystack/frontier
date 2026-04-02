import { EntityTerminologies } from "../../shared/types";

export const SCOPES = {
  ORG: "app/organization",
  PROJECT: "app/project",
  GROUP: "app/group",
  USER: "app/user",
} as const;

export const DEFAULT_ROLES = {
  ORG_MANAGER: "app_organization_manager",
  ORG_OWNER: "app_organization_owner",
  ORG_BILLING_MANAGER: "app_billing_manager",
  ORG_VIEWER: "app_organization_viewer",
  PROJECT_VIEWER: "app_project_viewer",
} as const;

export const NULL_DATE = "0001-01-01T00:00:00Z";

export interface AdminTerminologyConfig {
  organization?: EntityTerminologies;
  project?: EntityTerminologies;
  team?: EntityTerminologies;
  member?: EntityTerminologies;
  user?: EntityTerminologies;
  appName?: string;
}

export interface Config {
  title?: string;
  app_url?: string;
  token_product_id?: string;
  organization_types?: string[];
  terminology?: AdminTerminologyConfig;
}

export const defaultTerminology: Required<AdminTerminologyConfig> = {
  organization: { singular: "Organization", plural: "Organizations" },
  project: { singular: "Project", plural: "Projects" },
  team: { singular: "Team", plural: "Teams" },
  member: { singular: "Member", plural: "Members" },
  user: { singular: "User", plural: "Users" },
  appName: "Frontier Admin",
};

export const defaultConfig: Config = {
  title: "Frontier Admin",
  app_url: "example.com",
  token_product_id: "token",
  organization_types: [],
  terminology: defaultTerminology,
};
