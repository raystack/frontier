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
      enableColumnFilter: true,
    },
    {
      header: "name",
      accessorKey: "name",
      cell: (info) => info.getValue(),
      filterType: "string",
      enableColumnFilter: true,
    },
    {
      header: "interval",
      accessorKey: "interval",
      cell: (info) => info.getValue(),
      filterType: "string",
      enableColumnFilter: true,
    },
    {
      header: "Usage Type",
      accessorKey: "usageType",
      cell: (info) => info.getValue(),
      filterType: "string",
      enableColumnFilter: true,
    },
    {
      header: "billing_scheme",
      accessorKey: "billingScheme",
      cell: (info) => info.getValue(),
      filterType: "string",
      enableColumnFilter: true,
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
      enableColumnFilter: true,
    },
    {
      header: "creation date",
      accessorKey: "createdAt",
      filterType: "date",
      enableColumnFilter: true,
      cell: ({ getValue }) => formatTimestamp(getValue() as TimeStamp),
    },
    {
      header: "Updated date",
      accessorKey: "updatedAt",
      filterType: "date",
      enableColumnFilter: false,
      cell: ({ getValue }) => formatTimestamp(getValue() as TimeStamp),
    },
  ];
};
