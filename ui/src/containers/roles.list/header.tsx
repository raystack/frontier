import { DataTable, Flex, Text, useTable } from "@raystack/apsara";

export const RolesHeader = () => {
  const { filteredColumns, table } = useTable();
  const isFiltered = filteredColumns.length > 0;

  return (
    <>
      <Flex align="center" justify="between">
        <Text style={{ fontSize: "14px", fontWeight: "500" }}>Roles</Text>
        <Flex gap="small">
          {isFiltered ? <DataTable.ClearFilter /> : <DataTable.FilterOptions />}
          <DataTable.ViewOptions />
          <DataTable.GloabalSearch placeholder="Search roles..." />
        </Flex>
      </Flex>
    </>
  );
};
