import { V1Beta1Group } from "@raystack/frontier";
import type { ColumnDef } from "@tanstack/react-table";
import { createColumnHelper } from "@tanstack/react-table";
import { compose, map, pathOr, uniq } from "ramda";
import Skeleton from "react-loading-skeleton";
import { Link } from "react-router-dom";

import { keyToColumnMetaObject } from "~/utils/helper";

const columnHelper = createColumnHelper<V1Beta1Group>();

interface getColumnsOptions {
  isLoading: boolean;
  groups: V1Beta1Group[];
}

export const getColumns: (
  opt: getColumnsOptions
) => ColumnDef<V1Beta1Group, any>[] = ({ groups, isLoading }) => {
  // if (!groups) return [];

  // const uniqueValues = (propsValue: string) =>
  //   compose(
  //     map(keyToColumnMetaObject),
  //     uniq,
  //     map(pathOr([], propsValue.split(".")))
  //   )(groups);

  // const uniqueNames = uniqueValues("name");

  return [
    columnHelper.accessor("id", {
      header: "ID",
      //@ts-ignore
      filterVariant: "text",
      cell: isLoading
        ? () => <Skeleton />
        : ({ row, getValue }) => {
            return <Link to={`${row.getValue("id")}`}>{getValue()}</Link>;
          },
    }),
    {
      header: "Title",
      accessorKey: "title",
      cell: isLoading ? () => <Skeleton /> : (info: any) => info.getValue(),
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
      filterVariant: "text",
      meta: {
        headerFilter: false,
      },
      cell: isLoading
        ? () => <Skeleton />
        : (info: any) =>
            new Date(info.getValue() as Date).toLocaleString("en", {
              month: "long",
              day: "numeric",
              year: "numeric",
            }),

      footer: (props: any) => props.column.id,
    },
  ];
};
