import {
  DataTableColumnDef,
  getAvatarColor,
  Avatar,
  Flex,
  Text,
  DropdownMenu,
} from "@raystack/apsara";
import {
  SearchOrganizationUsersResponseOrganizationUser,
  V1Beta1Role,
} from "~/api/frontier";
import styles from "./members.module.css";
import dayjs from "dayjs";
import { NULL_DATE } from "~/utils/constants";
import { DotsHorizontalIcon } from "@radix-ui/react-icons";

const MemberStates = {
  enabled: "Active",
  disabled: "Suspended",
};

interface getColumnsOptions {
  roles: V1Beta1Role[];
  handleAssignRoleAction: (
    user: SearchOrganizationUsersResponseOrganizationUser,
  ) => void;
  handleRemoveMemberAction: (
    user: SearchOrganizationUsersResponseOrganizationUser,
  ) => void;
}

export const getColumns = ({
  roles = [],
  handleAssignRoleAction,
  handleRemoveMemberAction,
}: getColumnsOptions): DataTableColumnDef<
  SearchOrganizationUsersResponseOrganizationUser,
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
      accessorKey: "role_ids",
      header: "Role",
      cell: ({ getValue }) => {
        return roleMap[getValue() as string] || "-";
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
      accessorKey: "org_joined_at",
      header: "Joined On",
      cell: ({ getValue }) => {
        const value = getValue() as string;
        return value !== NULL_DATE ? dayjs(value).format("YYYY-MM-DD") : "-";
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
                data-test-id="admin-ui-assign-role-action"
              >
                Assign role...
              </DropdownMenu.Item>
              <DropdownMenu.Item
                onClick={() => handleRemoveMemberAction(row.original)}
                data-test-id="admin-ui-remove-member-action"
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
