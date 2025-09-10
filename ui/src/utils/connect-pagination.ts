import type {
  RQLQueryPaginationResponse,
  RQLQueryGroupResponse,
  RQLQueryGroupData,
  RQLRequest,
} from "@raystack/proton/frontier";

export const DEFAULT_PAGE_SIZE = 50;

export interface ConnectRPCPaginatedResponse<T = unknown> {
  pagination?: RQLQueryPaginationResponse;
  group?: RQLQueryGroupResponse;
  [key: string]: T[] | RQLQueryPaginationResponse | RQLQueryGroupResponse | undefined;
}

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
  const items = lastPage[itemsKey] as unknown[];
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
