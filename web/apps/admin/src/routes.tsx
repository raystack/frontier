// import { MagicLinkVerify } from "@raystack/frontier/react";
import * as R from "ramda";
import { memo, useContext } from "react";
import { Navigate, Route, Routes } from "react-router-dom";

import LoadingState from "./components/states/Loading";
import UnauthorizedState from "./components/states/Unauthorized";

import App from "./App";
import PlanList from "./containers/billingplans.list";
import PlanDetails from "./containers/billingplans.list/details";
import Login from "./containers/login";
import MagicLink from "./containers/magiclink";

import PreferencesList from "./containers/preferences.list";
import PreferenceDetails from "./containers/preferences.list/details";
import PreferencesLayout from "./containers/preferences.list/layout";
import { ProductsPage } from "./pages/products/ProductsPage";
import { ProductPricesPage } from "./pages/products/ProductPricesPage";

import { RolesPage } from "./pages/roles/RolesPage";

import { AppContext } from "./contexts/App";
import { SuperAdminList } from "./containers/super_admins/list";
import WebhooksList from "./containers/webhooks";
import CreateWebhooks from "./containers/webhooks/create";
import UpdateWebhooks from "./containers/webhooks/update";
import AuthLayout from "./layout/auth";

import { OrganizationList } from "./pages/organizations/list";
import { OrganizationDetails } from "./pages/organizations/details";
import { OrganizationSecurity } from "./pages/organizations/details/security";
import { OrganizationMembersPage } from "./pages/organizations/details/members";
import { OrganizationProjectssPage } from "./pages/organizations/details/projects";
import { OrganizationInvoicesPage } from "./pages/organizations/details/invoices";
import { OrganizationTokensPage } from "./pages/organizations/details/tokens";
import { OrganizationApisPage } from "./pages/organizations/details/apis";

import { UsersList } from "./pages/users/list";
import { UserDetails } from "./pages/users/details";
import { UserDetailsSecurityPage } from "./pages/users/details/security";

import { InvoicesPage } from "./pages/invoices/InvoicesPage";
import { AuditLogsList } from "./pages/audit-logs/list";

export default memo(function AppRoutes() {
  const { isAdmin, isLoading, user } = useContext(AppContext);

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
        <Route index element={<Navigate to="/organizations" />} />
        <Route path="organizations" element={<OrganizationList />} />
        <Route
          path="organizations/:organizationId"
          element={<OrganizationDetails />}>
          <Route index element={<Navigate to="members" />} />
          <Route path="members" element={<OrganizationMembersPage />} />
          <Route path="security" element={<OrganizationSecurity />} />
          <Route path="projects" element={<OrganizationProjectssPage />} />
          <Route path="invoices" element={<OrganizationInvoicesPage />} />
          <Route path="tokens" element={<OrganizationTokensPage />} />
          <Route path="apis" element={<OrganizationApisPage />} />
        </Route>
        <Route path="users" element={<UsersList />} />
        <Route path="users/:userId" element={<UserDetails />}>
          <Route index element={<Navigate to="security" />} />
          <Route path="security" element={<UserDetailsSecurityPage />} />
        </Route>

        <Route path="audit-logs" element={<AuditLogsList />} />

        <Route path="plans" element={<PlanList />}>
          <Route path=":planId" element={<PlanDetails />} />
        </Route>

        <Route path="roles" element={<RolesPage />}>
          <Route path=":roleId" element={<RolesPage />} />
        </Route>
        
        <Route path="products" element={<ProductsPage />}>
          <Route path=":productId" element={<ProductsPage />} />
        </Route>

        <Route path="products/:productId/prices" element={<ProductPricesPage />} />

        <Route path="preferences" element={<PreferencesLayout />}>
          <Route path="" element={<PreferencesList />} />
          <Route path=":name" element={<PreferenceDetails />} />
        </Route>

        <Route path="invoices" element={<InvoicesPage />} />
        <Route path="super-admins" element={<SuperAdminList />} />
        <Route path="webhooks" element={<WebhooksList />}>
          <Route path="create" element={<CreateWebhooks />} />
          <Route path=":webhookId" element={<UpdateWebhooks />} />
        </Route>
        <Route path="*" element={<Navigate to="/organizations" />} />
      </Route>
    </Routes>
  ) : (
    <UnauthorizedState />
  );
});
