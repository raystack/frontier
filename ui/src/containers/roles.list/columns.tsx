import { V1Beta1Role } from "@raystack/frontier";
import type { ColumnDef } from "@tanstack/react-table";
import { createColumnHelper } from "@tanstack/react-table";
import Skeleton from "react-loading-skeleton";
import { Link } from "react-router-dom";

const columnHelper = createColumnHelper<V1Beta1Role>();

interface getColumnsOptions {
  isLoading: boolean;
}

export const getColumns: (
  options: getColumnsOptions
) => ColumnDef<V1Beta1Role, any>[] = ({ isLoading }) => {
  return [
    columnHelper.accessor("id", {
      header: "ID",
      //@ts-ignore
      filterVariant: "text",
      cell: isLoading
        ? () => <Skeleton />
        : ({ row, getValue }) => {
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
      cell: isLoading ? () => <Skeleton /> : (info) => info.getValue(),
    },
    {
      header: "Types",
      accessorKey: "types",
      filterVariant: "text",
      cell: isLoading ? () => <Skeleton /> : (info) => info.getValue(),
      footer: (props) => props.column.id,
    },
  ];
};
