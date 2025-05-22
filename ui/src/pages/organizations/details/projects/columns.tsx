import {
  AvatarGroup,
  getAvatarColor,
  Avatar,
  Flex,
  Text,
  DropdownMenu,
} from "@raystack/apsara/v1";
import type { DataTableColumnDef } from "@raystack/apsara/v1";
import type {
  SearchOrganizationProjectsResponseOrganizationProject,
  V1Beta1User,
} from "~/api/frontier";
import styles from "./projects.module.css";

import dayjs from "dayjs";
import { NULL_DATE } from "~/utils/constants";
import { DotsHorizontalIcon } from "@radix-ui/react-icons";
import { RenameProjectDialog } from "./rename-project";
import { useState } from "react";
import type React from "react";
import Skeleton from "react-loading-skeleton";
import { useAddProjectMembers } from "./useAddProjectMembers";

const DropdownLoader = () => {
  return (
    <>
      <DropdownMenu.Item>
        <Skeleton containerClassName={styles["flex1"]} />
      </DropdownMenu.Item>
      <DropdownMenu.Item>
        <Skeleton containerClassName={styles["flex1"]} />
      </DropdownMenu.Item>
      <DropdownMenu.Item>
        <Skeleton containerClassName={styles["flex1"]} />
      </DropdownMenu.Item>
    </>
  );
};

interface AddMemberDropdownProps {
  onAddMember: (
    userId: string,
  ) => (e: React.MouseEvent<HTMLDivElement>) => void;
  eligibleMembers: V1Beta1User[];
  isLoading: boolean;
  setSearchQuery: (query: string) => void;
}

function AddMemberDropdown({
  onAddMember,
  eligibleMembers,
  isLoading,
  setSearchQuery,
}: AddMemberDropdownProps) {
  return (
    <DropdownMenu
      autocomplete
      autocompleteMode="manual"
      onSearch={setSearchQuery}
    >
      <DropdownMenu.TriggerItem data-test-id="add-members">
        Add member
      </DropdownMenu.TriggerItem>
      <DropdownMenu.Content>
        {isLoading ? (
          <DropdownLoader />
        ) : (
          <>
            {eligibleMembers?.slice(0, 5).map((user) => (
              <DropdownMenu.Item
                key={user.id}
                onClick={onAddMember(user?.id || "")}
                data-test-id={`admin-ui-add-member-${user.id}`}
                leadingIcon={
                  <Avatar
                    src={user.avatar}
                    fallback={user?.title?.[0] || user?.email?.[0]}
                    radius="full"
                    color={getAvatarColor(user.id || "")}
                  />
                }
              >
                <Text>{user.title || user.email}</Text>
              </DropdownMenu.Item>
            ))}
          </>
        )}
      </DropdownMenu.Content>
    </DropdownMenu>
  );
}

function ProjectActionsContent({
  project,
  handleProjectUpdate,
  handleRenameOptionOpen,
}: {
  project: SearchOrganizationProjectsResponseOrganizationProject;
  handleProjectUpdate: (
    project: SearchOrganizationProjectsResponseOrganizationProject,
  ) => void;
  handleRenameOptionOpen: () => void;
}) {
  const handleRenameOptionClick = (e: React.MouseEvent<HTMLDivElement>) => {
    handleRenameOptionOpen();
    e.stopPropagation();
    e.preventDefault();
  };
  const { isLoading, eligibleMembers, setSearchQuery, addMember } =
    useAddProjectMembers({
      projectId: project?.id || "",
    });

  function onAddMember(userId: string) {
    return async (e: React.MouseEvent<HTMLDivElement>) => {
      e.stopPropagation();
      const members = await addMember(userId);
      const userIds = members?.map((user) => user.id || "");
      handleProjectUpdate({ ...project, user_ids: userIds });
    };
  }

  return (
    <>
      <AddMemberDropdown
        onAddMember={onAddMember}
        eligibleMembers={eligibleMembers}
        isLoading={isLoading}
        setSearchQuery={setSearchQuery}
      />
      <DropdownMenu.Item
        onClick={handleRenameOptionClick}
        data-test-id="rename-project"
      >
        Rename project...
      </DropdownMenu.Item>
    </>
  );
}

function ProjectActions({
  project,
  handleProjectUpdate,
}: {
  project: SearchOrganizationProjectsResponseOrganizationProject;
  handleProjectUpdate: (
    project: SearchOrganizationProjectsResponseOrganizationProject,
  ) => void;
}) {
  const [open, setOpen] = useState(false);

  const [isRenameDialogOpen, setIsRenameDialogOpen] = useState(false);

  const preventClickBubbling = (e: React.MouseEvent<SVGElement>) => {
    e.stopPropagation();
  };

  const handleRenameOptionClose = () => {
    setIsRenameDialogOpen(false);
  };

  const handleRenameOptionOpen = () => {
    setIsRenameDialogOpen(true);
  };

  function handleOpen(v: boolean) {
    setOpen(v);
  }

  return (
    <>
      {isRenameDialogOpen ? (
        <RenameProjectDialog
          onClose={handleRenameOptionClose}
          project={project}
          onRename={handleProjectUpdate}
        />
      ) : null}
      <DropdownMenu open={open} setOpen={handleOpen}>
        <DropdownMenu.Trigger asChild>
          <DotsHorizontalIcon
            onClick={preventClickBubbling}
            data-test-id="admin-ui-project-actions"
          />
        </DropdownMenu.Trigger>
        <DropdownMenu.Content
          className={styles["table-action-dropdown"]}
          unmountOnHide={true}
        >
          <ProjectActionsContent
            project={project}
            handleProjectUpdate={handleProjectUpdate}
            handleRenameOptionOpen={handleRenameOptionOpen}
          />
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
                <Avatar
                  key={id}
                  src={user?.avatar}
                  fallback={message?.[0]}
                  radius="full"
                  color={avatarColor}
                />
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
