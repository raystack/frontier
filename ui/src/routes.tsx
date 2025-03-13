import "@raystack/apsara/style.css";
import { MagicLinkVerify } from "@raystack/frontier/react";
import * as R from "ramda";
import { memo, useContext } from "react";
import { Route, Routes } from "react-router-dom";

import LoadingState from "./components/states/Loading";
import UnauthorizedState from "./components/states/Unauthorized";

import App from "./App";
import PlanList from "./containers/billingplans.list";
import PlanDetails from "./containers/billingplans.list/details";
// import Dashboard from "./containers/dashboard";
import NewGroup from "./containers/groups.create";
import Groups from "./containers/groups.list";
import GroupDetails from "./containers/groups.list/details";
import Login from "./containers/login";
import MagicLink from "./containers/magiclink";
import NewOrganisation from "./containers/organisations.create";
// import Organisations from "./containers/organisations.list";
import OrganisationBillingAccounts from "./containers/organisations.list/billingaccounts";
import BillingAccountDetails from "./containers/organisations.list/billingaccounts/details";
import OrganisationBAInvoices from "./containers/organisations.list/billingaccounts/invoices";
import OrganisationBASubscriptions from "./containers/organisations.list/billingaccounts/subscriptions";
// import OrganisationDetails from "./containers/organisations.list/details";
import OrganisationProjects from "./containers/organisations.list/projects";
import OrganisationServiceUsers from "./containers/organisations.list/serviceusers";
import OrgSettingPage from "./containers/organisations.list/settings";
import OrganisationUsers from "./containers/organisations.list/users";
import PreferencesList from "./containers/preferences.list";
import PreferenceDetails from "./containers/preferences.list/details";
import PreferencesLayout from "./containers/preferences.list/layout";
import CreateOrUpdateProduct from "./containers/products.create";
import EditProduct from "./containers/products.edit";
import ProductList from "./containers/products.list";
import ProductDetails from "./containers/products.list/details";
import ProductPrices from "./containers/products.list/prices";
import NewProject from "./containers/projects.create";
import Projects from "./containers/projects.list";
import ProjectDetails from "./containers/projects.list/details";
import ProjectUsers from "./containers/projects.list/users";
import Roles from "./containers/roles.list";
import RoleDetails from "./containers/roles.list/details";
import NewUser from "./containers/users.create";
import Users from "./containers/users.list";
import UserDetails from "./containers/users.list/details";
import InvoicesList from "./containers/invoices.list";
import { AppContext } from "./contexts/App";
import NewServiceUsers from "./containers/organisations.list/serviceusers/create";
import OrganisationTokens from "./containers/organisations.list/billingaccounts/tokens";
import AddTokens from "./containers/organisations.list/billingaccounts/tokens/add";
import ServiceUserDetails from "./containers/organisations.list/serviceusers/details";
import AddServiceUserToken from "./containers/organisations.list/serviceusers/tokens/add";
import { SuperAdminList } from "./containers/super_admins/list";
import InviteUsers from "./containers/organisations.list/users/invite";
import WebhooksList from "./containers/webhooks";
import CreateWebhooks from "./containers/webhooks/create";
import UpdateWebhooks from "./containers/webhooks/update";
import AuthLayout from "./layout/auth";

import { OrganizationList } from "./pages/organizations/list";
import { OrganizationDetails } from "./pages/organizations/details";

export default memo(function AppRoutes() {
  const { isAdmin, isLoading, user } = useContext(AppContext);

  const isUserEmpty = R.either(R.isEmpty, R.isNil)(user);

  return isLoading ? (
    // The global full page loading state is causing issues with infinite scroll. Remove this in future
    <LoadingState />
  ) : isUserEmpty ? (
    <Routes>
      <Route element={<AuthLayout />}>
        <Route path="/" element={<Login />}></Route>
        <Route path="/magiclink-verify" element={<MagicLink />} />
        <Route path="*" element={<div>No match</div>} />
      </Route>
    </Routes>
  ) : isAdmin ? (
    <Routes>
      <Route path="/" element={<App />}>
        <Route index element={<OrganizationList />} />
        <Route path="organisations" element={<OrganizationList />}>
          <Route path="create" element={<NewOrganisation />} />
        </Route>
        <Route
          path="organisations/:organisationId"
          element={<OrganizationDetails />}
        />
        <Route
          path="organisations/:organisationId/users"
          element={<OrganisationUsers />}
        >
          <Route path="invite" element={<InviteUsers />} />
        </Route>
        <Route
          path="organisations/:organisationId/projects"
          element={<OrganisationProjects />}
        />
        <Route
          path="organisations/:organisationId/serviceusers"
          element={<OrganisationServiceUsers />}
        >
          <Route path="create" element={<NewServiceUsers />} />
        </Route>
        <Route
          path="organisations/:organisationId/serviceusers/:serviceUserId"
          element={<ServiceUserDetails />}
        >
          <Route path="create-token" element={<AddServiceUserToken />} />
        </Route>
        <Route
          path="organisations/:organisationId/billingaccounts"
          element={<OrganisationBillingAccounts />}
        >
          <Route path=":billingaccountId" element={<BillingAccountDetails />} />
        </Route>
        <Route
          path="organisations/:organisationId/billingaccounts/:billingaccountId/subscriptions"
          element={<OrganisationBASubscriptions />}
        />
        <Route
          path="organisations/:organisationId/billingaccounts/:billingaccountId/invoices"
          element={<OrganisationBAInvoices />}
        />
        <Route
          path="organisations/:organisationId/billingaccounts/:billingaccountId/tokens"
          element={<OrganisationTokens />}
        >
          <Route path="add" element={<AddTokens />} />
        </Route>
        <Route
          path="organisations/:organisationId/settings"
          element={<OrgSettingPage />}
        ></Route>

        <Route path="projects" element={<Projects />}>
          <Route path="create" element={<NewProject />} />
        </Route>
        <Route path="projects/:projectId" element={<ProjectDetails />} />
        <Route path="projects/:projectId/users" element={<ProjectUsers />} />

        <Route path="users" element={<Users />}>
          <Route path="create" element={<NewUser />} />
        </Route>
        <Route path="users/:userId" element={<UserDetails />} />

        <Route path="plans" element={<PlanList />}>
          <Route path=":planId" element={<PlanDetails />} />
        </Route>

        <Route path="groups" element={<Groups />}>
          <Route path="create" element={<NewGroup />} />
          <Route path=":groupId" element={<GroupDetails />} />
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
        <Route path="*" element={<div>No match</div>} />
      </Route>
    </Routes>
  ) : (
    <UnauthorizedState />
  );
});
