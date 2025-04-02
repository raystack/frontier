import {
  DataTable,
  DataTableQuery,
  DataTableSort,
  EmptyState,
  Flex,
} from "@raystack/apsara/v1";
import PageTitle from "~/components/page-title";
import styles from "./members.module.css";
import { useCallback, useContext, useState } from "react";
import { api } from "~/api";
import { getColumns } from "./columns";
import { SearchOrganizationUsersResponseOrganizationUser } from "~/api/frontier";
import UserIcon from "~/assets/icons/users.svg?react";
import { useDebounceCallback } from "usehooks-ts";
import { OrganizationContext } from "../contexts/organization-context";

const LIMIT = 50;
const DEFAULT_SORT: DataTableSort = { name: "org_joined_at", order: "desc" };

const NoMembers = () => {
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="No Member found"
      subHeading="We couldn’t find any matches for that keyword or filter. Try alternative terms or check for typos."
      icon={<UserIcon />}
    />
  );
};

export function OrganizationMembersPage() {
  const { roles = [], organization } = useContext(OrganizationContext);

  const organizationId = organization?.id || "";

  const [data, setData] = useState<
    SearchOrganizationUsersResponseOrganizationUser[]
  >([]);
  const [isDataLoading, setIsDataLoading] = useState(false);
  const [query, setQuery] = useState<DataTableQuery>({
    offset: 0,
    sort: [DEFAULT_SORT],
  });
  const [nextOffset, setNextOffset] = useState(0);
  const [hasMoreData, setHasMoreData] = useState(true);

  const title = `Members | ${organization?.title} | Organizations`;

  const fetchMembers = useCallback(
    async (org_id: string, apiQuery: DataTableQuery = {}) => {
      try {
        setIsDataLoading(true);
        const response = await api?.adminServiceSearchOrganizationUsers(
          org_id,
          { ...apiQuery, limit: LIMIT },
        );
        const members = response.data.org_users || [];
        setData((prev) => {
          return [...prev, ...members];
        });
        setNextOffset(response.data.pagination?.offset || 0);
        setHasMoreData(members.length !== 0 && members.length === LIMIT);
      } catch (error) {
        console.error(error);
      } finally {
        setIsDataLoading(false);
      }
    },
    [],
  );

  async function fetchMoreMembers() {
    if (isDataLoading || !hasMoreData || !organizationId) {
      return;
    }
    fetchMembers(organizationId, { offset: nextOffset + LIMIT, ...query });
  }

  const onTableQueryChange = useDebounceCallback((newQuery: DataTableQuery) => {
    setData([]);
    fetchMembers(organizationId, { ...newQuery, offset: 0 });
    setQuery(newQuery);
  }, 500);

  const columns = getColumns({ roles });

  return (
    <Flex justify="center" className={styles["container"]}>
      <PageTitle title={title} />
      <DataTable
        columns={columns}
        data={data}
        isLoading={isDataLoading}
        defaultSort={DEFAULT_SORT}
        mode="server"
        onTableQueryChange={onTableQueryChange}
        onLoadMore={fetchMoreMembers}
      >
        <Flex direction="column" style={{ width: "100%" }}>
          <DataTable.Toolbar />
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
    </Flex>
  );
}
