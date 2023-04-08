import type { ColumnDef } from "@tanstack/react-table";
import { createColumnHelper } from "@tanstack/react-table";
import { Link } from "react-router-dom";
import type { Group } from "~/types/group";

const columnHelper = createColumnHelper<Group>();
export const columns: ColumnDef<Group, any>[] = [
  columnHelper.accessor("id", {
    header: "ID",
    cell: ({ row, getValue }) => {
      return <Link to={`${row.getValue("id")}`}>{getValue()}</Link>;
    },
  }),
  {
    header: "Name",
    accessorKey: "name",
    cell: (info) => info.getValue(),
  },
  {
    header: "Create At",
    accessorKey: "createdAt",
    cell: (info) =>
      new Date(info.getValue() as Date).toLocaleString("en", {
        month: "long",
        day: "numeric",
        year: "numeric",
      }),

    footer: (props) => props.column.id,
  },
];
