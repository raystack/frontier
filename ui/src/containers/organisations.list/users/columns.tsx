import {
  DotsHorizontalIcon,
  TrashIcon,
  UpdateIcon,
} from "@radix-ui/react-icons";
import { DropdownMenu, Flex, Text } from "@raystack/apsara/v1";
import {
  V1Beta1ListOrganizationUsersResponseRolePair,
  V1Beta1Policy,
  V1Beta1Role,
  V1Beta1User,
} from "@raystack/frontier";
import type { ColumnDef } from "@tanstack/react-table";
import { Link } from "react-router-dom";
import * as R from "ramda";
import { toast } from "sonner";
import { api } from "~/api";

type UserWithInvitation = V1Beta1User & { isInvited?: boolean };

interface getColumnsOptions {
  users: UserWithInvitation[];
  orgId: string;
  userRolesMap: Record<string, V1Beta1ListOrganizationUsersResponseRolePair>;
  roles: V1Beta1Role[];
  refetchUsers: () => void;
}
export const getColumns: (
  opts: getColumnsOptions
) => ColumnDef<UserWithInvitation, any>[] = ({
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
        return row.original.isInvited ? (
          "Invited"
        ) : (
          <Link to={`/users/${row?.original?.id}`}>{getValue()}</Link>
        );
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
        const excludedRoles = R.filter(
          R.compose(R.not, (id) => R.includes(id, userRoleIds), R.path(["id"])),
          roles
        );
        return row.original.isInvited ? null : (
          <MembersActions
            member={row?.original}
            organizationId={orgId}
            excludedRoles={excludedRoles}
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
  async function deleteMember() {
    try {
      // @ts-ignore
      if (member?.invited) {
        await api?.frontierServiceDeleteOrganizationInvitation(
          // @ts-ignore
          member.org_id,
          member?.id as string
        );
      } else {
        await api?.frontierServiceRemoveOrganizationUser(
          organizationId,
          member?.id as string
        );
      }
      toast.success("Member deleted");
    } catch (error: any) {
      toast.error("Something went wrong", {
        description: error?.message,
      });
    }
  }
  async function updateRole(role: V1Beta1Role) {
    try {
      const resource = `app/organization:${organizationId}`;
      const principal = `app/user:${member?.id}`;
      const resp = await api?.frontierServiceListPolicies({
        org_id: organizationId,
        user_id: member.id,
      });
      const policies = resp?.data?.policies || [];
      const deletePromises = policies.map((p: V1Beta1Policy) =>
        api?.frontierServiceDeletePolicy(p.id as string)
      );

      await Promise.all(deletePromises);
      await api?.frontierServiceCreatePolicy({
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
    // @ts-ignore
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
                data-test-id="admin-ui-role-update-btn"
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
              data-test-id="admin-ui-remove-btn"
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
