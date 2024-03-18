import { V1Beta1Invoice, V1Beta1Organization } from "@raystack/frontier";
import { ColumnDef } from "@tanstack/react-table";
import Skeleton from "react-loading-skeleton";
import { Link } from "react-router-dom";
import { Text } from "@raystack/apsara";
import dayjs from "dayjs";
import * as R from "ramda";

interface getColumnsOptions {
  isLoading: boolean;
  billingOrgMap: Record<string, string>;
  orgMap: Record<string, V1Beta1Organization>;
}

export const getColumns: (
  opt: getColumnsOptions
) => ColumnDef<V1Beta1Invoice, any>[] = ({
  orgMap,
  isLoading,
  billingOrgMap,
}) => {
  return [
    {
      header: "Date",
      accessorKey: "due_date",
      cell: isLoading
        ? () => <Skeleton />
        : ({ row, getValue }) => {
            const date = getValue() || row.original.period_end_at;
            return <Text>{dayjs(date).format("MMM DD, YYYY")}</Text>;
          },
    },
    {
      header: "Amount",
      accessorKey: "amount",
      cell: isLoading
        ? () => <Skeleton />
        : ({ row, getValue }) => {
            const currency = row?.original?.currency;
            // TODO: handle currency and decimal
            return (
              <Text>
                {currency} {getValue()}
              </Text>
            );
          },
    },
    {
      header: "Organization",
      accessorKey: "customer_id",
      cell: isLoading
        ? () => <Skeleton />
        : ({ row, getValue }) => {
            const billingId = getValue();
            const orgId = R.pathOr("", [billingId], billingOrgMap);
            const orgName = R.pathOr("", [orgId, "title"], orgMap);
            return <Text>{orgName}</Text>;
          },
    },
    {
      header: "State",
      accessorKey: "state",
      cell: isLoading
        ? () => <Skeleton />
        : ({ row, getValue }) => {
            return <Text>{getValue()}</Text>;
          },
    },
    {
      header: "Link",
      accessorKey: "hosted_url",
      cell: isLoading
        ? () => <Skeleton />
        : ({ row, getValue }) => {
            const link = getValue();
            return link ? (
              <Link to={link} target="__blank">
                Link
              </Link>
            ) : (
              <Text>-</Text>
            );
          },
    },
  ];
};
