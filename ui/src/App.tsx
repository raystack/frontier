import { Flex } from "@raystack/apsara/v1";
import "@raystack/apsara/style.css";
import { Outlet } from "react-router-dom";
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
