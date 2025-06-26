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
import CreateOrUpdateProduct from "./containers/products.create";
import EditProduct from "./containers/products.edit";
import ProductList from "./containers/products.list";
import ProductDetails from "./containers/products.list/details";
import ProductPrices from "./containers/products.list/prices";

import Roles from "./containers/roles.list";
import RoleDetails from "./containers/roles.list/details";
import NewUser from "./containers/users.create";

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
import { UserDetailsAuditLogPage } from "./pages/users/details/audit-log";

import { InvoicesList } from "./pages/invoices/list";

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
          element={<OrganizationDetails />}
        >
          <Route index element={<Navigate to="members" />} />
          <Route path="members" element={<OrganizationMembersPage />} />
          <Route path="security" element={<OrganizationSecurity />} />
          <Route path="projects" element={<OrganizationProjectssPage />} />
          <Route path="invoices" element={<OrganizationInvoicesPage />} />
          <Route path="tokens" element={<OrganizationTokensPage />} />
          <Route path="apis" element={<OrganizationApisPage />} />
        </Route>
        <Route path="users" element={<UsersList />}>
          <Route path="create" element={<NewUser />} />
        </Route>
        <Route path="users/:userId" element={<UserDetails />}>
          <Route index element={<Navigate to="audit-log" />} />
          <Route path="audit-log" element={<UserDetailsAuditLogPage />} />
          <Route path="security" element={<UserDetailsSecurityPage />} />
        </Route>

        <Route path="plans" element={<PlanList />}>
          <Route path=":planId" element={<PlanDetails />} />
        </Route>

        <Route path="roles" element={<Roles />}>
          <Route path=":roleId" element={<RoleDetails />} />
        </Route>
        <Route path="products" element={<ProductList />}>
          <Route path="create" element={<CreateOrUpdateProduct />} />
          <Route path=":productId" element={<ProductDetails />} />
          <Route path=":productId/edit" element={<EditProduct />} />
        </Route>

        <Route path="products/:productId/prices" element={<ProductPrices />} />

        <Route path="preferences" element={<PreferencesLayout />}>
          <Route path="" element={<PreferencesList />} />
          <Route path=":name" element={<PreferenceDetails />} />
        </Route>

        <Route path="invoices" element={<InvoicesList />} />
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
