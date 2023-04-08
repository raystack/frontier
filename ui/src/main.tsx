import { ApsaraThemeProvider } from "@odpf/apsara";
import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import App from "./App";
import Dashboard from "./containers/dashboard";
import Groups from "./containers/groups.list";
import Home from "./containers/home";
import Organisations from "./containers/organisations.list";
import Projects from "./containers/projects.list";
import Users from "./containers/users.list";

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <BrowserRouter>
      <ApsaraThemeProvider>
        <Routes>
          <Route path="/" element={<App />}>
            <Route index element={<Home />} />
            <Route path="dashboard" element={<Dashboard />} />
            <Route path="organisations" element={<Organisations />} />
            <Route path="projects" element={<Projects />} />
            <Route path="users" element={<Users />} />
            <Route path="groups" element={<Groups />} />

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
