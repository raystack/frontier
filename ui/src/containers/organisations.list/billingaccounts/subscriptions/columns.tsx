import { V1Beta1Subscription } from "@raystack/frontier";
import type { ColumnDef } from "@tanstack/react-table";
import { createColumnHelper } from "@tanstack/react-table";

const columnHelper = createColumnHelper<V1Beta1Subscription>();
export const getColumns: (
  subscriptions: V1Beta1Subscription[]
) => ColumnDef<V1Beta1Subscription, any>[] = (
  subscriptions: V1Beta1Subscription[]
) => {
  return [
    {
      header: "Title",
      accessorKey: "title",
      cell: (info) => info.getValue(),
      filterVariant: "text",
    },
    {
      header: "Customer Id",
      accessorKey: "customer_id",
      cell: (info) => info.getValue(),
      filterVariant: "text",
    },
    {
      header: "Provider Id",
      accessorKey: "provider_id",
      cell: (info) => info.getValue(),
      filterVariant: "text",
    },
    {
      header: "Plan Id",
      accessorKey: "plan_id",
      cell: (info) => info.getValue(),
      filterVariant: "text",
    },
    {
      header: "Create At",
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
      header: "Ended At",
      accessorKey: "ended_at",
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
