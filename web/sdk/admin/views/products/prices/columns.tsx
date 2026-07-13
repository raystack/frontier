import type { Price as PriceType } from "@raystack/proton/frontier";
import { Amount } from "@raystack/apsara";
import type { DataTableColumnDef } from "@raystack/apsara";
import { formatTimestamp, TimeStamp } from "~/admin/utils/connect-timestamp";
import styles from "./prices.module.css";

export const getColumns = (
  prices: PriceType[]
): DataTableColumnDef<PriceType, unknown>[] => {
  return [
    {
      header: "Id",
      accessorKey: "id",
      classNames: {
        cell: styles["first-column"],
        header: styles["first-column"],
      },
      cell: (info) => info.getValue(),
      filterType: "string",
    },
    {
      header: "name",
      accessorKey: "name",
      cell: (info) => info.getValue(),
      filterType: "string",
    },
    {
      header: "interval",
      accessorKey: "interval",
      cell: (info) => info.getValue(),
      filterType: "string",
    },
    {
      header: "Usage Type",
      accessorKey: "usageType",
      cell: (info) => info.getValue(),
      filterType: "string",
    },
    {
      header: "billing_scheme",
      accessorKey: "billingScheme",
      cell: (info) => info.getValue(),
      filterType: "string",
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
      filterType: "string",
    },
    {
      header: "creation date",
      accessorKey: "createdAt",
      cell: ({ getValue }) => formatTimestamp(getValue() as TimeStamp),
    },
    {
      header: "Updated date",
      accessorKey: "updatedAt",
      enableColumnFilter: false,
      cell: ({ getValue }) => formatTimestamp(getValue() as TimeStamp),
    },
  ];
};
