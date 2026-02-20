import { Amount, type DataTableColumnDef, Link, Text } from "@raystack/apsara";
import dayjs from "dayjs";
import type { SearchInvoicesResponse_Invoice } from "@raystack/proton/frontier";
import {
  isNullTimestamp,
  TimeStamp,
  timestampToDate,
} from "../../utils/connect-timestamp";

export const getColumns = (): DataTableColumnDef<
  SearchInvoicesResponse_Invoice,
  unknown
>[] => {
  return [
    {
      accessorKey: "createdAt",
      header: "Billed on",
      filterType: "date",
      enableColumnFilter: true,
      cell: ({ getValue }) => {
        const value = getValue() as TimeStamp;
        const date = isNullTimestamp(value)
          ? "-"
          : dayjs(timestampToDate(value)).format("YYYY-MM-DD");
        return <Text>{date}</Text>;
      },
      enableHiding: true,
      enableSorting: true,
    },
    {
      enableColumnFilter: true,
      accessorKey: "state",
      filterOptions: ["paid", "open", "draft"].map(value => ({
        value: value,
        label: value,
      })),
      filterType: "select",
      header: "State",
      cell: ({ row, getValue }) => {
        return getValue() as string;
      },
    },
    {
      accessorKey: "orgTitle",
      header: "Organization",
      cell: ({ row, getValue }) => {
        return getValue() as string;
      },
    },
    {
      header: "Amount",
      accessorKey: "amount",
      cell: ({ row, getValue }) => {
        const value = Number(getValue());
        return <Amount value={value} currency={row.original.currency} />;
      },
    },
    {
      accessorKey: "invoiceLink",
      header: "",
      cell: ({ row, getValue }) => {
        const link = getValue() as string;
        return link ? (
          <Link href={link} external={true}>
            Link
          </Link>
        ) : (
          <Text>-</Text>
        );
      },
    },
  ];
};
