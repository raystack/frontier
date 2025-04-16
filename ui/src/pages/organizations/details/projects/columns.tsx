import {
  AvatarGroup,
  DataTableColumnDef,
  getAvatarColor,
  Tooltip,
  Avatar,
  Flex,
  Text,
  DropdownMenu,
} from "@raystack/apsara/v1";
import { SearchOrganizationProjectsResponseOrganizationProject } from "~/api/frontier";
import styles from "./projects.module.css";

import dayjs from "dayjs";
import { NULL_DATE } from "~/utils/constants";
import { V1Beta1User } from "@raystack/frontier";
import { DotsHorizontalIcon } from "@radix-ui/react-icons";
import { RenameProjectDialog } from "./rename-project";
import React, { useState } from "react";

function ProjectActions({
  project,
  handleProjectUpdate,
}: {
  project: SearchOrganizationProjectsResponseOrganizationProject;
  handleProjectUpdate: (
    project: SearchOrganizationProjectsResponseOrganizationProject,
  ) => void;
}) {
  const [isRenameDialogOpen, setIsRenameDialogOpen] = useState(false);

  const handleRenameOptionClick = (e: React.MouseEvent<HTMLDivElement>) => {
    setIsRenameDialogOpen(true);
    e.stopPropagation();
    e.preventDefault();
  };

  const handleRenameOptionClose = () => {
    setIsRenameDialogOpen(false);
  };

  return (
    <>
      {isRenameDialogOpen ? (
        <RenameProjectDialog
          onClose={handleRenameOptionClose}
          project={project}
          onRename={handleProjectUpdate}
        />
      ) : null}
      <DropdownMenu>
        <DropdownMenu.Trigger asChild>
          <DotsHorizontalIcon />
        </DropdownMenu.Trigger>
        <DropdownMenu.Content
          className={styles["table-action-dropdown"]}
          align="end"
        >
          <DropdownMenu.Item
            onClick={handleRenameOptionClick}
            data-test-id="rename-project"
          >
            Rename project...
          </DropdownMenu.Item>
        </DropdownMenu.Content>
      </DropdownMenu>
    </>
  );
}

export const getColumns = ({
  orgMembersMap,
  handleProjectUpdate,
}: {
  orgMembersMap: Record<string, V1Beta1User>;
  handleProjectUpdate: (
    project: SearchOrganizationProjectsResponseOrganizationProject,
  ) => void;
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
              const avatarColor = getAvatarColor(user?.id || "");
              return (
                <Tooltip message={message} key={id}>
                  <Avatar
                    src={user?.avatar}
                    fallback={message?.[0]}
                    radius="full"
                    color={avatarColor}
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
    {
      accessorKey: "id",
      header: "",
      classNames: {
        header: styles["table-action-column"],
        cell: styles["table-action-column"],
      },
      cell: ({ row }) => {
        return (
          <ProjectActions
            project={row?.original}
            handleProjectUpdate={handleProjectUpdate}
          />
        );
      },
    },
  ];
};
