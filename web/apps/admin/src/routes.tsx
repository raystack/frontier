// import { MagicLinkVerify } from "@raystack/frontier/react";
import * as R from "ramda";
import { memo, useContext } from "react";
import { Navigate, Route, Routes } from "react-router-dom";

import LoadingState from "./components/states/Loading";
import UnauthorizedState from "./components/states/Unauthorized";

import App from "./App";
import { PlansPage } from "./pages/plans/PlansPage";
import Login from "./containers/login";
import MagicLink from "./containers/magiclink";

import { PreferencesPage } from "./pages/preferences/PreferencesPage";
import { ProductsPage } from "./pages/products/ProductsPage";
import { ProductPricesPage } from "./pages/products/ProductPricesPage";

import { RolesPage } from "./pages/roles/RolesPage";

import { AppContext } from "./contexts/App";
import { AdminsPage } from "./pages/admins/AdminsPage";
import { WebhooksPage } from "./pages/webhooks/WebhooksPage";
import AuthLayout from "./layout/auth";

import { OrganizationListPage } from "./pages/organizations/list";
import { OrganizationDetailsPage } from "./pages/organizations/details";
import {
  OrganizationSecurity,
  OrganizationMembersView,
  OrganizationProjectsView,
  OrganizationInvoicesView,
  OrganizationTokensView,
  OrganizationApisView,
  OrganizationPatView,
  useAdminPaths,
} from "@raystack/frontier/admin";

import { UsersPage } from "./pages/users/UsersPage";

import { InvoicesPage } from "./pages/invoices/InvoicesPage";
import { AuditLogsPage } from "./pages/audit-logs/AuditLogsPage";

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
