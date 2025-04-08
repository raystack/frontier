import {
  DataTable,
  DataTableQuery,
  DataTableSort,
  EmptyState,
  Flex,
} from "@raystack/apsara/v1";
import PageTitle from "~/components/page-title";
import styles from "./projects.module.css";
import { useCallback, useContext, useEffect, useState } from "react";
import { api } from "~/api";
import { getColumns } from "./columns";
import { SearchOrganizationProjectsResponseOrganizationProject } from "~/api/frontier";
import { useDebounceCallback } from "usehooks-ts";
import { OrganizationContext } from "../contexts/organization-context";
import { FileIcon } from "@radix-ui/react-icons";

const LIMIT = 50;
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

  const [data, setData] = useState<
    SearchOrganizationProjectsResponseOrganizationProject[]
  >([]);
  const [isDataLoading, setIsDataLoading] = useState(false);
  const [query, setQuery] = useState<DataTableQuery>({
    offset: 0,
    sort: [DEFAULT_SORT],
  });
  const [nextOffset, setNextOffset] = useState(0);
  const [hasMoreData, setHasMoreData] = useState(true);

  const title = `Projects | ${organization?.title} | Organizations`;

  const fetchProjects = useCallback(
    async (org_id: string, apiQuery: DataTableQuery = {}) => {
      try {
        setIsDataLoading(true);
        const response = await api?.adminServiceSearchOrganizationProjects(
          org_id,
          { ...apiQuery, limit: LIMIT, search: search?.query || "" },
        );
        const data = response.data.org_projects || [];
        setData((prev) => {
          return [...prev, ...data];
        });
        setNextOffset(response.data.pagination?.offset || 0);
        setHasMoreData(data.length !== 0 && data.length === LIMIT);
      } catch (error) {
        console.error(error);
      } finally {
        setIsDataLoading(false);
      }
    },
    [search?.query],
  );

  async function fetchMoreProjects() {
    if (isDataLoading || !hasMoreData || !organizationId) {
      return;
    }
    fetchProjects(organizationId, { offset: nextOffset + LIMIT, ...query });
  }

  const onTableQueryChange = useDebounceCallback((newQuery: DataTableQuery) => {
    setData([]);
    fetchProjects(organizationId, { ...newQuery, offset: 0 });
    setQuery(newQuery);
  }, 500);

  useEffect(() => {
    setSearchVisibility(true);
    return () => {
      onSearchChange("");
      setSearchVisibility(false);
    };
  }, [setSearchVisibility, onSearchChange]);

  const columns = getColumns({ orgMembersMap });

  const isLoading = isOrgMembersMapLoading || isDataLoading;

  return (
    <Flex justify="center" className={styles["container"]}>
      <PageTitle title={title} />
      <DataTable
        columns={columns}
        data={data}
        isLoading={isLoading}
        defaultSort={DEFAULT_SORT}
        mode="server"
        onTableQueryChange={onTableQueryChange}
        onLoadMore={fetchMoreProjects}
        query={{ ...query, search: searchQuery }}
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
  );
}
