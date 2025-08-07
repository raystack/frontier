import { NULL_DATE } from "~/utils/constants";
import styles from "./invoices.module.css";
import dayjs from "dayjs";
import { DataTableColumnDef, Link, Amount } from "@raystack/apsara";
import { SearchOrganizationInvoicesResponseOrganizationInvoice } from "~/api/frontier";

// https://docs.stripe.com/invoicing/overview#invoice-statuses
const InvoiceStatusesMap = {
  draft: "Draft",
  open: "Open",
  paid: "Paid",
  void: "Void",
  uncollectible: "Uncollectible",
} as const;

type InvoiceStatusKey = keyof typeof InvoiceStatusesMap;

interface getColumnsOptions {
  groupCountMap: Record<string, Record<string, number>>;
}

export const getColumns = ({
  groupCountMap,
}: getColumnsOptions): DataTableColumnDef<
  SearchOrganizationInvoicesResponseOrganizationInvoice,
  unknown
>[] => [
  {
    accessorKey: "created_at",
    header: "Billed on",
    classNames: {
      cell: styles["first-column"],
      header: styles["first-column"],
    },
    cell: ({ getValue }) => {
      const value = getValue() as string;
      return value !== NULL_DATE ? dayjs(value).format("YYYY-MM-DD") : "-";
    },
    enableSorting: true,
    enableColumnFilter: true,
    filterType: "date",
  },
  {
    accessorKey: "state",
    header: "Status",
    cell: ({ getValue }) => {
      const value = getValue() as InvoiceStatusKey;
      return InvoiceStatusesMap[value];
    },
    enableColumnFilter: true,
    filterType: "select",
    filterOptions: Object.entries(InvoiceStatusesMap).map(([key, value]) => ({
      label: value,
      value: key,
    })),
    enableGrouping: true,
    groupLabelsMap: InvoiceStatusesMap,
    showGroupCount: true,
    groupCountMap: groupCountMap["state"] || {},
    enableHiding: true,
  },
  {
    accessorKey: "amount",
    header: "Amount",
    cell: ({ getValue, row }) => {
      const value = Number(getValue());
      return <Amount value={value} currency={row.original.currency} />;
    },
    enableSorting: true,
    enableHiding: true,
    enableColumnFilter: true,
    filterType: "number",
  },
  {
    accessorKey: "invoice_link",
    header: "",
    cell: ({ getValue }) => {
      const value = getValue() as string;
      return (
        <Link href={value} target="_blank" data-test-id="invoice-link">
          View Invoice
        </Link>
      );
    },
  },
];
