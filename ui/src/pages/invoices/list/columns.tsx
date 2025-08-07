import { Amount, type DataTableColumnDef, Link, Text } from "@raystack/apsara";
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
      enableColumnFilter: true,
      cell: ({ getValue }) => {
        const value = getValue() as string;
        return value !== NULL_DATE ? dayjs(value).format("YYYY-MM-DD") : "-";
      },
      enableHiding: true,
      enableSorting: true,
    },
    {
      enableColumnFilter: true,
      accessorKey: "state",
      filterOptions: ["paid", "open", "draft"].map((value) => ({
        value: value,
        label: value,
      })),
      filterType: "select",
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
      header: "Amount",
      accessorKey: "amount",
      cell: ({ row, getValue }) => {
        const currency = row?.original?.currency;
        const value = getValue() as number;
        return <Amount value={value} currency={currency} />;
      },
    },
    {
      accessorKey: "invoice_link",
      header: "",
      cell: ({ row, getValue }) => {
        const link = getValue() as string;
        return link ? (
          <Link href={link} external={true}>
            Link
          </Link>
        ) : (
          <Text>-</Text>
        );
      },
    },
  ];
};
