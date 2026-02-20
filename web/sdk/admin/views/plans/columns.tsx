import { Link } from "react-router-dom";
import { type DataTableColumnDef } from "@raystack/apsara";
import type { Plan } from "@raystack/proton/frontier";
import { timestampToDate, type TimeStamp } from "../../utils/connect-timestamp";

export const getColumns: () => DataTableColumnDef<
  Plan,
  unknown
>[] = () => {
  return [
    {
      header: "ID",
      accessorKey: "id",
      filterVariant: "text",
      cell: ({ row, getValue }) => (
        <Link to={`/plans/${row.getValue("id")}`}>{getValue() as string}</Link>
      ),
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
