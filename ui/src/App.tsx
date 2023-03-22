import { Outlet } from "react-router-dom";
import Layout from "./components/Layout";
import Sidebar from "./components/sidebar/Sidebar";
import { Box } from "./primitives/box/Box";
import { Flex } from "./primitives/flex/Flex";

function App() {
  return (
    <Layout
      sidebar={
        <Sidebar>
          <Flex>
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
