import { V1Beta1Invoice, V1Beta1Price } from "@raystack/frontier";
import type { ColumnDef } from "@tanstack/react-table";
import { Price } from "~/components/Price";

export const getColumns: (
  prices: V1Beta1Price[]
) => ColumnDef<V1Beta1Invoice, any>[] = (prices: V1Beta1Price[]) => {
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
      accessorKey: "usage_type",
      cell: (info) => info.getValue(),
      filterVariant: "text",
    },
    {
      header: "billing_scheme",
      accessorKey: "billing_scheme",
      cell: (info) => info.getValue(),
      filterVariant: "text",
    },

    {
      header: "Amount",
      accessorKey: "amount",
      cell: ({ row, getValue }) => (
        <Price value={row.original.amount} currency={row.original.currency} />
      ),
      filterVariant: "text",
    },

    {
      header: "creation date",
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

      footer: (props) => props.column.id,
    },
    {
      header: "Updated date",
      accessorKey: "updated_at",
      meta: {
        headerFilter: false,
      },
      enableColumnFilter: false,
      cell: (info) =>
        new Date(info.getValue() as Date).toLocaleString("en", {
          month: "long",
          day: "numeric",
          year: "numeric",
        }),

      footer: (props) => props.column.id,
    },
  ];
};
