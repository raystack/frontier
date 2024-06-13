import { V1Beta1Organization, V1Beta1Project } from "@raystack/frontier";
import { Link } from "react-router-dom";
import { ApsaraColumnDef, Text } from "@raystack/apsara";
import * as R from "ramda";

interface getColumnsOptions {
  orgMap: Record<string, V1Beta1Organization>;
}

export const getColumns: (
  opt: getColumnsOptions
) => ApsaraColumnDef<V1Beta1Project>[] = ({ orgMap = {} }) => {
  return [
    {
      header: "ID",
      accessorKey: "id",

      filterVariant: "text",
      cell: ({ row, getValue }) => {
        return <Link to={`/projects/${row.getValue("id")}`}>{getValue()}</Link>;
      },
    },
    {
      header: "Title",
      accessorKey: "title",
      cell: (info) => info.getValue(),
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
      filterVariant: "text",
    },

    {
      header: "Created At",
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
      filterVariant: "date",
      footer: (props) => props.column.id,
    },
  ];
};
