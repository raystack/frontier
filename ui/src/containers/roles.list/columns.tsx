import { V1Beta1Role } from "@raystack/frontier";
import type { ColumnDef } from "@tanstack/react-table";
import { createColumnHelper } from "@tanstack/react-table";
import { Link } from "react-router-dom";

const columnHelper = createColumnHelper<V1Beta1Role>();

export const getColumns: () => ColumnDef<V1Beta1Role, any>[] = () => {
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
