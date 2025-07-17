import { Frontierv1Beta1Invoice, V1Beta1Price } from "@raystack/frontier";
import { Price } from "~/components/Price";
import type { DataTableColumnDef } from "@raystack/apsara/v1";

export const getColumns: (
  prices: V1Beta1Price[],
) => DataTableColumnDef<Frontierv1Beta1Invoice, unknown>[] = (
  prices: V1Beta1Price[],
) => {
  return [
    {
      header: "Id",
      accessorKey: "id",
      cell: info => info.getValue(),
      filterVariant: "text",
    },
    {
      header: "name",
      accessorKey: "name",
      cell: info => info.getValue(),
      filterVariant: "text",
    },
    {
      header: "interval",
      accessorKey: "interval",
      cell: info => info.getValue(),
      filterVariant: "text",
    },
    {
      header: "Usage Type",
      accessorKey: "usage_type",
      cell: info => info.getValue(),
      filterVariant: "text",
    },
    {
      header: "billing_scheme",
      accessorKey: "billing_scheme",
      cell: info => info.getValue(),
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
      cell: info =>
        new Date(info.getValue() as Date).toLocaleString("en", {
          month: "long",
          day: "numeric",
          year: "numeric",
        }),
    },
    {
      header: "Updated date",
      accessorKey: "updated_at",
      enableColumnFilter: false,
      cell: info =>
        new Date(info.getValue() as Date).toLocaleString("en", {
          month: "long",
          day: "numeric",
          year: "numeric",
        }),
    },
  ];
};
