import {
  DataTable,
  DataTableQuery,
  Dialog,
  EmptyState,
  Flex,
} from "@raystack/apsara/v1";
import { useCallback, useEffect, useMemo, useState } from "react";
import Skeleton from "react-loading-skeleton";
import { api } from "~/api";
import {
  SearchProjectUsersResponseProjectUser,
  V1Beta1Project,
  V1Beta1Role,
} from "~/api/frontier";
import styles from "./members.module.css";
import UserIcon from "~/assets/icons/users.svg?react";
import { getColumns } from "./columns";
import { useDebounceCallback } from "usehooks-ts";
import { AssignRole } from "./assign-role";
import { PROJECT_NAMESPACE } from "../../types";
import { RemoveMember } from "./remove-member";
import { AddMembersDropdown } from "./add-members-dropdown";

const NoMembers = () => {
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="No Members found"
      subHeading="We couldn’t find any matches for that keyword or filter. Try alternative terms or check for typos."
      icon={<UserIcon />}
    />
  );
};

const LIMIT = 50;

export const ProjectMembersDialog = ({
  projectId,
  onClose,
}: {
  projectId: string;
  onClose: () => void;
}) => {
  const [project, setProject] = useState<V1Beta1Project>({});
  const [isProjectLoading, setIsProjectLoading] = useState<boolean>(false);

  const [projectRoles, setProjectRoles] = useState<V1Beta1Role[]>([]);
  const [isProjectRolesLoading, setIsProjectRolesLoading] =
    useState<boolean>(false);

  const [assignRoleConfig, setAssignRoleConfig] = useState<{
    isOpen: boolean;
    user: SearchProjectUsersResponseProjectUser | null;
  }>({ isOpen: false, user: null });

  const [removeMemberConfig, setRemoveMemberConfig] = useState<{
    isOpen: boolean;
    user: SearchProjectUsersResponseProjectUser | null;
  }>({ isOpen: false, user: null });

  const [query, setQuery] = useState<DataTableQuery>({
    offset: 0,
  });

  const [nextOffset, setNextOffset] = useState(0);
  const [hasMoreData, setHasMoreData] = useState(true);

  const [isProjectMembersLoading, setIsProjectMembersLoading] =
    useState<boolean>(false);
  const [members, setMembers] = useState<
    SearchProjectUsersResponseProjectUser[]
  >([]);

  const fetchProject = useCallback(async (id: string) => {
    setIsProjectLoading(true);
    try {
      const resp = await api?.frontierServiceGetProject(id);
      const project = resp.data?.project || {};
      setProject(project);
    } catch (error) {
      console.error(error);
    } finally {
      setIsProjectLoading(false);
    }
  }, []);

  const fetchProjectRoles = useCallback(async () => {
    setIsProjectRolesLoading(true);
    try {
      const resp = await api?.frontierServiceListRoles({
        scopes: [PROJECT_NAMESPACE],
      });
      const roles = resp.data?.roles || [];
      setProjectRoles(roles);
    } catch (error) {
      console.error(error);
    } finally {
      setIsProjectRolesLoading(false);
    }
  }, []);

  const fetchMembers = useCallback(
    async (apiQuery: DataTableQuery = {}) => {
      setIsProjectMembersLoading(true);
      try {
        delete apiQuery.sort;
        const response = await api?.adminServiceSearchProjectUsers(
          projectId,
          apiQuery,
        );
        const members = response.data?.project_users || [];
        setMembers((prev) => {
          return [...prev, ...members];
        });
        setNextOffset(response.data.pagination?.offset || 0);
        setHasMoreData(members.length !== 0 && members.length === LIMIT);
      } catch (error) {
        console.error(error);
      } finally {
        setIsProjectMembersLoading(false);
      }
    },
    [projectId],
  );

  useEffect(() => {
    if (projectId) {
      fetchProject(projectId);
      fetchProjectRoles();
    }
  }, [projectId, fetchProject, fetchProjectRoles]);

  function openAssignRoleDialog(user: SearchProjectUsersResponseProjectUser) {
    setAssignRoleConfig({ isOpen: true, user });
  }

  function closeAssignRoleDialog() {
    setAssignRoleConfig({ isOpen: false, user: null });
  }

  function openRemoveMemberDialog(user: SearchProjectUsersResponseProjectUser) {
    setRemoveMemberConfig({ isOpen: true, user });
  }

  function closeRemoveMemberDialog() {
    setRemoveMemberConfig({ isOpen: false, user: null });
  }

  const columns = useMemo(
    () =>
      getColumns({
        roles: projectRoles,
        handleAssignRoleAction: openAssignRoleDialog,
        handleRemoveAction: openRemoveMemberDialog,
      }),
    [projectRoles],
  );

  const onTableQueryChange = useDebounceCallback((newQuery: DataTableQuery) => {
    setMembers([]);
    fetchMembers({ ...newQuery, offset: 0 });
    setQuery(newQuery);
  }, 500);

  async function fetchMoreMembers() {
    if (isProjectMembersLoading || !hasMoreData) {
      return;
    }
    fetchMembers({ ...query, offset: nextOffset + LIMIT });
  }

  async function removeMember(user: SearchProjectUsersResponseProjectUser) {
    setMembers((prevMembers) => {
      return prevMembers.filter((member) => member.id !== user.id);
    });
    setRemoveMemberConfig({ isOpen: false, user: null });
  }

  async function updateMember(user: SearchProjectUsersResponseProjectUser) {
    setMembers((prevMembers) => {
      const updatedMembers = prevMembers.map((member) =>
        member.id === user.id ? user : member,
      );
      return updatedMembers;
    });
    setAssignRoleConfig({ isOpen: false, user: null });
  }

  const isLoading =
    isProjectMembersLoading || isProjectLoading || isProjectRolesLoading;

  return (
    <>
      {assignRoleConfig.isOpen && assignRoleConfig.user ? (
        <AssignRole
          roles={projectRoles}
          user={assignRoleConfig.user}
          projectId={projectId}
          onRoleUpdate={updateMember}
          onClose={closeAssignRoleDialog}
        />
      ) : null}
      {removeMemberConfig.isOpen && removeMemberConfig.user ? (
        <RemoveMember
          projectId={projectId}
          user={removeMemberConfig.user}
          onRemove={removeMember}
          onClose={closeRemoveMemberDialog}
        />
      ) : null}
      <Dialog open onOpenChange={onClose}>
        <Dialog.Content className={styles["dialog-content"]}>
          <Dialog.Header>
            {isProjectLoading ? (
              <Skeleton containerClassName={styles["flex1"]} width={"200px"} />
            ) : (
              <Dialog.Title>{project.title}</Dialog.Title>
            )}
            <Dialog.CloseButton data-test-id="close-button" />
          </Dialog.Header>
          <Dialog.Body className={styles["dialog-body"]}>
            <DataTable
              data={members}
              columns={columns}
              isLoading={isLoading}
              mode="server"
              defaultSort={{ name: "", order: "desc" }}
              onTableQueryChange={onTableQueryChange}
              onLoadMore={fetchMoreMembers}
            >
              <Flex
                direction="column"
                gap={5}
                className={styles["table-content-wrapper"]}
              >
                <Flex>
                  <DataTable.Search className={styles["table-search"]} />
                  <AddMembersDropdown projectId={projectId} />
                </Flex>
                <DataTable.Content
                  emptyState={<NoMembers />}
                  classNames={{
                    table: styles["table"],
                    root: styles["table-wrapper"],
                    header: styles["table-header"],
                  }}
                />
              </Flex>
            </DataTable>
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>
    </>
  );
};
