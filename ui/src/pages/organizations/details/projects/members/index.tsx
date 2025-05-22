import { DataTable, Dialog, EmptyState, Flex } from "@raystack/apsara/v1";
import type { DataTableQuery, DataTableSort } from "@raystack/apsara/v1";
import { useCallback, useEffect, useMemo, useState } from "react";
import Skeleton from "react-loading-skeleton";
import { api } from "~/api";
import type {
  SearchProjectUsersResponseProjectUser,
  V1Beta1Project,
  V1Beta1Role,
} from "~/api/frontier";
import styles from "./members.module.css";
import UserIcon from "~/assets/icons/users.svg?react";
import { getColumns } from "./columns";
import { AssignRole } from "./assign-role";
import { PROJECT_NAMESPACE } from "../../types";
import { RemoveMember } from "./remove-member";
import { AddMembersDropdown } from "./add-members-dropdown";
import { useRQL } from "~/hooks/useRQL";

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

const DEFAULT_SORT: DataTableSort = { name: "", order: "desc" };

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

  const apiCallback = useCallback(
    async (apiQuery: DataTableQuery = {}) => {
      delete apiQuery.sort;
      const response = await api?.adminServiceSearchProjectUsers(
        projectId,
        apiQuery,
      );
      return response.data;
    },
    [projectId],
  );

  const { data, setData, loading, query, onTableQueryChange, fetchMore } =
    useRQL<SearchProjectUsersResponseProjectUser>({
      initialQuery: { offset: 0 },
      key: projectId,
      dataKey: "project_users",
      fn: apiCallback,
      onError: (error: Error | unknown) =>
        console.error("Failed to fetch project users:", error),
    });

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

  async function removeMember(user: SearchProjectUsersResponseProjectUser) {
    setData((prevMembers) => {
      return prevMembers.filter((member) => member.id !== user.id);
    });
    setRemoveMemberConfig({ isOpen: false, user: null });
  }

  async function updateMember(user: SearchProjectUsersResponseProjectUser) {
    setData((prevMembers) => {
      const updatedMembers = prevMembers.map((member) =>
        member.id === user.id ? user : member,
      );
      return updatedMembers;
    });
    setAssignRoleConfig({ isOpen: false, user: null });
  }

  const isLoading = loading || isProjectLoading || isProjectRolesLoading;

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
              query={query}
              data={data}
              columns={columns}
              isLoading={isLoading}
              mode="server"
              defaultSort={DEFAULT_SORT}
              onTableQueryChange={onTableQueryChange}
              onLoadMore={fetchMore}
            >
              <Flex
                direction="column"
                gap={5}
                className={styles["table-content-wrapper"]}
              >
                <Flex>
                  <DataTable.Search className={styles["table-search"]} />
                  <AddMembersDropdown
                    projectId={projectId}
                    refetchMembers={refetchMembers}
                  />
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
