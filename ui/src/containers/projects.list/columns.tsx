import type { ColumnDef } from "@tanstack/react-table";
import { createColumnHelper } from "@tanstack/react-table";
import { Link } from "react-router-dom";
import type { Project } from "~/types/project";

const columnHelper = createColumnHelper<Project>();
export const getColumns: (projects: Project[]) => ColumnDef<Project, any>[] = (
  projects: Project[]
) => {
  return [
    columnHelper.accessor("id", {
      header: "ID",
      //@ts-ignore
      filterVariant: "text",
      cell: ({ row, getValue }) => {
        return <Link to={`${row.getValue("id")}`}>{getValue()}</Link>;
      },
    }),
    {
      header: "Title",
      accessorKey: "title",
      cell: (info) => info.getValue(),
      filterVariant: "text",
    },
    {
      header: "Name",
      accessorKey: "name",
      filterVariant: "text",
      cell: (info) => info.getValue(),
      footer: (props) => props.column.id,
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
