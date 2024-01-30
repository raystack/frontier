import { V1Beta1Organization } from "@raystack/frontier";
import type { ColumnDef } from "@tanstack/react-table";
import { createColumnHelper } from "@tanstack/react-table";
import { Link } from "react-router-dom";

const columnHelper = createColumnHelper<V1Beta1Organization>();

export const getColumns: (
  organisations: V1Beta1Organization[]
) => ColumnDef<V1Beta1Organization, any>[] = (
  organisations: V1Beta1Organization[]
) => {
  return [
    columnHelper.accessor("id", {
      header: "ID",
      //@ts-ignore
      filterVariant: "text",
      cell: ({ row, getValue }) => {
        return (
          <Link to={`/organisations/${row.getValue("id")}`}>{getValue()}</Link>
        );
      },
    }),
    {
      header: "Title",
      accessorKey: "title",
      filterVariant: "text",
      cell: (info) => info.getValue(),
      footer: (props) => props.column.id,
    },
    {
      header: "Status",
      accessorKey: "state",
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
