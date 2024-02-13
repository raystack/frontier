import "@raystack/apsara/index.css";
import { MagicLinkVerify, useFrontier } from "@raystack/frontier/react";
import { memo, useEffect, useState } from "react";
import { Route, Routes } from "react-router-dom";
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
  const { client, user } = useFrontier();
  const [isAdmin, setIsAdmin] = useState(false);

  useEffect(() => {
    try {
      async function getOrganizations() {
        await client?.adminServiceListAllOrganizations();
        setIsAdmin(true);
      }
      getOrganizations();
    } catch (error) {
      setIsAdmin(false);
    }
  }, []);

  return user && isAdmin ? (
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

        <Route path="*" element={<div>No match</div>} />
      </Route>
    </Routes>
  ) : (
    <Routes>
      <Route path="/" element={<Login />}>
        <Route path="*" element={<div>No match</div>} />
      </Route>
      <Route path="/magiclink-verify" element={<MagicLink />} />
    </Routes>
  );
});
