import type { ColumnDef } from "@tanstack/react-table";
import { createColumnHelper } from "@tanstack/react-table";
import { Link } from "react-router-dom";
import type { User } from "~/types/user";

const columnHelper = createColumnHelper<User>();
export const columns: ColumnDef<User, any>[] = [
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
    header: "Slug",
    accessorKey: "slug",
    cell: (info) => info.getValue(),
    footer: (props) => props.column.id,
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
