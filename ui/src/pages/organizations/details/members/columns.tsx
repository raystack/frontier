import { DataTableColumnDef } from "@raystack/apsara/v1";
import {
  SearchOrganizationUsersResponseOrganizationUser,
  V1Beta1Role,
} from "~/api/frontier";
import styles from "./members.module.css";
import { Avatar, Flex, Text } from "@raystack/apsara/v1";
import dayjs from "dayjs";
import { NULL_DATE } from "~/utils/constants";

const MemberStates = {
  enabled: "Active",
  disabled: "Suspended",
};

export const getColumns = ({
  roles = [],
}: {
  roles: V1Beta1Role[];
}): DataTableColumnDef<
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
        return (
          <Flex gap={4} align="center">
            <Avatar src={row.original.avatar} fallback={nameInitial} />
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
    },
    {
      accessorKey: "joined_at",
      header: "Joined On",
      cell: ({ getValue }) => {
        const value = getValue() as string;
        return value !== NULL_DATE ? dayjs(value).format("YYYY-MM-DD") : "-";
      },
      enableColumnFilter: true,
    },
  ];
};
