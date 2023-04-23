import type { ColumnDef } from "@tanstack/react-table";
import { createColumnHelper } from "@tanstack/react-table";
import { compose, map, pathOr, uniq } from "ramda";
import { Link } from "react-router-dom";
import type { Group } from "~/types/group";

import { keyToColumnMetaObject } from "~/utils/helper";

const columnHelper = createColumnHelper<Group>();
export const getColumns: (groups: Group[]) => ColumnDef<Group, any>[] = (
  groups: Group[]
) => {
  if (!groups) return [];

  const uniqueValues = (propsValue: string) =>
    compose(
      map(keyToColumnMetaObject),
      uniq,
      map(pathOr([], propsValue.split(".")))
    )(groups);

  const uniqueNames = uniqueValues("name");

  return [
    columnHelper.accessor("id", {
      header: "ID",
      //@ts-ignore
      filterVariant: "text",
      cell: ({ row, getValue }) => {
        return <Link to={`${row.getValue("id")}`}>{getValue()}</Link>;
      },
    }),
    {
      header: "Name",
      accessorKey: "name",
      cell: (info: any) => info.getValue(),
      filterVariant: "text",
    },
    {
      header: "Slug",
      accessorKey: "slug",
      cell: (info: any) => info.getValue(),
      filterVariant: "text",
    },
    {
      header: "Create At",
      accessorKey: "createdAt",
      filterVariant: "text",
      meta: {
        headerFilter: false,
      },
      cell: (info: any) =>
        new Date(info.getValue() as Date).toLocaleString("en", {
          month: "long",
          day: "numeric",
          year: "numeric",
        }),

      footer: (props: any) => props.column.id,
    },
  ];
};
