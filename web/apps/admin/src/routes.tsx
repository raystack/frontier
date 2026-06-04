import * as R from "ramda";
import { lazy, memo, useContext } from "react";
import { Navigate, Route, Routes } from "react-router-dom";

import LoadingState from "./components/states/Loading";
import UnauthorizedState from "./components/states/Unauthorized";

// Eager: the app shell and the unauthenticated flow. Keeping these in the
// initial bundle means the login screen paints instantly.
import App from "./App";
import Login from "./containers/login";
import MagicLink from "./containers/magiclink";
import AuthLayout from "./layout/auth";

import { AppContext } from "./contexts/App";
// Static imports from the SDK barrel. `useAdminPaths` is a hook used during
// render. The organization-detail child views are imported eagerly too: they
// come from the `@raystack/frontier/admin` barrel, which is already pinned into
// the initial chunk (the app shell imports from it), so wrapping them in
// React.lazy produced no separate chunk — only Suspense churn. Lazy loading is
// kept on the app-owned page files below, which DO split into their own chunks.
import {
  useAdminPaths,
  OrganizationSecurity,
  OrganizationMembersView,
  OrganizationProjectsView,
  OrganizationInvoicesView,
  OrganizationTokensView,
  OrganizationApisView,
  OrganizationPatView,
} from "@raystack/frontier/admin";

// Lazily-loaded route pages — each becomes its own async chunk, so the heavy
// admin code stays out of the initial/unauthenticated bundle. The pages are
// default exports, so React.lazy imports them directly (per the React docs).
const PlansPage = lazy(() => import("./pages/plans/PlansPage"));
const PreferencesPage = lazy(() => import("./pages/preferences/PreferencesPage"));
const ProductsPage = lazy(() => import("./pages/products/ProductsPage"));
const ProductPricesPage = lazy(() => import("./pages/products/ProductPricesPage"));
const RolesPage = lazy(() => import("./pages/roles/RolesPage"));
const AdminsPage = lazy(() => import("./pages/admins/AdminsPage"));
const WebhooksPage = lazy(() => import("./pages/webhooks/WebhooksPage"));
const OrganizationListPage = lazy(() => import("./pages/organizations/list"));
const OrganizationDetailsPage = lazy(() => import("./pages/organizations/details"));
const UsersPage = lazy(() => import("./pages/users/UsersPage"));
const InvoicesPage = lazy(() => import("./pages/invoices/InvoicesPage"));
const AuditLogsPage = lazy(() => import("./pages/audit-logs/AuditLogsPage"));

export default memo(function AppRoutes() {
  const { isAdmin, isLoading, user } = useContext(AppContext);
  const paths = useAdminPaths();

  const isUserEmpty = R.either(R.isEmpty, R.isNil)(user);

  return isLoading ? (
    // The global full page loading state is causing issues with infinite scroll. Remove this in future
    <LoadingState />
  ) : isUserEmpty ? (
    <Routes>
      <Route element={<AuthLayout />}>
        <Route index element={<Navigate to="/login" />} />
        <Route path="/login" element={<Login />} />
        <Route path="/magiclink-verify" element={<MagicLink />} />
        <Route path="*" element={<Navigate to="/login" />} />
      </Route>
    </Routes>
  ) : isAdmin ? (
    <Routes>
      <Route path="/" element={<App />}>
        <Route index element={<Navigate to={`/${paths.organizations}`} />} />
        <Route path={paths.organizations} element={<OrganizationListPage />} />
        <Route
          path={`${paths.organizations}/:organizationId`}
          element={<OrganizationDetailsPage />}>
          <Route index element={<Navigate to={paths.members} replace />} />
          <Route path={paths.members} element={<OrganizationMembersView />} />
          <Route path="security" element={<OrganizationSecurity />} />
          <Route path={paths.projects} element={<OrganizationProjectsView />} />
          <Route path="invoices" element={<OrganizationInvoicesView />} />
          <Route path="tokens" element={<OrganizationTokensView />} />
          <Route path="apis" element={<OrganizationApisView />} />
          <Route path="pat" element={<OrganizationPatView />} />
        </Route>
        <Route path={paths.users} element={<UsersPage />}>
          <Route path=":userId" element={<UsersPage />} />
          <Route path=":userId/security" element={<UsersPage />} />
        </Route>

        <Route path="audit-logs" element={<AuditLogsPage />} />

        <Route path="plans" element={<PlansPage />}>
          <Route path=":planId" element={<PlansPage />} />
        </Route>

        <Route path="roles" element={<RolesPage />}>
          <Route path=":roleId" element={<RolesPage />} />
        </Route>

        <Route path="products" element={<ProductsPage />}>
          <Route path=":productId" element={<ProductsPage />} />
        </Route>

        <Route path="products/:productId/prices" element={<ProductPricesPage />} />

        <Route path="preferences" element={<PreferencesPage />}>
          <Route path=":name" element={<PreferencesPage />} />
        </Route>

        <Route path="invoices" element={<InvoicesPage />} />
        <Route path="super-admins" element={<AdminsPage />} />
        <Route path="webhooks" element={<WebhooksPage />}>
          <Route path="create" element={<WebhooksPage />} />
          <Route path=":webhookId" element={<WebhooksPage />} />
        </Route>
        <Route path="*" element={<Navigate to={`/${paths.organizations}`} />} />
      </Route>
    </Routes>
  ) : (
    <UnauthorizedState />
  );
});
