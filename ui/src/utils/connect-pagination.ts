import type { V1Beta1RQLQueryPaginationResponse, V1Beta1RQLQueryGroupResponse, V1Beta1RQLQueryGroupData } from "~/api/frontier";

export const DEFAULT_PAGE_SIZE = 50;

export interface ConnectRPCPaginatedResponse {
  pagination?: V1Beta1RQLQueryPaginationResponse;
  group?: V1Beta1RQLQueryGroupResponse;
  [key: string]: any;
}

export function getConnectNextPageParam<T extends ConnectRPCPaginatedResponse>(
  lastPage: T,
  queryParams: { query: any },
  itemsKey: string = "organizations"
) {
  // Use pagination info from response to determine next page
  const pagination = lastPage.pagination;
  const currentOffset = pagination?.offset || 0;
  const limit = pagination?.limit || DEFAULT_PAGE_SIZE;

  // Check if there are more pages based on returned results
  const items = lastPage[itemsKey] as any[];
  const hasMorePages =
    items &&
    items.length !== 0 &&
    items.length === limit;
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

export function getGroupCountMapFromFirstPage<T extends ConnectRPCPaginatedResponse>(
  infiniteData?: { pages: T[] }
): Record<string, Record<string, number>> {
  if (!infiniteData?.pages?.[0]) {
    return {};
  }

  const firstPage = infiniteData.pages[0];
  const group = firstPage.group;
  
  if (!group?.data || !group.name) {
    return {};
  }

  const groupCount = group.data.reduce(
    (
      acc: Record<string, number>,
      groupItem: V1Beta1RQLQueryGroupData,
    ) => {
      acc[groupItem.name || ""] = groupItem.count || 0;
      return acc;
    },
    {} as Record<string, number>,
  );
  
  const groupKey = group.name;
  return { [groupKey]: groupCount };
}