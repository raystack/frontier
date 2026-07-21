import {
  AvatarGroup,
  getAvatarColor,
  Avatar,
  Flex,
  Text,
  Menu,
} from "@raystack/apsara";
import type { DataTableColumnDef } from "@raystack/apsara";
import type {
  SearchOrganizationProjectsResponse_OrganizationProject,
  User,
} from "@raystack/proton/frontier";
import styles from "./projects.module.css";

import dayjs from "dayjs";
import {
  isNullTimestamp,
  TimeStamp,
  timestampToDate,
} from "../../../../utils/connect-timestamp";
import { DotsHorizontalIcon } from "@radix-ui/react-icons";
import { RenameProjectDialog } from "./rename-project";
import { useState } from "react";
import type React from "react";
import Skeleton from "react-loading-skeleton";
import { useAddProjectMembers } from "./use-add-project-members";
import { useTerminology, TerminologyEntity } from "../../../../hooks/useTerminology";

const DropdownLoader = () => {
  return (
    <>
      <Menu.Item>
        <Skeleton containerClassName={styles["flex1"]} />
      </Menu.Item>
      <Menu.Item>
        <Skeleton containerClassName={styles["flex1"]} />
      </Menu.Item>
      <Menu.Item>
        <Skeleton containerClassName={styles["flex1"]} />
      </Menu.Item>
    </>
  );
};

interface AddMemberDropdownProps {
  onAddMember: (
    userId: string,
  ) => (e: React.MouseEvent<HTMLDivElement>) => void;
  eligibleMembers: User[];
  isLoading: boolean;
  setSearchQuery: (query: string) => void;
  disabled?: boolean;
}

function AddMemberDropdown({
  onAddMember,
  eligibleMembers,
  isLoading,
  setSearchQuery,
  disabled,
}: AddMemberDropdownProps) {
  const t = useTerminology();
  return (
    <Menu.Submenu
      autocomplete
      autocompleteMode="manual"
      onInputValueChange={setSearchQuery}
    >
      <Menu.SubmenuTrigger disabled={disabled} data-test-id="add-members">
        Add {t.member({ case: "lower" })}
      </Menu.SubmenuTrigger>
      <Menu.SubmenuContent>
        {isLoading ? (
          <DropdownLoader />
        ) : (
          <>
            {eligibleMembers?.slice(0, 5).map((user) => (
              <Menu.Item
                key={user.id}
                onClick={onAddMember(user?.id || "")}
                data-test-id={`admin-add-member-${user.id}`}
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
              </Menu.Item>
            ))}
          </>
        )}
      </Menu.SubmenuContent>
    </Menu.Submenu>
  );
}

function ProjectActionsContent({
  project,
  handleProjectUpdate,
  handleRenameOptionOpen,
  canAddMember,
}: {
  project: SearchOrganizationProjectsResponse_OrganizationProject;
  handleProjectUpdate: (
    project: SearchOrganizationProjectsResponse_OrganizationProject,
  ) => void;
  handleRenameOptionOpen: () => void;
  canAddMember: boolean;
}) {
  const t = useTerminology();
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
      handleProjectUpdate({ ...project, userIds: userIds || [] });
    };
  }

  return (
    <>
      <AddMemberDropdown
        onAddMember={onAddMember}
        eligibleMembers={eligibleMembers}
        isLoading={isLoading}
        setSearchQuery={setSearchQuery}
        disabled={!canAddMember}
      />
      <Menu.Item
        onClick={handleRenameOptionClick}
        data-test-id="rename-project"
      >
        Rename {t.project({ case: "lower" })}...
      </Menu.Item>
    </>
  );
}

function ProjectActions({
  project,
  handleProjectUpdate,
  canAddMember,
}: {
  project: SearchOrganizationProjectsResponse_OrganizationProject;
  handleProjectUpdate: (
    project: SearchOrganizationProjectsResponse_OrganizationProject,
  ) => void;
  canAddMember: boolean;
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
      <Menu open={open} onOpenChange={handleOpen}>
        <Menu.Trigger
          render={
            <DotsHorizontalIcon
              onClick={preventClickBubbling}
              data-test-id="admin-project-actions"
            />
          }
        />
        <Menu.Content className={styles["table-action-dropdown"]}>
          <ProjectActionsContent
            project={project}
            handleProjectUpdate={handleProjectUpdate}
            handleRenameOptionOpen={handleRenameOptionOpen}
            canAddMember={canAddMember}
          />
        </Menu.Content>
      </Menu>
    </>
  );
}

export const getColumns = ({
  orgMembersMap,
  handleProjectUpdate,
  t,
}: {
  orgMembersMap: Record<string, User>;
  handleProjectUpdate: (
    project: SearchOrganizationProjectsResponse_OrganizationProject,
  ) => void;
  t: {
    member: TerminologyEntity;
  };
}): DataTableColumnDef<
  SearchOrganizationProjectsResponse_OrganizationProject,
  unknown
>[] => {
  // Adding a project member requires searching org members. With a single org
  // member there is no one to add, so the action is disabled until more members
  // are invited.
  const canAddMember = Object.keys(orgMembersMap).length > 1;
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
      accessorKey: "userIds",
      header: t.member({ plural: true, case: "capital" }),
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
      accessorKey: "createdAt",
      header: "Created On",
      cell: ({ getValue }) => {
        const value = getValue() as TimeStamp;
        const date = isNullTimestamp(value)
          ? "-"
          : dayjs(timestampToDate(value)).format("YYYY-MM-DD");
        return date;
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
            canAddMember={canAddMember}
          />
        );
      },
    },
  ];
};
