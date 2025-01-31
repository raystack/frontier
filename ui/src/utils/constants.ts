export const DEFAULT_DATE_FORMAT = "MMM DD, YYYY";

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
}

export const defaultConfig: Config = {
  title: "Frontier Admin",
};
