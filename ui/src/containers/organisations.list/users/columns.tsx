import { Pencil2Icon } from "@radix-ui/react-icons";
import { Flex } from "@raystack/apsara";
import { V1Beta1User } from "@raystack/frontier";
import type { ColumnDef } from "@tanstack/react-table";
import { createColumnHelper } from "@tanstack/react-table";
import { Link, NavLink } from "react-router-dom";

const columnHelper = createColumnHelper<V1Beta1User>();

interface getColumnsOptions {
  users: V1Beta1User[];
  orgId: string;
}
export const getColumns: (
  opts: getColumnsOptions
) => ColumnDef<V1Beta1User, any>[] = ({ users, orgId }) => {
  return [
    columnHelper.accessor("id", {
      header: "ID",
      //@ts-ignore
      filterVariant: "text",
      cell: ({ row, getValue }) => {
        return <Link to={`/users/${row.getValue("id")}`}>{getValue()}</Link>;
      },
    }),
    {
      header: "Title",
      accessorKey: "title",
      filterVariant: "text",
      cell: (info) => info.getValue(),
    },
    {
      header: "Email",
      accessorKey: "email",
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

      footer: (props) => props.column.id,
    },
    {
      header: "Actions",
      cell: ({ row, getValue }) => {
        return (
          <Flex align="center" justify="center" gap="small">
            <NavLink to={`/organisations/${orgId}/users/${row?.original?.id}`}>
              <Pencil2Icon />
            </NavLink>
          </Flex>
        );
      },
    },
  ];
};
