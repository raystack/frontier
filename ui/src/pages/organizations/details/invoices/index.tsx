import {
  DataTable,
  DataTableQuery,
  DataTableSort,
  EmptyState,
  Flex,
} from "@raystack/apsara/v1";
import styles from "./invoices.module.css";
import { FileTextIcon } from "@radix-ui/react-icons";
import { useCallback, useContext, useEffect, useState } from "react";
import { OrganizationContext } from "../contexts/organization-context";
import PageTitle from "~/components/page-title";
import { getColumns } from "./columns";
import { api } from "~/api";
import { SearchOrganizationInvoicesResponseOrganizationInvoice } from "~/api/frontier";
import { useDebounceCallback } from "usehooks-ts";

const NoInvoices = () => {
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="No Invoice found"
      subHeading="We couldn’t find any matches for that keyword or filter. Try alternative terms or check for typos."
      // TODO: update icon with raystack icon
      icon={<FileTextIcon />}
    />
  );
};

const LIMIT = 50;
const DEFAULT_SORT: DataTableSort = { name: "created_at", order: "desc" };

export function OrganizationInvoicesPage() {
  const { organization, search } = useContext(OrganizationContext);
  const organizationId = organization?.id || "";

  const {
    onChange: onSearchChange,
    setVisibility: setSearchVisibility,
    query: searchQuery,
  } = search;

  const title = `Invoices | ${organization?.title} | Organizations`;

  const [query, setQuery] = useState<DataTableQuery>({
    offset: 0,
  });

  const [isDataLoading, setIsDataLoading] = useState(false);
  const [data, setData] = useState<
    SearchOrganizationInvoicesResponseOrganizationInvoice[]
  >([]);
  const [nextOffset, setNextOffset] = useState(0);
  const [hasMoreData, setHasMoreData] = useState(true);

  const fetchInvoices = useCallback(
    async (org_id: string, apiQuery: DataTableQuery = {}) => {
      try {
        setIsDataLoading(true);
        const response = await api?.adminServiceSearchOrganizationInvoices(
          org_id,
          { ...apiQuery, limit: LIMIT, search: search?.query || "" },
        );
        const invoices = response.data.organization_invoices || [];
        setData((prev) => {
          return [...prev, ...invoices];
        });
        setNextOffset(response.data.pagination?.offset || 0);
        setHasMoreData(invoices.length !== 0 && invoices.length === LIMIT);
      } catch (error) {
        console.error(error);
      } finally {
        setIsDataLoading(false);
      }
    },
    [search?.query],
  );

  async function fetchMoreInvoices() {
    if (isDataLoading || !hasMoreData || !organizationId) {
      return;
    }
    fetchInvoices(organizationId, { ...query, offset: nextOffset + LIMIT });
  }

  useEffect(() => {
    setSearchVisibility(true);
    return () => {
      onSearchChange("");
      setSearchVisibility(false);
    };
  }, [setSearchVisibility, onSearchChange]);

  const onTableQueryChange = useDebounceCallback((newQuery: DataTableQuery) => {
    setData([]);
    fetchInvoices(organizationId, { ...newQuery, offset: 0 });
    setQuery(newQuery);
  }, 500);

  const columns = getColumns();
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
        onLoadMore={fetchMoreInvoices}
        query={{ ...query, search: searchQuery }}
      >
        <Flex direction="column" style={{ width: "100%" }}>
          <DataTable.Toolbar />
          <DataTable.Content
            emptyState={<NoInvoices />}
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
