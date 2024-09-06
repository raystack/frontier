import { Link } from "react-router-dom";
import { ApsaraColumnDef } from "@raystack/apsara";
import { V1Beta1Plan } from "@raystack/frontier";

export const getColumns: () => ApsaraColumnDef<V1Beta1Plan>[] = () => {
  return [
    {
      header: "ID",
      accessorKey: "id",
      filterVariant: "text",
      cell: ({ row, getValue }) => (
        <Link to={`/plans/${row.getValue("id")}`}>{getValue()}</Link>
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
      footer: (props) => props.column.id,
    },
    {
      header: "Created At",
      accessorKey: "created_at",
      meta: {
        headerFilter: false,
      },
      cell: (info) =>
        new Date(info.getValue() as Date).toLocaleString("en", {
          month: "long",
          day: "numeric",
          year: "numeric",
        }),
      filterVariant: "date",
      footer: (props) => props.column.id,
    },
  ];
};
