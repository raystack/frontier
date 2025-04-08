import { AvatarGroup, DataTableColumnDef, Tooltip } from "@raystack/apsara/v1";
import { SearchOrganizationProjectsResponseOrganizationProject } from "~/api/frontier";
import styles from "./projects.module.css";
import { Avatar, Flex, Text } from "@raystack/apsara/v1";
import dayjs from "dayjs";
import { NULL_DATE } from "~/utils/constants";
import { V1Beta1User } from "@raystack/frontier";

export const getColumns = ({
  orgMembersMap,
}: {
  orgMembersMap: Record<string, V1Beta1User>;
}): DataTableColumnDef<
  SearchOrganizationProjectsResponseOrganizationProject,
  unknown
>[] => {
  return [
    {
      accessorKey: "title",
      header: "Title",
      classNames: {
        cell: styles["title-column"],
        header: styles["title-column"],
      },
      cell: ({ row }) => {
        return (
          <Flex gap={4} align="center">
            <Text>{row.original.title || "-"}</Text>
          </Flex>
        );
      },
      enableColumnFilter: true,
      enableSorting: true,
    },
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ getValue }) => {
        return getValue();
      },
      enableColumnFilter: true,
    },
    {
      accessorKey: "user_ids",
      header: "Members",
      enableHiding: true,
      cell: ({ getValue }) => {
        const user_ids = (getValue() as string[]) || [];
        return (
          <AvatarGroup max={4}>
            {user_ids.map((id: string) => {
              const user = orgMembersMap[id];
              const message = user?.title || user?.email || id;
              return (
                <Tooltip message={message} key={id}>
                  <Avatar
                    src={user?.avatar}
                    fallback={message?.[0]}
                    radius="full"
                  />
                </Tooltip>
              );
            })}
          </AvatarGroup>
        );
      },
    },
    {
      accessorKey: "created_at",
      header: "Created On",
      cell: ({ getValue }) => {
        const value = getValue() as string;
        return value !== NULL_DATE ? dayjs(value).format("YYYY-MM-DD") : "-";
      },
      enableSorting: true,
      enableHiding: true,
      enableColumnFilter: true,
      filterType: "date",
    },
  ];
};
