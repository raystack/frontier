import {
  DataTable,
  DataTableQuery,
  DataTableSort,
  EmptyState,
  Flex,
} from "@raystack/apsara/v1";
import styles from "./tokens.module.css";
import { CoinIcon } from "@raystack/apsara/icons";
import { useCallback, useContext, useEffect } from "react";
import { OrganizationContext } from "../contexts/organization-context";
import PageTitle from "~/components/page-title";
import { api } from "~/api";
import { SearchOrganizationTokensResponseOrganizationToken } from "~/api/frontier";
import { useRQL } from "~/hooks/useRQL";

const NoTokens = () => {
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="No tokens present"
      subHeading="We couldn’t find any matches for that keyword or filter. Try alternative terms or check for typos."
      icon={<CoinIcon />}
    />
  );
};

const DEFAULT_SORT: DataTableSort = { name: "created_at", order: "desc" };

export function OrganizationTokensPage() {
  const { organization, search } = useContext(OrganizationContext);
  const organizationId = organization?.id || "";
  const {
    onChange: onSearchChange,
    setVisibility: setSearchVisibility,
    query: searchQuery,
  } = search;

  const title = `Tokens | ${organization?.title} | Organizations`;

  const apiCallback = useCallback(
    async (apiQuery: DataTableQuery = {}) => {
      const response = await api?.adminServiceSearchOrganizationTokens(
        organizationId,
        { ...apiQuery, search: searchQuery || "" },
      );
      return response?.data;
    },
    [organizationId, searchQuery],
  );

  const columns: Array<any> = [];

  const { data, loading, query, onTableQueryChange, fetchData } =
    useRQL<SearchOrganizationTokensResponseOrganizationToken>({
      initialQuery: { offset: 0 },
      resourceId: organizationId,
      dataKey: "organization_tokens",
      fn: apiCallback,
      searchParam: searchQuery || "",
      onError: (error: Error | unknown) =>
        console.error("Failed to fetch tokens:", error),
    });

  useEffect(() => {
    setSearchVisibility(true);
    return () => {
      onSearchChange("");
      setSearchVisibility(false);
    };
  }, [setSearchVisibility, onSearchChange]);

  return (
    <Flex justify="center">
      <PageTitle title={title} />
      <DataTable
        columns={columns}
        data={data}
        isLoading={loading}
        defaultSort={DEFAULT_SORT}
        mode="server"
        onTableQueryChange={onTableQueryChange}
        onLoadMore={fetchData}
        query={{ ...query, search: searchQuery }}
      >
        <Flex direction="column" style={{ width: "100%" }}>
          <DataTable.Toolbar />
          <DataTable.Content
            emptyState={<NoTokens />}
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
