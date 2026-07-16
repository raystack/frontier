import { Text, type DataTableColumnDef } from "@raystack/apsara";
import styles from "./webhooks.module.css";
import { type Webhook } from "@raystack/proton/frontier";
import {
  formatTimestamp,
  type TimeStamp,
} from "~/admin/utils/connect-timestamp";

export const getColumns: () => DataTableColumnDef<Webhook, unknown>[] = () => {
  return [
    {
      header: "Description",
      accessorKey: "description",
      classNames: {
        cell: styles["first-column"],
        header: styles["first-column"],
      },
      filterType: "string",
      cell: (info) => info.getValue() || "-",
    },
    {
      header: "State",
      accessorKey: "state",
      filterType: "string",
      classNames: { cell: styles.stateColumn, header: styles.stateColumn },
      cell: (info) => info.getValue() || "-",
    },
    {
      header: "URL",
      accessorKey: "url",
      cell: (info) => info.getValue() || "-",
    },
    {
      header: "Created at",
      accessorKey: "createdAt",
      classNames: { cell: styles.dateColumn, header: styles.dateColumn },
      cell: ({ getValue }) => (
        <Text>{formatTimestamp(getValue() as TimeStamp)}</Text>
      ),
    },
  ];
};
