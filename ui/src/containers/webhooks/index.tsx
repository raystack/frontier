import { DataTable, Flex } from "@raystack/apsara";

export function WebhooksList() {
  const tableStyle = { width: "100%" };

  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={[]}
        columns={[]}
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
