"use client";

// views exports
export { default as RolesView } from "./views/roles";
export { default as InvoicesView } from "./views/invoices";
export { ProductsView, ProductPricesView } from "./views/products/exports";
export { default as AuditLogsView } from "./views/audit-logs";
export { default as AdminsView } from "./views/admins";
export { default as PlansView } from "./views/plans";
export { default as WebhooksView } from "./views/webhooks/webhooks";
export { default as PreferencesView } from "./views/preferences/PreferencesView";
export { default as UsersView } from "./views/users/UsersView";
export { OrganizationListView, type OrganizationListViewProps } from "./views/organizations/list";
export {
  OrganizationDetailsView,
  type OrganizationDetailsViewProps,
} from "./views/organizations/details";
export { OrganizationSecurity } from "./views/organizations/details/security";
export { OrganizationMembersView } from "./views/organizations/details/members";
export { OrganizationProjectsView } from "./views/organizations/details/projects";
export { OrganizationInvoicesView } from "./views/organizations/details/invoices";
export { OrganizationTokensView } from "./views/organizations/details/tokens";
export { OrganizationApisView } from "./views/organizations/details/apis";

// context exports
export {
  AdminConfigProvider,
  useAdminConfig,
  type AdminConfigProviderProps,
} from "./contexts/AdminConfigContext";

// hook exports
export { useTerminology } from "./hooks/useTerminology";
export { useAdminPaths, type AdminPaths } from "./hooks/useAdminPaths";

// utils exports
export {
  getConnectNextPageParam,
  DEFAULT_PAGE_SIZE,
  getGroupCountMapFromFirstPage,
  type ConnectRPCPaginatedResponse,
} from "./utils/connect-pagination";
export {
  transformDataTableQueryToRQLRequest,
  type TransformOptions,
} from "./utils/transform-query";
export {
  type Config,
  type AdminTerminologyConfig,
  defaultConfig,
  defaultTerminology,
} from "./utils/constants";
