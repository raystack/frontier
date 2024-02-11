import { V1Beta1Invoice, V1Beta1Subscription } from "@raystack/frontier";
import type { ColumnDef } from "@tanstack/react-table";
import { createColumnHelper } from "@tanstack/react-table";
import { Link, useParams } from "react-router-dom";

const columnHelper = createColumnHelper<V1Beta1Subscription>();
export const getColumns: (
  invoices: V1Beta1Invoice[]
) => ColumnDef<V1Beta1Invoice, any>[] = (invoices: V1Beta1Invoice[]) => {
  let { organisationId } = useParams();
  return [
    {
      header: "Id",
      accessorKey: "id",
      cell: (info) => info.getValue(),
      filterVariant: "text",
    },
    {
      header: "Org Id",
      accessorKey: "id",
      cell: ({ row, getValue }) => {
        return (
          <Link to={`/organisations/${organisationId}`}>{getValue()}</Link>
        );
      },
      filterVariant: "text",
    },
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
      filterVariant: "text",
    },
    {
      header: "URL",
      accessorKey: "hosted_url",
      cell: (info) => info.getValue(),
      filterVariant: "text",
    },
    {
      header: "Amount",
      accessorKey: "amount",
      cell: (info) => info.getValue(),
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
