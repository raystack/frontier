import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import type { DataTableQuery, DataTableSort } from "@raystack/apsara";
import PageTitle from "~/components/page-title";
import styles from "./projects.module.css";
import { useCallback, useContext, useEffect, useState } from "react";
import { api } from "~/api";
import { getColumns } from "./columns";
import type {
  SearchOrganizationProjectsResponseOrganizationProject,
  Frontierv1Beta1Project,
} from "~/api/frontier";
import { OrganizationContext } from "../contexts/organization-context";
import { FileIcon } from "@radix-ui/react-icons";
import { ProjectMembersDialog } from "./members";
import { useRQL } from "~/hooks/useRQL";

const DEFAULT_SORT: DataTableSort = { name: "created_at", order: "desc" };

const NoProjects = () => {
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="No Projects found"
      subHeading="We couldn’t find any matches for that keyword or filter. Try alternative terms or check for typos."
      icon={<FileIcon />}
    />
  );
};

export function OrganizationProjectssPage() {
  const { organization, search, orgMembersMap, isOrgMembersMapLoading } =
    useContext(OrganizationContext);
  const {
    onChange: onSearchChange,
    setVisibility: setSearchVisibility,
    query: searchQuery,
  } = search;

  const organizationId = organization?.id || "";

  const [memberDialogConfig, setMemberDialogConfig] = useState({
    open: false,
    projectId: "",
  });

  const title = `Projects | ${organization?.title} | Organizations`;

  const apiCallback = useCallback(
    async (apiQuery: DataTableQuery = {}) => {
      const response = await api?.adminServiceSearchOrganizationProjects(
        organizationId,
        { ...apiQuery, search: searchQuery || "" },
      );
      return response?.data;
    },
    [organizationId, searchQuery],
  );

  const { data, setData, loading, query, onTableQueryChange, fetchMore } =
    useRQL<SearchOrganizationProjectsResponseOrganizationProject>({
      initialQuery: { offset: 0 },
      key: organizationId,
      dataKey: "org_projects",
      fn: apiCallback,
      searchParam: searchQuery || "",
      onError: (error: Error | unknown) =>
        console.error("Failed to fetch tokens:", error),
    });

  function handleProjectUpdate(
    project: SearchOrganizationProjectsResponseOrganizationProject,
  ) {
    setData((prev) => prev.map((p) => (p.id === project.id ? project : p)));
  }

  useEffect(() => {
    setSearchVisibility(true);
    return () => {
      onSearchChange("");
      setSearchVisibility(false);
    };
  }, [setSearchVisibility, onSearchChange]);

  function handleMemberDialogOpen(project: Frontierv1Beta1Project) {
    setMemberDialogConfig({
      projectId: project.id || "",
      open: true,
    });
  }

  function handleMemberDialogClose() {
    setMemberDialogConfig({
      projectId: "",
      open: false,
    });
  }

  const columns = getColumns({ orgMembersMap, handleProjectUpdate });

  const isLoading = isOrgMembersMapLoading || loading;

  return (
    <>
      {memberDialogConfig.open && memberDialogConfig.projectId ? (
        <ProjectMembersDialog
          projectId={memberDialogConfig.projectId}
          onClose={handleMemberDialogClose}
        />
      ) : null}
      <Flex justify="center" className={styles["container"]}>
        <PageTitle title={title} />
        <DataTable
          columns={columns}
          data={data}
          isLoading={isLoading}
          defaultSort={DEFAULT_SORT}
          mode="server"
          onTableQueryChange={onTableQueryChange}
          onLoadMore={fetchMore}
          query={{ ...query, search: searchQuery }}
          onRowClick={handleMemberDialogOpen}
        >
          <Flex direction="column" style={{ width: "100%" }}>
            <DataTable.Toolbar />
            <DataTable.Content
              emptyState={<NoProjects />}
              classNames={{
                table: styles["table"],
                root: styles["table-wrapper"],
                header: styles["table-header"],
              }}
            />
          </Flex>
        </DataTable>
      </Flex>
    </>
  );
}
