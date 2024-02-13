import { V1Beta1Invoice, V1Beta1Subscription } from "@raystack/frontier";
import type { ColumnDef } from "@tanstack/react-table";
import { createColumnHelper } from "@tanstack/react-table";
import { useParams } from "react-router-dom";

const columnHelper = createColumnHelper<V1Beta1Subscription>();
export const getColumns: (
  invoices: V1Beta1Invoice[]
) => ColumnDef<V1Beta1Invoice, any>[] = (invoices: V1Beta1Invoice[]) => {
  let { organisationId } = useParams();
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
          <a target="_blank" href={info.getValue()}>
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
      cell: ({ row, getValue }) =>
        `${parseInt(row.original.amount ?? "") / 100} ${
          row?.original?.currency
        }`,
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
