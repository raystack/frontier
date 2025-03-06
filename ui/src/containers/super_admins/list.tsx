import { DataTable } from "@raystack/apsara";
import { Flex } from "@raystack/apsara/v1";
import { useContext } from "react";
import { AppContext } from "~/contexts/App";
import { getColumns } from "./columns";

export function SuperAdminList() {
  const { platformUsers } = useContext(AppContext);

  const tableStyle = { width: "100%" };
  const columns = getColumns();
  const data = [
    ...(platformUsers?.users || []),
    ...(platformUsers?.serviceusers || []),
  ];
  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={data || []}
        // @ts-ignore
        columns={columns}
        parentStyle={{ height: "calc(100vh - 60px)" }}
        style={tableStyle}
      >
        <DataTable.Toolbar>
          <DataTable.FilterChips style={{ padding: "8px 24px" }} />
        </DataTable.Toolbar>
      </DataTable>
    </Flex>
  );
}
