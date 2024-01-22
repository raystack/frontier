import type { ColumnDef } from "@tanstack/react-table";
import { createColumnHelper } from "@tanstack/react-table";
import { Link } from "react-router-dom";
import type { V1Beta1Organization } from "@raystack/frontier";
import Skeleton from "react-loading-skeleton";

const columnHelper = createColumnHelper<V1Beta1Organization>();

interface getColumnsOptions {
  organizations: V1Beta1Organization[];
  isLoading?: boolean;
}

export const getColumns: (
  options: getColumnsOptions
) => ColumnDef<V1Beta1Organization, any>[] = ({ organizations, isLoading }) => {
  return [
    {
      header: "Title",
      accessorKey: "title",
      filterVariant: "text",
      cell: isLoading
        ? () => <Skeleton />
        : ({ row, getValue }) => {
            return (
              <Link to={`/organisations/${row.original?.id}`}>
                {getValue()}
              </Link>
            );
          },
      footer: (props) => props.column.id,
    },
    {
      header: "Name",
      accessorKey: "name",
      filterVariant: "text",
      cell: isLoading ? () => <Skeleton /> : (info) => info.getValue(),
      footer: (props) => props.column.id,
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
