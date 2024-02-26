import "@raystack/apsara/index.css";
import { MagicLinkVerify, useFrontier } from "@raystack/frontier/react";
import * as R from "ramda";
import { memo, useEffect, useState } from "react";
import { Route, Routes } from "react-router-dom";

import LoadingState from "./components/states/Loading";
import UnauthorizedState from "./components/states/Unauthorized";

import App from "./App";
import PlanList from "./containers/billingplans.list";
import PlanDetails from "./containers/billingplans.list/details";
import Dashboard from "./containers/dashboard";
import NewGroup from "./containers/groups.create";
import Groups from "./containers/groups.list";
import GroupDetails from "./containers/groups.list/details";
import Login from "./containers/login";
import MagicLink from "./containers/magiclink";
import NewOrganisation from "./containers/organisations.create";
import Organisations from "./containers/organisations.list";
import OrganisationBillingAccounts from "./containers/organisations.list/billingaccounts";
import BillingAccountDetails from "./containers/organisations.list/billingaccounts/details";
import OrganisationBAInvoices from "./containers/organisations.list/billingaccounts/invoices";
import OrganisationBASubscriptions from "./containers/organisations.list/billingaccounts/subscriptions";
import OrganisationDetails from "./containers/organisations.list/details";
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

export default memo(() => {
  const { client, user, isUserLoading } = useFrontier();
  const [isOrgListLoading, setIsOrgListLoading] = useState(false);
  const [isAdmin, setIsAdmin] = useState(false);

  const isUserEmpty = R.either(R.isEmpty, R.isNil)(user);

  useEffect(() => {
    async function getOrganizations() {
      setIsOrgListLoading(true);
      try {
        const resp = await client?.adminServiceListAllOrganizations();
        if (resp?.data?.organizations) {
          setIsAdmin(true);
        }
      } catch (error) {
        setIsAdmin(false);
      } finally {
        setIsOrgListLoading(false);
      }
    }

    if (!isUserEmpty) {
      getOrganizations();
    }
  }, [client, isUserEmpty]);

  const isLoading = isOrgListLoading || isUserLoading;

  return isLoading ? (
    <LoadingState />
  ) : isUserEmpty ? (
    <Routes>
      <Route path="/" element={<Login />}>
        <Route path="*" element={<div>No match</div>} />
      </Route>
      <Route path="/magiclink-verify" element={<MagicLink />} />
    </Routes>
  ) : isAdmin ? (
    <Routes>
      <Route path="/" element={<App />}>
        <Route index element={<Organisations />} />
        <Route path="dashboard" element={<Dashboard />} />
        <Route path="magiclink-verify" element={<MagicLinkVerify />} />

        <Route path="organisations" element={<Organisations />}>
          <Route path="create" element={<NewOrganisation />} />
        </Route>
        <Route
          path="organisations/:organisationId"
          element={<OrganisationDetails />}
        />
        <Route
          path="organisations/:organisationId/users"
          element={<OrganisationUsers />}
        />
        <Route
          path="organisations/:organisationId/projects"
          element={<OrganisationProjects />}
        />
        <Route
          path="organisations/:organisationId/serviceusers"
          element={<OrganisationServiceUsers />}
        />
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

        <Route path="*" element={<div>No match</div>} />
      </Route>
    </Routes>
  ) : (
    <UnauthorizedState />
  );
});
