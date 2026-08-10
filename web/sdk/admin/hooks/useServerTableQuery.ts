import { useCallback, useMemo, useRef, useState } from "react";
import type { DataTableQuery, DataTableSort } from "@raystack/apsara";
import type { RQLRequest } from "@raystack/proton/frontier";

import { DEFAULT_PAGE_SIZE } from "~/utils/connect-pagination";
import {
  transformDataTableQueryToRQLRequest,
  type TransformOptions,
} from "~/utils/transform-query";
import { useDebouncedValue } from "~hooks";

export interface ServerTableQueryOptions {
  /** Sort applied until the user picks another. Must match the table's `defaultSort`. */
  defaultSort?: DataTableSort;
  /** Field name mapping for the RQL request. Read through a ref, so an inline object is fine. */
  transformOptions?: TransformOptions;
  /** Search owned outside the table, e.g. the organization page's shared box. */
  search?: string;
  /** Adjust the query before it becomes a request, e.g. converting units. */
  mapQuery?: (query: DataTableQuery) => DataTableQuery;
  /** Debounce applied to the request, not to the table's own state. */
  debounceMs?: number;
}

export interface ServerTableQuery {
  /** Pass to DataTable's `query` prop. Updates immediately. */
  tableQuery: DataTableQuery;
  /** Pass to the RPC. Trails `tableQuery` by `debounceMs`. */
  rqlQuery: RQLRequest;
  /** Pass to DataTable's `onTableQueryChange` prop. */
  onTableQueryChange: (query: DataTableQuery) => void;
}

/**
 * Query state for a `mode="server"` DataTable.
 *
 * The initial query carries `defaultSort` on purpose. DataTable seeds its own
 * state from that prop and emits it on mount unconditionally; if the initial
 * query here disagreed, that emit would change the request and every table
 * would fetch its first page twice. Keep the `defaultSort` passed to DataTable
 * and the one passed here identical.
 */
export function useServerTableQuery({
  defaultSort,
  transformOptions,
  search,
  mapQuery,
  debounceMs = 200,
}: ServerTableQueryOptions = {}): ServerTableQuery {
  const [tableQuery, setTableQuery] = useState<DataTableQuery>(() => ({
    offset: 0,
    limit: DEFAULT_PAGE_SIZE,
    sort: defaultSort ? [defaultSort] : [],
  }));

  /*
   * Field mappings are fixed per view, so read them through a ref. Callers
   * passing an inline object would otherwise change the memo's identity every
   * render, restarting the debounce timer and never letting it settle.
   */
  const transformOptionsRef = useRef(transformOptions);
  transformOptionsRef.current = transformOptions;
  const mapQueryRef = useRef(mapQuery);
  mapQueryRef.current = mapQuery;

  const computedQuery = useMemo(() => {
    const mapped = mapQueryRef.current
      ? mapQueryRef.current(tableQuery)
      : tableQuery;
    const rql = transformDataTableQueryToRQLRequest(
      mapped,
      transformOptionsRef.current,
    );
    return search === undefined ? rql : { ...rql, search };
  }, [tableQuery, search]);

  const rqlQuery = useDebouncedValue(computedQuery, debounceMs);

  /* Any change to filters, sort or search starts again from the first page. */
  const onTableQueryChange = useCallback((query: DataTableQuery) => {
    setTableQuery({
      ...query,
      offset: 0,
      limit: query.limit || DEFAULT_PAGE_SIZE,
    });
  }, []);

  return { tableQuery, rqlQuery, onTableQueryChange };
}
