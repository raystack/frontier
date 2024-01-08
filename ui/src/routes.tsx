import "@raystack/apsara/index.css";
import { useFrontier } from "@raystack/frontier/react";
import { Route, Routes } from "react-router-dom";
import App from "./App";
import Dashboard from "./containers/dashboard";
import NewGroup from "./containers/groups.create";
import Groups from "./containers/groups.list";
import GroupDetails from "./containers/groups.list/details";
import Login from "./containers/login";
import NewOrganisation from "./containers/organisations.create";
import Organisations from "./containers/organisations.list";
import OrganisationDetails from "./containers/organisations.list/details";
import NewProject from "./containers/projects.create";
import Projects from "./containers/projects.list";
import ProjectDetails from "./containers/projects.list/details";
import Roles from "./containers/roles.list";
import RoleDetails from "./containers/roles.list/details";
import NewUser from "./containers/users.create";
import Users from "./containers/users.list";
import UserDetails from "./containers/users.list/details";

export default () => {
  const { user } = useFrontier();
  return user ? (
    <Routes>
      <Route path="/console" element={<App />}>
        <Route index element={<Organisations />} />
        <Route path="dashboard" element={<Dashboard />} />
        <Route path="organisations" element={<Organisations />}>
          <Route path="create" element={<NewOrganisation />} />
          <Route path=":organisationId" element={<OrganisationDetails />} />
        </Route>
        <Route path="projects" element={<Projects />}>
          <Route path="create" element={<NewProject />} />
          <Route path=":projectId" element={<ProjectDetails />} />
        </Route>
        <Route path="users" element={<Users />}>
          <Route path="create" element={<NewUser />} />
          <Route path=":userId" element={<UserDetails />} />
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
    <Login />
  );
};
