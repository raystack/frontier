import type { Price as PriceType } from "@raystack/proton/frontier";
import { Amount } from "@raystack/apsara";
import type { DataTableColumnDef } from "@raystack/apsara";
import { timestampToDate, TimeStamp } from "../../../utils/connect-timestamp";

export const getColumns = (
  prices: PriceType[]
): DataTableColumnDef<PriceType, unknown>[] => {
  return [
    {
      header: "Id",
      accessorKey: "id",
      cell: (info) => info.getValue(),
      filterVariant: "text",
    },
    {
      header: "name",
      accessorKey: "name",
      cell: (info) => info.getValue(),
      filterVariant: "text",
    },
    {
      header: "interval",
      accessorKey: "interval",
      cell: (info) => info.getValue(),
      filterVariant: "text",
    },
    {
      header: "Usage Type",
      accessorKey: "usageType",
      cell: (info) => info.getValue(),
      filterVariant: "text",
    },
    {
      header: "billing_scheme",
      accessorKey: "billingScheme",
      cell: (info) => info.getValue(),
      filterVariant: "text",
    },
    {
      header: "Amount",
      accessorKey: "amount",
      cell: ({ row }) => (
        <Amount
          value={Number(row.original.amount ?? 0)}
          currency={row.original.currency}
        />
      ),
      filterVariant: "text",
    },
    {
      header: "creation date",
      accessorKey: "createdAt",
      cell: ({ getValue }) => {
        const timestamp = getValue() as TimeStamp | undefined;
        const date = timestampToDate(timestamp);
        if (!date) return "-";
        return date.toLocaleString("en", {
          month: "long",
          day: "numeric",
          year: "numeric",
        });
      },
    },
    {
      header: "Updated date",
      accessorKey: "updatedAt",
      enableColumnFilter: false,
      cell: ({ getValue }) => {
        const timestamp = getValue() as TimeStamp | undefined;
        const date = timestampToDate(timestamp);
        if (!date) return "-";
        return date.toLocaleString("en", {
          month: "long",
          day: "numeric",
          year: "numeric",
        });
      },
    },
  ];
};
