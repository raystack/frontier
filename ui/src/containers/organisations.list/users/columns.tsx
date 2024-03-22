import {
  DotsHorizontalIcon,
  Pencil2Icon,
  TrashIcon,
  UpdateIcon,
} from "@radix-ui/react-icons";
import { DropdownMenu, Flex, Text } from "@raystack/apsara";
import {
  V1Beta1ListOrganizationUsersResponseRolePair,
  V1Beta1Policy,
  V1Beta1Role,
  V1Beta1User,
} from "@raystack/frontier";
import type { ColumnDef } from "@tanstack/react-table";
import { Link, NavLink } from "react-router-dom";
import * as R from "ramda";
import { useFrontier } from "@raystack/frontier/react";
import { toast } from "sonner";

interface getColumnsOptions {
  users: V1Beta1User[];
  orgId: string;
  userRolesMap: Record<string, V1Beta1ListOrganizationUsersResponseRolePair>;
  roles: V1Beta1Role[];
  refetchUsers: () => void;
}
export const getColumns: (
  opts: getColumnsOptions
) => ColumnDef<V1Beta1User, any>[] = ({
  users,
  orgId,
  userRolesMap,
  roles = [],
  refetchUsers,
}) => {
  return [
    {
      header: "Title",
      accessorKey: "title",
      filterVariant: "text",
      cell: ({ row, getValue }) => {
        return <Link to={`/users/${row?.original?.id}`}>{getValue()}</Link>;
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
        const userRoleIds = R.pipe(
          R.pathOr([], [row?.original?.id || "", "roles"]),
          R.map(R.path(["id"]))
        )(userRolesMap);
        const excluedRoles = R.filter(
          R.compose(R.not, (id) => R.includes(id, userRoleIds), R.path(["id"])),
          roles
        );
        return (
          <MembersActions
            member={row?.original}
            organizationId={orgId}
            excludedRoles={excluedRoles}
            refetch={refetchUsers}
          />
        );
      },
    },
  ];
};

const MembersActions = ({
  member,
  organizationId,
  excludedRoles = [],
  refetch = () => null,
}: {
  member: V1Beta1User;
  organizationId: string;
  excludedRoles: V1Beta1Role[];
  refetch?: () => void;
}) => {
  const { client } = useFrontier();

  async function deleteMember() {
    try {
      // @ts-ignore
      if (member?.invited) {
        await client?.frontierServiceDeleteOrganizationInvitation(
          // @ts-ignore
          member.org_id,
          member?.id as string
        );
      } else {
        await client?.frontierServiceRemoveOrganizationUser(
          organizationId,
          member?.id as string
        );
      }
      toast.success("Member deleted");
    } catch ({ error }: any) {
      toast.error("Something went wrong", {
        description: error.message,
      });
    }
  }
  async function updateRole(role: V1Beta1Role) {
    try {
      const resource = `app/organization:${organizationId}`;
      const principal = `app/user:${member?.id}`;
      const {
        // @ts-ignore
        data: { policies = [] },
      } = await client?.frontierServiceListPolicies({
        org_id: organizationId,
        user_id: member.id,
      });
      const deletePromises = policies.map((p: V1Beta1Policy) =>
        client?.frontierServiceDeletePolicy(p.id as string)
      );

      await Promise.all(deletePromises);
      await client?.frontierServiceCreatePolicy({
        role_id: role.id as string,
        title: role.name as string,
        resource: resource,
        principal: principal,
      });
      refetch();
      toast.success("Member role updated");
    } catch (error: any) {
      toast.error("Something went wrong", {
        description: error?.message,
      });
    }
  }

  return (
    <DropdownMenu style={{ padding: "0 !important" }}>
      <DropdownMenu.Trigger asChild style={{ cursor: "pointer" }}>
        <DotsHorizontalIcon />
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="end">
        <DropdownMenu.Group style={{ padding: 0 }}>
          {excludedRoles.map((role: V1Beta1Role) => (
            <DropdownMenu.Item style={{ padding: 0 }} key={role.id}>
              <Flex
                onClick={() => updateRole(role)}
                style={{ padding: "8px" }}
                gap={"small"}
              >
                <UpdateIcon />
                Make {role.title}
              </Flex>
            </DropdownMenu.Item>
          ))}

          <DropdownMenu.Item style={{ padding: 0 }}>
            <Flex
              onClick={deleteMember}
              style={{ padding: "8px" }}
              gap={"small"}
            >
              <TrashIcon />
              Remove
            </Flex>
          </DropdownMenu.Item>
        </DropdownMenu.Group>
      </DropdownMenu.Content>
    </DropdownMenu>
  );
};
