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
import { useQuery } from '@connectrpc/connect-query';
import { create } from '@bufbuild/protobuf';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import {
  FrontierServiceQueries,
  SearchCurrentUserPATsRequestSchema,
  RQLRequestSchema
} from '@raystack/proton/frontier';
import { useFrontier } from '../../contexts/FrontierContext';
import { useTerminology } from '../../hooks/useTerminology';
import { ViewContainer } from '../../components/view-container';
import { ViewHeader } from '../../components/view-header';
import { DEFAULT_DATE_FORMAT } from '../../utils/constants';
import { getColumns } from './components/pat-columns';
import { PATFormDialog } from './components/pat-form-dialog';
import { PATCreatedDialog } from './components/pat-created-dialog';
import { RevokePATDialog } from './components/revoke-pat-dialog';
import styles from './pat-view.module.css';

dayjs.extend(relativeTime);

const createPATDialogHandle = Dialog.createHandle();
const patCreatedDialogHandle = Dialog.createHandle<string>();
const revokePATDialogHandle = AlertDialog.createHandle<string>();

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

  const {
    data: patsData,
    isLoading: isPatsLoading,
    refetch
  } = useQuery(
    FrontierServiceQueries.searchCurrentUserPATs,
    create(SearchCurrentUserPATsRequestSchema, {
      orgId,
      query: create(RQLRequestSchema, {})
    }),
    {
      enabled: Boolean(orgId)
    }
  );

  const pats = useMemo(() => patsData?.pats ?? [], [patsData]);

  const isInitialLoading = !organization?.id || isActiveOrganizationLoading;
  const hasNoPats = !isInitialLoading && !isPatsLoading && pats.length === 0;

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
          isLoading={isPatsLoading}
          defaultSort={{ name: 'title', order: 'asc' }}
          mode="client"
          onRowClick={row => onPATClick?.(row.id)}
        >
          <Flex direction="column" gap={7}>
            <Flex justify="between" gap={3}>
              <DataTable.Search
                placeholder="Search by name."
                size="large"
                width={360}
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
