'use client';

import { useMemo } from 'react';
import { LockClosedIcon } from '@radix-ui/react-icons';
import {
  AlertDialog,
  Button,
  DataTable,
  Dialog,
  EmptyState,
  Flex,
  Skeleton
} from '@raystack/apsara-v1';
import type { DataTableQuery, DataTableSort } from '@raystack/apsara-v1';
import { useDebouncedState } from '@raystack/apsara-v1/hooks';
import { useInfiniteQuery } from '@connectrpc/connect-query';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import { FrontierServiceQueries } from '@raystack/proton/frontier';
import { useFrontier } from '../../contexts/FrontierContext';
import { useTerminology } from '../../hooks/useTerminology';
import { ViewContainer } from '../../components/view-container';
import { ViewHeader } from '../../components/view-header';
import { DEFAULT_DATE_FORMAT } from '../../utils/constants';
import {
  DEFAULT_PAGE_SIZE,
  getConnectNextPageParam
} from '~/utils/connect-pagination';
import { transformDataTableQueryToRQLRequest } from '~/utils/transform-query';
import { getColumns } from './components/pat-columns';
import { PATFormDialog } from './components/pat-form-dialog';
import { PATCreatedDialog } from './components/pat-created-dialog';
import { RevokePATDialog } from './components/revoke-pat-dialog';
import styles from './pat-view.module.css';

dayjs.extend(relativeTime);

const createPATDialogHandle = Dialog.createHandle();
const patCreatedDialogHandle = Dialog.createHandle<string>();
const revokePATDialogHandle = AlertDialog.createHandle<string>();

const DEFAULT_SORT: DataTableSort = { name: 'title', order: 'asc' };
const INITIAL_QUERY: DataTableQuery = {
  offset: 0,
  limit: DEFAULT_PAGE_SIZE
};
const TRANSFORM_OPTIONS = {
  fieldNameMapping: {
    createdAt: 'created_at',
    updatedAt: 'updated_at',
    expiresAt: 'expires_at',
    lastUsedAt: 'last_used_at'
  }
};

export interface PatsViewProps {
  onPATClick?: (patId: string) => void;
}

export function PatsView({ onPATClick }: PatsViewProps = {}) {
  const {
    activeOrganization: organization,
    isActiveOrganizationLoading,
    config
  } = useFrontier();
  const t = useTerminology();

  const orgId = organization?.id ?? '';

  const [tableQuery, setTableQuery] = useDebouncedState<DataTableQuery>(
    INITIAL_QUERY,
    200
  );

  const query = useMemo(
    () => transformDataTableQueryToRQLRequest(tableQuery, TRANSFORM_OPTIONS),
    [tableQuery]
  );

  const {
    data: infiniteData,
    isLoading: isPatsLoading,
    isFetchingNextPage,
    fetchNextPage,
    hasNextPage,
    refetch
  } = useInfiniteQuery(
    FrontierServiceQueries.searchCurrentUserPATs,
    { orgId, query },
    {
      enabled: Boolean(orgId),
      pageParamKey: 'query',
      getNextPageParam: lastPage =>
        getConnectNextPageParam(lastPage, { query }, 'pats'),
      staleTime: 0,
      refetchOnWindowFocus: false
    }
  );

  const pats = useMemo(
    () => infiniteData?.pages?.flatMap(page => page?.pats ?? []) ?? [],
    [infiniteData]
  );

  const hasActiveQuery = Boolean(
    tableQuery.search || tableQuery.filters?.length
  );
  const isInitialLoading = !organization?.id || isActiveOrganizationLoading;
  const isTableLoading = isPatsLoading || isFetchingNextPage;
  const hasNoPats =
    !isInitialLoading &&
    !isPatsLoading &&
    !hasActiveQuery &&
    pats.length === 0;

  const dateFormat = config?.dateFormat || DEFAULT_DATE_FORMAT;

  const columns = useMemo(
    () =>
      getColumns({
        dateFormat,
        onRevoke: (patId: string) =>
          revokePATDialogHandle.openWithPayload(patId)
      }),
    [dateFormat]
  );

  const onTableQueryChange = (newQuery: DataTableQuery) => {
    setTableQuery({
      ...newQuery,
      offset: 0,
      limit: newQuery.limit || DEFAULT_PAGE_SIZE
    });
  };

  const handleLoadMore = async () => {
    if (hasNextPage && !isFetchingNextPage) {
      await fetchNextPage();
    }
  };

  const handlePATCreated = (token: string) => {
    patCreatedDialogHandle.openWithPayload(token);
  };

  const handleSuccessDialogClose = () => {
    refetch();
  };

  return (
    <ViewContainer>
      <ViewHeader
        title="Personal access token"
        description={`Create a personal access token to enable secure access to ${t.appName()} resources via PAT token`}
      />

      {isInitialLoading ? (
        <Flex direction="column" gap={7}>
          <Skeleton height="34px" width="360px" />
          <Skeleton height="48px" width="100%" />
          <Skeleton height="80px" width="100%" />
          <Skeleton height="80px" width="100%" />
        </Flex>
      ) : hasNoPats ? (
        <EmptyState
          icon={<LockClosedIcon />}
          heading="No Personal Access Token Found"
          subHeading={`Create a new to use the Keys of ${t.appName()} platform`}
          primaryAction={
            <Button
              variant="outline"
              color="neutral"
              onClick={() => createPATDialogHandle.open(null)}
              data-test-id="frontier-sdk-add-pat-btn"
            >
              Add new personal access token
            </Button>
          }
        />
      ) : (
        <DataTable
          data={pats}
          columns={columns}
          isLoading={isTableLoading}
          defaultSort={DEFAULT_SORT}
          mode="server"
          query={tableQuery}
          onTableQueryChange={onTableQueryChange}
          onLoadMore={handleLoadMore}
          onRowClick={row => onPATClick?.(row.id)}
        >
          <Flex direction="column" gap={7}>
            <Flex justify="between" gap={3}>
              <DataTable.Search
                placeholder="Search by name."
                size="large"
                width={360}
                disabled={false}
              />
              <Button
                variant="solid"
                color="accent"
                onClick={() => createPATDialogHandle.open(null)}
                data-test-id="frontier-sdk-create-pat-btn"
              >
                Create new PAT
              </Button>
            </Flex>
            <DataTable.Content
              emptyState={
                <EmptyState
                  icon={<LockClosedIcon />}
                  heading="No tokens matching your search"
                />
              }
              classNames={{
                root: styles.tableRoot
              }}
            />
          </Flex>
        </DataTable>
      )}

      <PATFormDialog
        handle={createPATDialogHandle}
        onCreated={handlePATCreated}
      />
      <PATCreatedDialog
        handle={patCreatedDialogHandle}
        onClose={handleSuccessDialogClose}
      />
      <RevokePATDialog
        handle={revokePATDialogHandle}
        onRevoked={() => refetch()}
      />
    </ViewContainer>
  );
}
