import type { ColumnDef } from "@tanstack/react-table";
import { createColumnHelper } from "@tanstack/react-table";
import { Link } from "react-router-dom";
import type { Organisation } from "~/types/organisation";

const columnHelper = createColumnHelper<Organisation>();

export const getColumns: (
  organisations: Organisation[]
) => ColumnDef<Organisation, any>[] = (organisations: Organisation[]) => {
  return [
    columnHelper.accessor("id", {
      header: "ID",
      //@ts-ignore
      filterVariant: "text",
      cell: ({ row, getValue }) => {
        return (
          <Link to={`/console/organisations/${row.getValue("id")}`}>
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
      footer: (props) => props.column.id,
    },

    {
      header: "Create At",
      accessorKey: "created_at",

      meta: {
        headerFilter: false,
      },
      cell: (info) =>
        new Date(info.getValue() as Date).toLocaleString("en", {
          month: "long",
          day: "numeric",
          year: "numeric",
        }),

      footer: (props) => props.column.id,
    },
  ];
};
