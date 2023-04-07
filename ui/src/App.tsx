import { Box, Flex } from "@odpf/apsara";
import { Outlet } from "react-router-dom";
import "./App.css";
import Layout from "./components/Layout";
import Sidebar from "./components/sidebar/Sidebar";

function App() {
  return (
    <Layout
      sidebar={
        <Sidebar>
          <Flex direction="column" css={{ width: "100%" }}>
            <Sidebar.Header></Sidebar.Header>
            <Sidebar.Content></Sidebar.Content>
          </Flex>
          <Sidebar.Footer></Sidebar.Footer>
        </Sidebar>
      }
    >
      <Box css={{ padding: "$8" }}>
        <Outlet />
      </Box>
    </Layout>
  );
}

export default App;
