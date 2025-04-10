import {
  DataTable,
  DataTableQuery,
  DataTableSort,
  Dialog,
  EmptyState,
} from "@raystack/apsara/v1";
import { useCallback, useEffect, useMemo, useState } from "react";
import Skeleton from "react-loading-skeleton";
import { api } from "~/api";
import {
  SearchProjectUsersResponseProjectUser,
  V1Beta1Project,
} from "~/api/frontier";
import styles from "./members.module.css";
import UserIcon from "~/assets/icons/users.svg?react";
import { getColumns } from "./columns";
import { useDebounceCallback } from "usehooks-ts";

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
    }
  }, [projectId, fetchProject]);

  const columns = useMemo(() => getColumns(), []);

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

  const isLoading = isProjectMembersLoading || isProjectLoading;

  return (
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
        <Dialog.Body>
          <DataTable
            data={members}
            columns={columns}
            isLoading={isLoading}
            mode="server"
            defaultSort={{ name: "", order: "desc" }}
            onTableQueryChange={onTableQueryChange}
            onLoadMore={fetchMoreMembers}
          >
            <DataTable.Content emptyState={<NoMembers />} />
          </DataTable>
        </Dialog.Body>
      </Dialog.Content>
    </Dialog>
  );
};
