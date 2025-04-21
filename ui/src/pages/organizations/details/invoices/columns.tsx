import { NULL_DATE } from "~/utils/constants";
import styles from "./invoices.module.css";
import dayjs from "dayjs";
import { DataTableColumnDef, Link } from "@raystack/apsara/v1";
import { SearchOrganizationInvoicesResponseOrganizationInvoice } from "~/api/frontier";
import { Amount } from "@raystack/frontier/react";

// https://docs.stripe.com/invoicing/overview#invoice-statuses
const InvoiceStatusesMap = {
  draft: "Draft",
  open: "Open",
  paid: "Paid",
  void: "Void",
  uncollectible: "Uncollectible",
} as const;

type InvoiceStatusKey = keyof typeof InvoiceStatusesMap;

export const getColumns = (): DataTableColumnDef<
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
  },
  {
    accessorKey: "amount",
    header: "Amount",
    cell: ({ getValue, row }) => {
      const value = Number(getValue());
      return <Amount value={value} currency={row.original.currency} />;
    },
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
