import {
  Avatar,
  Flex,
  getAvatarColor,
  Text,
  Tooltip,
  type DataTableColumnDef,
} from "@raystack/apsara";
import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import type { PAT, Project, User } from "@raystack/proton/frontier";
import {
  isNullTimestamp,
  timestampToDayjs,
} from "../../../../utils/connect-timestamp";
import { SCOPES } from "../../../../utils/constants";
import styles from "./pat.module.css";

dayjs.extend(relativeTime);

interface GetColumnsOptions {
  orgMembersMap: Record<string, User>;
  projectsMap: Record<string, Project>;
}

const DATE_FORMAT = "DD MMM YYYY";

export function getColumns({
  orgMembersMap,
  projectsMap,
}: GetColumnsOptions): DataTableColumnDef<PAT, unknown>[] {
  return [
    {
      accessorKey: "title",
      header: "Title",
      classNames: {
        cell: styles["first-column"],
        header: styles["first-column"],
      },
      cell: ({ getValue }) => {
        const value = (getValue() as string) || "";
        return <Text className={styles["truncate-text"]}>{value}</Text>;
      },
      enableColumnFilter: true,
    },
    {
      accessorKey: "scopes",
      header: "Project",
      classNames: { cell: styles["truncate-cell"] },
      enableSorting: false,
      cell: ({ row }) => {
        const projectScope = row.original.scopes?.find(
          (scope) => scope.resourceType === SCOPES.PROJECT,
        );
        const resourceIds = projectScope?.resourceIds ?? [];
        if (resourceIds.length === 0) {
          return <Text className={styles["truncate-text"]}>-</Text>;
        }
        const projectNamesText = resourceIds.map(
          (id) => projectsMap[id]?.title || projectsMap[id]?.name || id,
        ).join(", ");
        return (
          <Tooltip
            message={projectNamesText}
            contentStyle={{ maxWidth: "600px" }}
            showArrow={false}
            side="top-left"
          >
            <Text className={styles["truncate-text"]}>{projectNamesText}</Text>
          </Tooltip>
        );
      },
    },
    {
      accessorKey: "userId",
      header: "Created By",
      enableSorting: false,
      cell: ({ getValue }) => {
        const userId = (getValue() as string) || "";
        const user = orgMembersMap[userId];
        const title = user?.title || user?.name || user?.email || userId;
        const avatarColor = getAvatarColor(userId);
        return (
          <Flex gap={4} align="center">
            <Avatar
              src={user?.avatar}
              fallback={title?.[0]}
              color={avatarColor}
            />
            <Text className={styles["truncate-text"]}>{title}</Text>
          </Flex>
        );
      },
    },
    {
      accessorKey: "createdAt",
      header: "Created On",
      styles: { header: { width: "152px" } },
      cell: ({ row }) => {
        const date = timestampToDayjs(row.original.createdAt);
        return date ? <Text>{date.format(DATE_FORMAT)}</Text> : <Text>-</Text>;
      },
      enableSorting: true,
      enableColumnFilter: true,
      filterType: "date",
    },
    {
      accessorKey: "expiresAt",
      header: "Expiry Date",
      styles: { header: { width: "152px" } },
      cell: ({ row }) => {
        const expiresAt = row.original.expiresAt;
        if (!expiresAt || isNullTimestamp(expiresAt)) return <Text>-</Text>;
        const date = timestampToDayjs(expiresAt);
        return date ? <Text>{date.format(DATE_FORMAT)}</Text> : <Text>-</Text>;
      },
      enableSorting: true,
      enableColumnFilter: true,
      filterType: "date",
    },
    {
      accessorKey: "usedAt",
      header: "Last used",
      styles: { header: { width: "152px" } },
      enableSorting: false,
      cell: ({ row }) => {
        const usedAt = row.original.usedAt;
        if (!usedAt || isNullTimestamp(usedAt)) return <Text>-</Text>;
        const date = timestampToDayjs(usedAt);
        return date ? <Text>{date.fromNow()}</Text> : <Text>-</Text>;
      },
    },
  ];
}
