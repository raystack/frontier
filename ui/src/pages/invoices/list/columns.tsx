import {
  type DataTableColumnDef,
  EmptyFilterValue,
  Link,
  Text,
} from "@raystack/apsara/v1";
import dayjs from "dayjs";
import type { V1Beta1SearchInvoicesResponseInvoice } from "~/api/frontier";
import { NULL_DATE } from "~/utils/constants";
const currencyDecimal = 2;

export const getColumns = (): DataTableColumnDef<
  V1Beta1SearchInvoicesResponseInvoice,
  unknown
>[] => {
  return [
    {
      accessorKey: "created_at",
      header: "Billed on",
      filterType: "date",
      enableColumnFilter: true,
      cell: ({ getValue }) => {
        const value = getValue() as string;
        return value !== NULL_DATE ? dayjs(value).format("YYYY-MM-DD") : "-";
      },
      enableHiding: true,
      enableSorting: true,
    },
    {
      enableColumnFilter: true,
      accessorKey: "state",
      filterOptions: ["paid", "open", "draft"].map((value) => ({
        value: value === "" ? EmptyFilterValue : value,
        label: value,
      })),
      filterType: "select",
      header: "State",
      cell: ({ row, getValue }) => {
        return getValue() as string;
      },
    },
    {
      accessorKey: "org_title",
      header: "Organization",
      cell: ({ row, getValue }) => {
        return getValue() as string;
      },
    },
    {
      header: "Amount",
      accessorKey: "amount",
      cell: ({ row, getValue }) => {
        const currency = row?.original?.currency;

        const value = getValue() as number;

        const calculatedValue = (value / Math.pow(10, currencyDecimal)).toFixed(
          currencyDecimal,
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
      accessorKey: "invoice_link",
      header: "",
      cell: ({ row, getValue }) => {
        const link = getValue() as string;
        return link ? (
          <Link href={link} target="__blank">
            Link
          </Link>
        ) : (
          <Text>-</Text>
        );
      },
    },
  ];
};
