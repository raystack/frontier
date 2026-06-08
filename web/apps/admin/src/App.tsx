import { Flex } from "@raystack/apsara";
import { Suspense } from "react";
import { Outlet } from "react-router-dom";
import "@raystack/apsara/normalize.css";
import "@raystack/apsara/style.css";
import "./App.css";
import IAMSidebar from "./components/Sidebar";
import LoadingState from "./components/states/Loading";

function App() {
  return (
    <div style={{ display: "flex", height: "100vh", overflow: "hidden" }}>
      <IAMSidebar />
      <Flex style={{ flexGrow: "1", overflow: "auto" }}>
        {/* Boundary for the lazily-loaded route pages. Sits inside the layout
            so the sidebar stays mounted while a route chunk loads. */}
        <Suspense fallback={<LoadingState />}>
          <Outlet />
        </Suspense>
      </Flex>
    </div>
  );
}

export default App;
