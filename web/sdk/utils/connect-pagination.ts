import type {
  RQLQueryPaginationResponse,
  RQLQueryGroupResponse,
  RQLQueryGroupData,
  RQLRequest,
} from "@raystack/proton/frontier";

/** Default page size used for paginated RQL queries. */
export const DEFAULT_PAGE_SIZE = 50;

/** Shape of a Connect RPC response that includes pagination and group metadata. */
export type ConnectRPCPaginatedResponse = {
  pagination?: RQLQueryPaginationResponse;
  group?: RQLQueryGroupResponse;
  [key: string]: unknown;
}

/**
 * Returns the next page params for `useInfiniteQuery` based on the last page's pagination metadata.
 * Returns `undefined` when there are no more pages.
 *
 * @param lastPage - The most recent page returned by the query.
 * @param queryParams - The current query params (contains the RQLRequest).
 * @param itemsKey - The key in the response that holds the list of items (defaults to `"organizations"`).
 */
export function getConnectNextPageParam<T extends ConnectRPCPaginatedResponse>(
  lastPage: T,
  queryParams: { query: RQLRequest },
  itemsKey: string = "organizations",
) {
  // Use pagination info from response to determine next page
  const pagination = lastPage.pagination;
  const currentOffset = pagination?.offset || 0;
  const limit = pagination?.limit || DEFAULT_PAGE_SIZE;

  // Check if there are more pages based on returned results
  const items = Array.isArray(lastPage[itemsKey]) ? lastPage[itemsKey] : [];
  const hasMorePages = items && items.length !== 0 && items.length === limit;
  if (!hasMorePages) {
    return undefined; // No more pages
  }

  const nextOffset = currentOffset + limit;
  const nextParams = {
    ...queryParams.query,
    offset: nextOffset,
  };

  return nextParams;
}

/**
 * Extracts group-by count map from the first page of an infinite query.
 * Returns a nested map: `{ [groupName]: { [value]: count } }`.
 */
export function getGroupCountMapFromFirstPage<
  T extends ConnectRPCPaginatedResponse,
>(infiniteData?: { pages: T[] }): Record<string, Record<string, number>> {
  if (!infiniteData?.pages?.[0]) {
    return {};
  }

  const firstPage = infiniteData.pages[0];
  const group = firstPage.group;

  if (!group?.data || !group.name) {
    return {};
  }

  const groupCount = group.data.reduce(
    (acc: Record<string, number>, groupItem: RQLQueryGroupData) => {
      acc[groupItem.name || ""] = groupItem.count || 0;
      return acc;
    },
    {} as Record<string, number>,
  );

  const groupKey = group.name;
  return { [groupKey]: groupCount };
}
