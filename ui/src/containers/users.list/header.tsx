import { PlusIcon } from "@radix-ui/react-icons";
import { Button, DataTable, Flex, Text, useTable } from "@raystack/apsara";
import { useNavigate } from "react-router-dom";

export const UsersHeader = () => {
  const navigate = useNavigate();
  const { filteredColumns, table } = useTable();
  const isFiltered = filteredColumns.length > 0;

  return (
    <Flex align="center" justify="between" style={{ padding: "16px 24px" }}>
      <Text style={{ fontSize: "14px", fontWeight: "500" }}>Users</Text>
      <Flex gap="small">
        {isFiltered ? <DataTable.ClearFilter /> : <DataTable.FilterOptions />}
        <DataTable.ViewOptions />
        <DataTable.GloabalSearch placeholder="Search users..." />
        <Button
          variant="secondary"
          onClick={() => navigate("/console/users/create")}
          style={{ width: "100%" }}
        >
          <Flex
            direction="column"
            align="center"
            style={{ paddingRight: "var(--pd-4)" }}
          >
            <PlusIcon />
          </Flex>
          new user
        </Button>
      </Flex>
    </Flex>
  );
};
