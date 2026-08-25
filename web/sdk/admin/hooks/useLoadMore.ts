import { useCallback, useRef } from "react";

export interface UseLoadMoreOptions {
  hasNextPage?: boolean;
  isFetchingNextPage: boolean;
  fetchNextPage: () => Promise<unknown>;
  /** Skip while the query is errored, so scrolling cannot retry a failed page. */
  isError?: boolean;
  /** Names the rows in the console message, e.g. "audit logs". */
  label: string;
}

/*
  Guarded "load more" for a server table's infinite query.
  - VirtualizedContent calls this straight from onScroll and react-query
    notifies observers on a macrotask, so hasNextPage/isFetchingNextPage are
    still last render's values through a scroll burst
  - fetchNextPage defaults to cancelRefetch: true, so an unguarded repeat aborts
    the in-flight page and re-issues it; only the ref flips in time to stop that
  - the render-derived flags stay as a cheap first filter
*/
export const useLoadMore = ({
  hasNextPage,
  isFetchingNextPage,
  fetchNextPage,
  isError,
  label,
}: UseLoadMoreOptions) => {
  const isLoadingMoreRef = useRef(false);

  return useCallback(async () => {
    if (
      !hasNextPage ||
      isFetchingNextPage ||
      isError ||
      isLoadingMoreRef.current
    ) {
      return;
    }
    isLoadingMoreRef.current = true;
    try {
      await fetchNextPage();
    } catch (error) {
      console.error(`Error loading more ${label}:`, error);
    } finally {
      isLoadingMoreRef.current = false;
    }
  }, [hasNextPage, isFetchingNextPage, isError, fetchNextPage, label]);
};
