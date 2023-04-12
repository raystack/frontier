import { ApsaraThemeProvider } from "@odpf/apsara";
import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import App from "./App";
import Dashboard from "./containers/dashboard";
import NewGroup from "./containers/groups.create";
import Groups from "./containers/groups.list";
import GroupDetails from "./containers/groups.list/details";
import Home from "./containers/home";
import NewOrganisation from "./containers/organisations.create";
import Organisations from "./containers/organisations.list";
import OrganisationDetails from "./containers/organisations.list/details";
import NewProject from "./containers/projects.create";
import Projects from "./containers/projects.list";
import ProjectDetails from "./containers/projects.list/details";
import NewUser from "./containers/users.create";
import Users from "./containers/users.list";
import UserDetails from "./containers/users.list/details";

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <BrowserRouter>
      <ApsaraThemeProvider>
        <Routes>
          <Route path="/" element={<App />}>
            <Route index element={<Home />} />
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

            {/* Using path="*"" means "match anything", so this route
          acts like a catch-all for URLs that we don't have explicit
          routes for. */}
            <Route path="*" element={<div>No match</div>} />
          </Route>
        </Routes>
      </ApsaraThemeProvider>
    </BrowserRouter>
  </React.StrictMode>
);
