export const DEFAULT_DATE_FORMAT = "MMM DD, YYYY";
export const DEFAULT_TOKEN_PRODUCT_NAME = "token";

export const PERMISSIONS = {
  OrganizationNamespace: "app/organization",
} as const;

export const SUBSCRIPTION_STATUSES = [
  { label: "Active", value: "active" },
  { label: "Trialing", value: "trialing" },
  { label: "Past due", value: "past_due" },
  { label: "Canceled", value: "canceled" },
  { label: "Ended", value: "ended" },
];

export interface Config {
  title: string;
  logo?: string;
  app_url?: string;
  token_product_id?: string;
}

export const defaultConfig: Config = {
  title: "Frontier Admin",
  app_url: "example.com",
  token_product_id: DEFAULT_TOKEN_PRODUCT_NAME,
};

export const NULL_DATE = "0001-01-01T00:00:00Z";
