import { V1Beta1Group, V1Beta1Organization } from "@raystack/frontier";
import type { ColumnDef } from "@tanstack/react-table";
import { createColumnHelper } from "@tanstack/react-table";
import Skeleton from "react-loading-skeleton";
import { Link } from "react-router-dom";
import * as R from "ramda";
import { Text } from "@raystack/apsara";
const columnHelper = createColumnHelper<V1Beta1Group>();

interface getColumnsOptions {
  isLoading: boolean;
  orgMap: Record<string, V1Beta1Organization>;
}

export const getColumns: (
  opt: getColumnsOptions
) => ColumnDef<V1Beta1Group, any>[] = ({ orgMap, isLoading }) => {
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
      header: "Organization",
      accessorKey: "org_id",
      cell: isLoading
        ? () => <Skeleton />
        : (info) => {
            const orgId = info.getValue();
            const orgName = R.pathOr(orgId, [orgId, "title"], orgMap);
            return <Text>{orgName}</Text>;
          },
      meta: {
        data: Object.entries(orgMap).map(([k, v]) => ({
          label: v?.title,
          value: k,
        })),
      },
      filterFn: (row, id, value) => {
        return value.includes(row.getValue(id));
      },
    },
    {
      header: "Created At",
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
