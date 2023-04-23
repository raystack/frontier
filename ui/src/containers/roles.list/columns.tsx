import type { ColumnDef } from "@tanstack/react-table";
import { createColumnHelper } from "@tanstack/react-table";
import { Link } from "react-router-dom";
import { Role } from "~/types/role";

const columnHelper = createColumnHelper<Role>();
export const getColumns: (Roles: Role[]) => ColumnDef<Role, any>[] = (
  Roles: Role[]
) => {
  return [
    columnHelper.accessor("id", {
      header: "ID",
      //@ts-ignore
      filterVariant: "text",
      cell: ({ row, getValue }) => {
        return (
          <Link to={`${encodeURIComponent(row.getValue("id"))}`}>
            {getValue()}
          </Link>
        );
      },
    }),
    {
      header: "Name",
      accessorKey: "name",
      filterVariant: "text",
      cell: (info) => info.getValue(),
    },
    {
      header: "Types",
      accessorKey: "types",
      filterVariant: "text",
      cell: (info) => info.getValue(),
      footer: (props) => props.column.id,
    },
  ];
};
