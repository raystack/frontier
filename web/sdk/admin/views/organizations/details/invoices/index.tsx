import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import type { DataTableQuery, DataTableSort } from "@raystack/apsara";
import styles from "./invoices.module.css";
import { FileTextIcon, ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { useContext, useEffect, useMemo, useState } from "react";
import { OrganizationContext } from "../contexts/organization-context";
import { PageTitle } from "../../../../components/PageTitle";
import { getColumns } from "./columns";
import { FrontierServiceQueries } from "@raystack/proton/frontier";
import { useInfiniteQuery } from "@connectrpc/connect-query";
import {
  getConnectNextPageParam,
  DEFAULT_PAGE_SIZE,
  getGroupCountMapFromFirstPage,
} from "~/utils/connect-pagination";
import { transformDataTableQueryToRQLRequest } from "~/utils/transform-query";
import { useDebouncedValue } from "~hooks";
import { useTerminology } from "../../../../hooks/useTerminology";

const DEFAULT_SORT: DataTableSort = { name: 'createdAt', order: 'desc' };
const INITIAL_QUERY: DataTableQuery = {
  offset: 0,
  limit: DEFAULT_PAGE_SIZE,
};
const TRANSFORM_OPTIONS = {
  fieldNameMapping: {
    createdAt: "created_at",
    updatedAt: "updated_at",
    dueDate: "due_date",
    paidAt: "paid_at",
    effectiveAt: "effective_at",
    periodStart: "period_start",
    periodEnd: "period_end",
    invoiceLink: "invoice_link",
  },
};

// The backend stores invoice amounts in cents, but the amount filter input
// takes dollars — convert the filter value to cents before sending the query.
function convertAmountFiltersToCents(query: DataTableQuery): DataTableQuery {
  if (!query.filters?.length) return query;
  return {
    ...query,
    filters: query.filters.map(filter =>
      filter.name === "amount"
        ? {
            ...filter,
            value: Math.round(Number(filter.value) * 100),
            ...(filter.numberValue !== undefined && {
              numberValue: Math.round(filter.numberValue * 100),
            }),
          }
        : filter,
    ),
  };
}

const NoInvoices = () => {
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="No Invoice found"
      subHeading="We couldn't find any matches for that keyword or filter. Try alternative terms or check for typos."
      // TODO: update icon with raystack icon
      icon={<FileTextIcon />}
    />
  );
};

const ZeroState = () => {
  return (
    <div className={styles["zero-state-container"]}>
      <EmptyState
        variant="empty2"
        className={styles["zero-state"]}
        icon={<FileTextIcon />}
        heading="Invoices"
        subHeading="Invoices generated for this organization's billing activity will appear here."
      />
    </div>
  );
};

const ErrorState = () => {
  const t = useTerminology();
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="Error Loading Invoices"
      subHeading={`Something went wrong while loading ${t.organization({ case: "lower" })} invoices. Please try refreshing the page.`}
      icon={<ExclamationTriangleIcon />}
    />
  );
};

export function OrganizationInvoicesView() {
  const t = useTerminology();
  const { organization, search } = useContext(OrganizationContext);
  const organizationId = organization?.id || "";

  const {
    onChange: onSearchChange,
    setVisibility: setSearchVisibility,
    query: searchQuery,
  } = search;

  const [tableQuery, setTableQuery] = useState<DataTableQuery>(INITIAL_QUERY);

  const title = `Invoices | ${organization?.title} | ${t.organization({ plural: true, case: "capital" })}`;

  const computedQuery = useMemo(() => {
    const tempQuery = transformDataTableQueryToRQLRequest(
      convertAmountFiltersToCents(tableQuery),
      TRANSFORM_OPTIONS,
    );
    return {
      ...tempQuery,
      search: searchQuery || "",
    };
  }, [tableQuery, searchQuery]);

  const query = useDebouncedValue(computedQuery, 200);

  const {
    data: infiniteData,
    isLoading,
    isFetchingNextPage,
    fetchNextPage,
    hasNextPage,
    isError
  } = useInfiniteQuery(
    FrontierServiceQueries.searchOrganizationInvoices,
    { id: organizationId, query: query },
    {
      enabled: !!organizationId,
      pageParamKey: "query",
      getNextPageParam: lastPage =>
        getConnectNextPageParam(
          lastPage,
          { query: query },
          "organizationInvoices",
        ),
      staleTime: 0,
      refetchOnWindowFocus: false,
      retry: 1,
      retryDelay: 1000,
    },
  );
  const data =
    infiniteData?.pages?.flatMap(page => page.organizationInvoices) || [];
  const loading = (isLoading || isFetchingNextPage) && !isError;
  const groupCountMap = infiniteData
    ? getGroupCountMapFromFirstPage(infiniteData)
    : {};

  const onTableQueryChange = (newQuery: DataTableQuery) => {
    setTableQuery(newQuery);
  };

  const fetchMore = async () => {
    if (hasNextPage && !isFetchingNextPage && !isError) {
      await fetchNextPage();
    }
  };

  useEffect(() => {
    setSearchVisibility(true);
    return () => {
      onSearchChange("");
      setSearchVisibility(false);
    };
  }, [setSearchVisibility, onSearchChange]);

  const columns = getColumns({ groupCountMap });
  return (
    <Flex justify="center" className={styles["container"]}>
      <PageTitle title={title} />
      <DataTable
        columns={columns}
        data={data}
        isLoading={loading}
        defaultSort={DEFAULT_SORT}
        mode="server"
        onTableQueryChange={onTableQueryChange}
        onLoadMore={fetchMore}
        query={tableQuery}>
        <Flex direction="column" style={{ width: "100%" }}>
          <DataTable.Toolbar />
          <DataTable.Content
            emptyState={isError ? <ErrorState /> : <NoInvoices />}
            zeroState={<ZeroState />}
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
