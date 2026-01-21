import { Flex } from "@raystack/apsara";
import { Outlet } from "react-router-dom";
import "@raystack/apsara/style.css";
import "@raystack/apsara/normalize.css";
import "./App.css";
import IAMSidebar from "./components/Sidebar";

function App() {
  return (
    <div style={{ display: "flex", height: "100vh", overflow: "hidden" }}>
      <IAMSidebar />
      <Flex style={{ flexGrow: "1", overflow: "auto" }}>
        <Outlet />
      </Flex>
    </div>
  );
}

export default App;
