import { Text, type DataTableColumnDef } from "@raystack/apsara";
import type { Plan } from "@raystack/proton/frontier";
import { timestampToDate, type TimeStamp } from "../../utils/connect-timestamp";

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
      filterVariant: "text",
      cell: ({ getValue }) => {
        const id = getValue() as string;
        return (
          <Text
            style={{ cursor: "pointer" }}
            onClick={() => onSelectPlan?.(id)}
          >
            {id}
          </Text>
        );
      },
    },
    {
      header: "Title",
      accessorKey: "title",
      filterVariant: "text",
      cell: (info) => info.getValue(),
    },
    {
      header: "Interval",
      accessorKey: "interval",
      filterVariant: "text",
      cell: (info) => info.getValue(),
    },
    {
      header: "Created At",
      accessorKey: "createdAt",
      cell: ({ getValue }) => {
        const timestamp = getValue() as TimeStamp | undefined;
        const date = timestampToDate(timestamp);
        if (!date) return "-";
        return date.toLocaleString("en", {
          month: "long",
          day: "numeric",
          year: "numeric",
        });
      },
    },
  ];
};
