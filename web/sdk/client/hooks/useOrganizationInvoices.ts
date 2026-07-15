import { useMemo } from 'react';
import { useInfiniteQuery } from '@connectrpc/connect-query';
import { FrontierServiceQueries } from '@raystack/proton/frontier';
import type {
  RQLRequest,
  SearchOrganizationInvoicesResponse_OrganizationInvoice
} from '@raystack/proton/frontier';
import { useFrontier } from '../contexts/FrontierContext';
import {
  getConnectNextPageParam,
  getGroupCountMapFromFirstPage
} from '../utils/connect-pagination';

export interface UseOrganizationInvoicesOptions {
  query: RQLRequest;
  enabled?: boolean;
}

export interface UseOrganizationInvoicesReturn {
  invoices: SearchOrganizationInvoicesResponse_OrganizationInvoice[];
  groupCountMap: Record<string, Record<string, number>>;
  isLoading: boolean;
  isFetchingNextPage: boolean;
  fetchNextPage: () => Promise<unknown>;
  hasNextPage: boolean;
  isError: boolean;
  error: Error | null;
}

/**
 * Common data source for organization invoices, backed by the
 * `SearchOrganizationInvoices` RPC as a paginated infinite query.
 */
export function useOrganizationInvoices({
  query,
  enabled = true
}: UseOrganizationInvoicesOptions): UseOrganizationInvoicesReturn {
  const { activeOrganization } = useFrontier();
  const organizationId = activeOrganization?.id || '';

  const {
    data: infiniteData,
    isLoading,
    isFetchingNextPage,
    fetchNextPage,
    hasNextPage,
    isError,
    error
  } = useInfiniteQuery(
    FrontierServiceQueries.searchOrganizationInvoices,
    { id: organizationId, query },
    {
      enabled: !!organizationId && enabled,
      pageParamKey: 'query',
      getNextPageParam: lastPage =>
        getConnectNextPageParam(lastPage, { query }, 'organizationInvoices'),
      staleTime: 0,
      refetchOnWindowFocus: false,
      retry: 1,
      retryDelay: 1000
    }
  );

  const invoices = useMemo(
    () => infiniteData?.pages?.flatMap(page => page.organizationInvoices) ?? [],
    [infiniteData]
  );

  const groupCountMap = useMemo(
    () => (infiniteData ? getGroupCountMapFromFirstPage(infiniteData) : {}),
    [infiniteData]
  );

  return {
    invoices,
    groupCountMap,
    isLoading,
    isFetchingNextPage,
    fetchNextPage,
    hasNextPage,
    isError,
    error
  };
}
