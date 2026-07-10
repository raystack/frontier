import { Text, type DataTableColumnDef } from "@raystack/apsara";
import type { Plan } from "@raystack/proton/frontier";
import { timestampCell } from "../../utils/connect-timestamp";
import styles from "./plans.module.css";

export const getColumns: (options?: {
  onSelectPlan?: (planId: string) => void;
}) => DataTableColumnDef<
  Plan,
  unknown
>[] = ({ onSelectPlan } = {}) => {
  return [
    {
      header: "ID",
      accessorKey: "id",
      classNames: {
        cell: styles["first-column"],
        header: styles["first-column"],
      },
      filterType: "string",
      cell: ({ getValue }) => {
        const id = getValue() as string;
        return (
          <Text
            style={{ cursor: "pointer" }}
            onClick={() => onSelectPlan?.(id)}
            data-test-id="frontier-admin-plan-link"
          >
            {id}
          </Text>
        );
      },
    },
    {
      header: "Title",
      accessorKey: "title",
      filterType: "string",
      cell: (info) => info.getValue(),
    },
    {
      header: "Interval",
      accessorKey: "interval",
      filterType: "string",
      cell: (info) => info.getValue(),
    },
    {
      header: "Created At",
      accessorKey: "createdAt",
      cell: timestampCell,
    },
  ];
};
