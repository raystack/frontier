import { V1Beta1Invoice, V1Beta1Organization } from "@raystack/frontier";
import { Link } from "react-router-dom";
import { ApsaraColumnDef } from "@raystack/apsara";
import { Text } from "@raystack/apsara/v1";
import dayjs from "dayjs";
import * as R from "ramda";
import { DEFAULT_DATE_FORMAT } from "~/utils/constants";

interface getColumnsOptions {
  billingOrgMap: Record<string, string>;
  orgMap: Record<string, V1Beta1Organization>;
}

const currencyDecimal = 2;

export const getColumns: (
  opt: getColumnsOptions
) => ApsaraColumnDef<V1Beta1Invoice>[] = ({ orgMap, billingOrgMap }) => {
  return [
    {
      header: "Date",
      filterVariant: "date",
      accessorKey: "due_date",
      cell: ({ row, getValue }) => {
        const date = getValue() || row.original.period_end_at;
        return <Text>{dayjs(date).format(DEFAULT_DATE_FORMAT)}</Text>;
      },
    },
    {
      header: "Amount",
      accessorKey: "amount",
      cell: ({ row, getValue }) => {
        const currency = row?.original?.currency;

        const value = getValue();

        const calculatedValue = (value / Math.pow(10, currencyDecimal)).toFixed(
          currencyDecimal
        );
        const [intValue, decimalValue] = calculatedValue.toString().split(".");

        return (
          <Text>
            {currency} {intValue}.{decimalValue}
          </Text>
        );
      },
    },
    {
      header: "Organization",
      accessorKey: "customer_id",
      cell: ({ row, getValue }) => {
        const billingId = getValue();
        const orgId = R.pathOr("", [billingId], billingOrgMap);
        const orgName = R.pathOr("", [orgId, "title"], orgMap);
        return <Text>{orgName}</Text>;
      },
    },
    {
      header: "State",
      accessorKey: "state",
      cell: ({ row, getValue }) => {
        return <Text>{getValue()}</Text>;
      },
    },
    {
      header: "Link",
      accessorKey: "hosted_url",
      enableColumnFilter: false,
      cell: ({ row, getValue }) => {
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
