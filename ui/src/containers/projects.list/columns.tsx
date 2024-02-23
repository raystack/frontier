import { V1Beta1Project } from "@raystack/frontier";
import type { ColumnDef } from "@tanstack/react-table";
import { createColumnHelper } from "@tanstack/react-table";
import Skeleton from "react-loading-skeleton";
import { Link } from "react-router-dom";

const columnHelper = createColumnHelper<V1Beta1Project>();

interface getColumnsOptions {
  isLoading: boolean;
}

export const getColumns: (
  opt: getColumnsOptions
) => ColumnDef<V1Beta1Project, any>[] = ({ isLoading }) => {
  return [
    columnHelper.accessor("id", {
      header: "ID",
      //@ts-ignore
      filterVariant: "text",
      cell: isLoading
        ? () => <Skeleton />
        : ({ row, getValue }) => {
            return (
              <Link to={`/projects/${row.getValue("id")}`}>{getValue()}</Link>
            );
          },
    }),

    {
      header: "Title",
      accessorKey: "title",
      cell: isLoading ? () => <Skeleton /> : (info) => info.getValue(),
      filterVariant: "text",
    },
    {
      header: "Organization Id",
      accessorKey: "org_id",
      cell: isLoading ? () => <Skeleton /> : (info) => info.getValue(),
      filterVariant: "text",
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
