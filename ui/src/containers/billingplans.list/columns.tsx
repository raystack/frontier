import { Link } from "react-router-dom";
import { type DataTableColumnDef } from "@raystack/apsara/v1";
import type { V1Beta1Plan } from "@raystack/frontier";

export const getColumns: () => DataTableColumnDef<
  V1Beta1Plan,
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
      accessorKey: "created_at",
      cell: (info) =>
        new Date(info.getValue() as Date).toLocaleString("en", {
          month: "long",
          day: "numeric",
          year: "numeric",
        }),
    },
  ];
};
