import { type DataTableColumnDef, Link, Text } from "@raystack/apsara/v1";
import dayjs from "dayjs";
import type { V1Beta1SearchInvoicesResponseInvoice } from "~/api/frontier";
import { NULL_DATE } from "~/utils/constants";

export const getColumns = (): DataTableColumnDef<
  V1Beta1SearchInvoicesResponseInvoice,
  unknown
>[] => {
  return [
    {
      accessorKey: "created_at",
      header: "Billed on",
      filterType: "date",
      cell: ({ getValue }) => {
        const value = getValue() as string;
        return value !== NULL_DATE ? dayjs(value).format("YYYY-MM-DD") : "-";
      },
      enableHiding: true,
      enableSorting: true,
    },
    {
      accessorKey: "state",
      header: "State",
      cell: ({ row, getValue }) => {
        return getValue() as string;
      },
    },
    {
      accessorKey: "org_title",
      header: "Organization",
      cell: ({ row, getValue }) => {
        return getValue() as string;
      },
    },
    {
      accessorKey: "amount",
      header: "Amount",
      cell: ({ row, getValue }) => {
        return getValue() as string;
      },
    },
    {
      accessorKey: "invoice_link",
      header: "",
      cell: ({ row, getValue }) => {
        const link = getValue() as string;
        return link ? (
          <Link href={link} target="__blank">
            Link
          </Link>
        ) : (
          <Text>-</Text>
        );
      },
    },
  ];
};
