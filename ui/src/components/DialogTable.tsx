import React from "react";
import { DataTable } from "@raystack/apsara";
import { ColumnDef } from "@tanstack/table-core";
import { Flex } from "@raystack/apsara/v1";

type DialogTableProps = {
  columns: ColumnDef<any, any>[];
  data: any[];
  header?: React.ReactNode;
};
export default function DialogTable({
  columns,
  data,
  header,
}: DialogTableProps) {
  return (
    <Flex
      direction="row"
      style={{ height: "100%", width: "100%", minWidth: "1020px" }}
    >
      <DataTable
        data={data ?? []}
        // @ts-ignore
        columns={columns}
        style={{ width: "100%" }}
      >
        {header && <DataTable.Toolbar>{header}</DataTable.Toolbar>}
      </DataTable>
    </Flex>
  );
}
