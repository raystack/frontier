import { V1Beta1Organization } from "@raystack/frontier";
import type { ColumnDef } from "@tanstack/react-table";
import { createColumnHelper } from "@tanstack/react-table";
import Skeleton from "react-loading-skeleton";
import { Link } from "react-router-dom";

const columnHelper = createColumnHelper<V1Beta1Organization>();

interface getColumnsOptions {
  isLoading: boolean;
}

export const getColumns: (
  options: getColumnsOptions
) => ColumnDef<V1Beta1Organization, any>[] = ({ isLoading }) => {
  return [
    columnHelper.accessor("id", {
      header: "ID",
      //@ts-ignore
      filterVariant: "text",
      cell: isLoading
        ? () => <Skeleton />
        : ({ row, getValue }) => {
            return (
              <Link to={`/organisations/${row.getValue("id")}`}>
                {getValue()}
              </Link>
            );
          },
    }),
    {
      header: "Title",
      accessorKey: "title",
      filterVariant: "text",
      cell: isLoading ? () => <Skeleton /> : (info) => info.getValue(),
      footer: (props) => props.column.id,
    },
    {
      header: "Status",
      accessorKey: "state",
      meta: {
        data: [
          { label: "Enabled", value: "enabled" },
          { label: "Disabled", value: "disabled" },
        ],
      },
      cell: isLoading ? () => <Skeleton /> : (info) => info.getValue(),
      footer: (props) => props.column.id,
      filterFn: (row, id, value) => {
        return value.includes(row.getValue(id));
      },
    },

    {
      header: "Create At",
      accessorKey: "created_at",

      meta: {
        headerFilter: false,
      },
      cell: isLoading
        ? () => <Skeleton />
        : (info) =>
            new Date(info.getValue() as Date).toLocaleString("en", {
              month: "long",
              day: "numeric",
              year: "numeric",
            }),

      footer: (props) => props.column.id,
    },
  ];
};
