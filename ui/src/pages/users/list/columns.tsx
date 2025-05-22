import {
  Avatar,
  DataTableColumnDef,
  EmptyFilterValue,
  Flex,
  getAvatarColor,
  Text,
} from "@raystack/apsara/v1";
import { V1Beta1User } from "@raystack/frontier";
import dayjs from "dayjs";
import styles from "./list.module.css";
import { NULL_DATE } from "~/utils/constants";
import { getUserName, USER_STATES, UserState } from "../util";

interface getColumnsOptions {
  groupCountMap: Record<string, Record<string, number>>;
}

export const getColumns = ({
  groupCountMap,
}: getColumnsOptions): DataTableColumnDef<V1Beta1User, unknown>[] => {
  return [
    {
      accessorKey: "title",
      header: "Name",
      classNames: {
        cell: styles["name-column"],
        header: styles["name-column"],
      },
      cell: ({ row }) => {
        const avatarColor = getAvatarColor(row?.original?.id || "");
        const name = getUserName(row.original);
        return (
          <Flex gap={4} align="center">
            <Avatar
              src={row.original.avatar}
              fallback={name?.[0]?.toUpperCase()}
              color={avatarColor}
            />
            <Text>{name}</Text>
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
      accessorKey: "created_at",
      header: "Joined on",
      filterType: "date",
      cell: ({ getValue }) => {
        const value = getValue() as string;
        return value !== NULL_DATE ? dayjs(value).format("YYYY-MM-DD") : "-";
      },
      enableHiding: true,
      enableSorting: true,
    },
    {
      accessorKey: "state",
      header: "Status",
      cell: ({ getValue }) => {
        return USER_STATES?.[getValue() as UserState] ?? "-";
      },
      filterType: "select",
      filterOptions: Object.entries(USER_STATES).map(([value, label]) => ({
        value: value === "" ? EmptyFilterValue : value,
        label,
      })),
      enableColumnFilter: true,
      enableHiding: true,
      enableSorting: true,
      enableGrouping: true,
      showGroupCount: true,
      groupCountMap: groupCountMap["state"] || {},
      groupLabelsMap: USER_STATES,
    },
  ];
};
