import { Flex, Table } from "@raystack/apsara";
import { ColumnDef } from "@tanstack/table-core";
import { tableStyle } from "~/styles";

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
      css={{ height: "100%", width: "100%", minWidth: "1020px" }}
    >
      <Table css={tableStyle} columns={columns} data={data ?? []}>
        {header && <Table.TopContainer>{header}</Table.TopContainer>}
      </Table>
    </Flex>
  );
}
