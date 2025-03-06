import { V1Beta1Group, V1Beta1Organization } from "@raystack/frontier";
import { Link } from "react-router-dom";
import * as R from "ramda";
import { ApsaraColumnDef } from "@raystack/apsara";
import { Text } from "@raystack/apsara/v1";

interface getColumnsOptions {
  orgMap: Record<string, V1Beta1Organization>;
}

export const getColumns: (
  opt: getColumnsOptions
) => ApsaraColumnDef<V1Beta1Group>[] = ({ orgMap }) => {
  return [
    {
      header: "Id",
      accessorKey: "id",
      filterVariant: "text",
      cell: ({ row, getValue }) => {
        return <Link to={`${row.getValue("id")}`}>{getValue()}</Link>;
      },
    },
    {
      header: "Title",
      accessorKey: "title",
      cell: (info: any) => info.getValue(),
      filterVariant: "text",
    },
    {
      header: "Organization",
      accessorKey: "org_id",
      cell: (info) => {
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
      filterVariant: "select",
    },
    {
      header: "Created At",
      accessorKey: "created_at",
      filterVariant: "date",
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
