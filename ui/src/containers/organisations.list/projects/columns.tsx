import { V1Beta1Project } from "@raystack/frontier";
import type { ColumnDef } from "@tanstack/react-table";
import { createColumnHelper } from "@tanstack/react-table";
import { Link } from "react-router-dom";

const columnHelper = createColumnHelper<V1Beta1Project>();
export const getColumns: (
  projects: V1Beta1Project[]
) => ColumnDef<V1Beta1Project, any>[] = (projects: V1Beta1Project[]) => {
  return [
    columnHelper.accessor("id", {
      header: "ID",
      //@ts-ignore
      filterVariant: "text",
      cell: ({ row, getValue }) => {
        return <Link to={`/projects/${row.getValue("id")}`}>{getValue()}</Link>;
      },
    }),
    {
      header: "Title",
      accessorKey: "title",
      cell: (info) => info.getValue(),
      filterVariant: "text",
    },
    {
      header: "Organization Id",
      accessorKey: "org_id",
      cell: (info) => info.getValue(),
      filterVariant: "text",
    },
    {
      header: "Create At",
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

      footer: (props) => props.column.id,
    },
  ];
};
