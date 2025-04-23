import { useCallback, useState } from "react";
import { DataTableQuery } from "@raystack/apsara/v1";
import {
  V1Beta1RQLQueryPaginationResponse,
  V1Beta1RQLQueryGroupResponse,
} from "~/api/frontier";
import { useDebounceCallback } from "usehooks-ts";

interface GroupCountMap {
  [key: string]: Record<string, number>;
}

interface UseRQLResponse<T> {
  data: T[];
  loading: boolean;
  query: DataTableQuery;
  setQuery: (query: DataTableQuery) => void;
  fetchData: (apiQuery?: DataTableQuery) => Promise<void>;
  group?: V1Beta1RQLQueryGroupResponse;
  hasMore: boolean;
  onTableQueryChange: (query: DataTableQuery) => void;
  groupCountMap: GroupCountMap;
  nextOffset: number;
  fetchMore: () => void;
}

type UseRQLProps<T> = {
  initialQuery?: DataTableQuery;
  resourceId: string;
  dataKey: string;
  fn: (apiQuery?: DataTableQuery) => unknown;
  searchParam?: string;
  limit?: number;
  onError?: (error: Error | unknown) => void;
};

export const useRQL = <T extends unknown>({
  initialQuery = { offset: 0 },
  resourceId,
  dataKey,
  fn,
  searchParam = "",
  limit = 50,
  onError,
}: UseRQLProps<T>): UseRQLResponse<T> => {
  const [data, setData] = useState<T[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [query, setQuery] = useState<DataTableQuery>(initialQuery);
  const [hasMore, setHasMore] = useState(false);
  const [nextOffset, setNextOffset] = useState(0);
  const [groupCountMap, setGroupCountMap] = useState<GroupCountMap>({});

  const fetchData = useCallback(
    async (apiQuery: DataTableQuery = {}) => {
      if (!resourceId) return;

      try {
        setLoading(true);
        const queryParams = {
          ...apiQuery,
          limit,
          search: searchParam,
        };

        const response = (await fn(queryParams)) as Record<string, any>;
        const responseItems = response[dataKey] || [];
        // Ensure we have a proper array of items
        const items = Array.isArray(responseItems)
          ? (responseItems as T[])
          : [];

        if (apiQuery.offset === 0) {
          setData(items);
        } else {
          setData((prev) => [...prev, ...items]);
        }

        const pagination = response[
          "pagination"
        ] as V1Beta1RQLQueryPaginationResponse;
        if (pagination) {
          setNextOffset(pagination?.offset || 0);
        }
        setHasMore(items.length !== 0 && items.length === limit);

        const group = response.group as V1Beta1RQLQueryGroupResponse;
        // Handle group counts
        if (group?.data && group.name) {
          const groupCount = group.data.reduce(
            (
              acc: Record<string, number>,
              group: { name?: string; count?: number },
            ) => {
              acc[group.name || ""] = group.count || 0;
              return acc;
            },
            {} as Record<string, number>,
          );
          const groupKey = group.name;
          setGroupCountMap((prev) => ({ ...prev, [groupKey]: groupCount }));
        }
      } catch (error: unknown) {
        if (onError) {
          onError(error);
        } else {
          console.error("An unknown error occurred", error);
        }
      } finally {
        setLoading(false);
      }
    },
    [resourceId, fn, dataKey, limit, searchParam, onError],
  );

  const fetchMore = useCallback(() => {
    if (loading || !hasMore || !resourceId) return;
    fetchData({ ...query, offset: nextOffset + limit });
  }, [fetchData, hasMore, loading, limit, nextOffset, query, resourceId]);

  const onTableQueryChange = useDebounceCallback((newQuery: DataTableQuery) => {
    setData([]);
    fetchData({ ...newQuery, offset: 0 });
    setQuery(newQuery);
  }, 500);

  return {
    data,
    loading,
    query,
    setQuery,
    fetchData,
    hasMore,
    onTableQueryChange,
    groupCountMap,
    nextOffset,
    fetchMore,
  };
};
