import { Text, type DataTableColumnDef } from "@raystack/apsara";
import styles from "./webhooks.module.css";
import { type Webhook } from "@raystack/proton/frontier";
import {
  timestampToDate,
  isNullTimestamp,
  type TimeStamp,
} from "../../../utils/connect-timestamp";
import dayjs from "dayjs";

export const getColumns: () => DataTableColumnDef<Webhook, unknown>[] = () => {
  return [
    {
      header: "Description",
      accessorKey: "description",
      classNames: {
        cell: styles["first-column"],
        header: styles["first-column"],
      },
      filterVariant: "text",
      cell: (info) => info.getValue() || "-",
    },
    {
      header: "State",
      accessorKey: "state",
      filterVariant: "text",
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
      cell: ({ getValue }) => {
        const value = getValue() as TimeStamp;
        const date = isNullTimestamp(value)
          ? "-"
          : dayjs(timestampToDate(value)).format("YYYY-MM-DD");
        return <Text>{date}</Text>;
      },
    },
  ];
};
