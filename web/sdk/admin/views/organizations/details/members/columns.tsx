import {
  DataTableColumnDef,
  getAvatarColor,
  Avatar,
  Flex,
  Text,
  DropdownMenu,
} from "@raystack/apsara";
import type {
  SearchOrganizationUsersResponse_OrganizationUser,
  Role,
} from "@raystack/proton/frontier";
import styles from "./members.module.css";
import dayjs from "dayjs";
import { DotsHorizontalIcon } from "@radix-ui/react-icons";
import {
  isNullTimestamp,
  TimeStamp,
  timestampToDate,
} from "../../../../utils/connect-timestamp";

const MemberStates = {
  enabled: "Active",
  disabled: "Suspended",
};

interface getColumnsOptions {
  roles: Role[];
  handleAssignRoleAction: (
    user: SearchOrganizationUsersResponse_OrganizationUser,
  ) => void;
  handleRemoveMemberAction: (
    user: SearchOrganizationUsersResponse_OrganizationUser,
  ) => void;
}

export const getColumns = ({
  roles = [],
  handleAssignRoleAction,
  handleRemoveMemberAction,
}: getColumnsOptions): DataTableColumnDef<
  SearchOrganizationUsersResponse_OrganizationUser,
  unknown
>[] => {
  const roleMap = roles.reduce(
    (acc, role) => {
      const id = role?.id ?? "";
      acc[id] = role.title || "";
      return acc;
    },
    {} as Record<string, string>,
  );

  return [
    {
      accessorKey: "title",
      header: "Name",
      classNames: {
        cell: styles["name-column"],
        header: styles["name-column"],
      },
      cell: ({ row }) => {
        const nameInitial =
          row.original.title?.[0] || row?.original?.email?.[0];
        const avatarColor = getAvatarColor(row?.original?.id || "");
        return (
          <Flex gap={4} align="center">
            <Avatar
              src={row.original.avatar}
              fallback={nameInitial}
              color={avatarColor}
            />
            <Text>{row.original.title || "-"}</Text>
          </Flex>
        );
      },
      enableColumnFilter: true,
      enableSorting: true,
    },
    {
      accessorKey: "email",
      header: "Email",
      cell: ({ getValue }) => {
        return getValue();
      },
      enableColumnFilter: true,
    },
    {
      accessorKey: "roleIds",
      header: "Role",
      cell: ({ getValue }) => {
        const roleIds = getValue() as string[];
        return roleIds.map((id) => roleMap[id] || "-").join(", ");
      },
      enableColumnFilter: true,
      filterType: "select",
      filterOptions: roles.map((role) => ({
        label: role.title,
        value: role.id,
      })),
    },
    {
      accessorKey: "state",
      header: "Status",
      cell: ({ getValue }) => {
        return MemberStates[getValue() as keyof typeof MemberStates];
      },
      enableColumnFilter: true,
      filterType: "select",
      filterOptions: Object.entries(MemberStates).map(([key, value]) => ({
        label: value,
        value: key,
      })),
      enableHiding: true,
    },
    {
      accessorKey: "orgJoinedAt",
      header: "Joined On",
      cell: ({ getValue }) => {
        const value = getValue() as TimeStamp;
        const date = isNullTimestamp(value)
          ? "-"
          : dayjs(timestampToDate(value)).format("YYYY-MM-DD");
        return <Text>{date}</Text>;
      },
      enableSorting: true,
      enableHiding: true,
      // enableColumnFilter: true,
      // filterType: "date",
    },
    {
      accessorKey: "id",
      header: "",
      classNames: {
        header: styles["table-action-column"],
        cell: styles["table-action-column"],
      },
      cell: ({ row }) => {
        return (
          <DropdownMenu>
            <DropdownMenu.Trigger asChild>
              <DotsHorizontalIcon />
            </DropdownMenu.Trigger>
            <DropdownMenu.Content className={styles["table-action-dropdown"]}>
              <DropdownMenu.Item
                onClick={() => handleAssignRoleAction(row.original)}
                data-test-id="admin-assign-role-action"
              >
                Assign role...
              </DropdownMenu.Item>
              <DropdownMenu.Item
                onClick={() => handleRemoveMemberAction(row.original)}
                data-test-id="admin-remove-member-action"
              >
                Remove...
              </DropdownMenu.Item>
            </DropdownMenu.Content>
          </DropdownMenu>
        );
      },
    },
  ];
};
