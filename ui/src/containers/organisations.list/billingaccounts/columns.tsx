import { V1Beta1BillingAccount } from "@raystack/frontier";
import type { ColumnDef } from "@tanstack/react-table";
import { createColumnHelper } from "@tanstack/react-table";
import { Link, useParams } from "react-router-dom";

const columnHelper = createColumnHelper<V1Beta1BillingAccount>();
export const getColumns: (
  billingAccounts: V1Beta1BillingAccount[]
) => ColumnDef<V1Beta1BillingAccount, any>[] = (
  billingAccounts: V1Beta1BillingAccount[]
) => {
  let { organisationId } = useParams();
  return [
    columnHelper.accessor("id", {
      header: "ID",
      //@ts-ignore
      filterVariant: "text",
      cell: ({ row, getValue }) => {
        return (
          <Link
            to={`/organisations/${organisationId}/billingaccounts/${row.getValue(
              "id"
            )}`}
          >
            {getValue()}
          </Link>
        );
      },
    }),
    {
      header: "Organization Id",
      accessorKey: "org_id",
      cell: (info) => info.getValue(),
      filterVariant: "text",
    },
    {
      header: "Title",
      accessorKey: "name",
      cell: (info) => info.getValue(),
      filterVariant: "text",
    },
    {
      header: "Provider",
      accessorKey: "provider",
      cell: (info) => info.getValue(),
      filterVariant: "text",
    },
    {
      header: "State",
      accessorKey: "state",
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
  ];
};
