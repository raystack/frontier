import { V1Beta1Invoice } from "@raystack/frontier";
import type { ColumnDef } from "@tanstack/react-table";
import { Price } from "~/components/Price";

export const getColumns: (
  invoices: V1Beta1Invoice[]
) => ColumnDef<V1Beta1Invoice, any>[] = (invoices: V1Beta1Invoice[]) => {
  return [
    {
      header: "Customer Id",
      accessorKey: "customer_id",
      cell: (info) => info.getValue(),
      filterVariant: "text",
    },
    {
      header: "Payment status",
      accessorKey: "state",
      cell: (info) => info.getValue(),
      meta: {
        data: [
          {
            label: "Paid",
            value: "paid",
          },
          {
            label: "Draft",
            value: "draft",
          },
        ],
      },
      filterFn: (row, id, value) => {
        return value.includes(row.getValue(id));
      },
    },
    {
      header: "URL",
      accessorKey: "hosted_url",
      cell: (info) => (
        <div style={{ width: "320px", wordWrap: "break-word" }}>
          <a target="_blank" href={info.getValue()} data-test-id="admin-ui-hosted-url-anchor">
            {info.getValue()}
          </a>
        </div>
      ),
      filterVariant: "text",
      style: { width: "100px" },
    },
    {
      header: "Amount",
      accessorKey: "amount",
      cell: ({ row }) => (
        <Price value={row.original.amount} currency={row.original.currency} />
      ),
      filterVariant: "text",
    },
    {
      header: "Invoice date",
      accessorKey: "effective_at",
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
      header: "Invoice creation date",
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
      header: "Due date",
      accessorKey: "due_date",
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
