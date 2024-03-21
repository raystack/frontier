import { Pencil2Icon } from "@radix-ui/react-icons";
import { Flex, Text } from "@raystack/apsara";
import {
  V1Beta1ListOrganizationUsersResponseRolePair,
  V1Beta1User,
} from "@raystack/frontier";
import type { ColumnDef } from "@tanstack/react-table";
import { Link, NavLink } from "react-router-dom";
import * as R from "ramda";

interface getColumnsOptions {
  users: V1Beta1User[];
  orgId: string;
  userRolesMap: Record<string, V1Beta1ListOrganizationUsersResponseRolePair>;
}
export const getColumns: (
  opts: getColumnsOptions
) => ColumnDef<V1Beta1User, any>[] = ({ users, orgId, userRolesMap }) => {
  return [
    {
      header: "Title",
      accessorKey: "title",
      filterVariant: "text",
      cell: ({ row, getValue }) => {
        return <Link to={`/users/${row.getValue("id")}`}>{getValue()}</Link>;
      },
    },
    {
      header: "Email",
      accessorKey: "email",
      filterVariant: "text",
      cell: (info) => info.getValue(),
      footer: (props) => props.column.id,
    },
    {
      header: "Roles",
      filterVariant: "text",
      cell: ({ row }) => {
        const userRoles = R.pipe(
          R.pathOr([], [row?.original?.id || "", "roles"]),
          R.map(R.path(["title"])),
          R.join(", ")
        )(userRolesMap);

        return <Text>{userRoles}</Text>;
      },
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
