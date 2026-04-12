'use client';

import { useMemo } from 'react';
import { LockClosedIcon, MagnifyingGlassIcon } from '@radix-ui/react-icons';
import { useDebouncedState } from '@raystack/apsara-v1/hooks';
import {
  AlertDialog,
  Button,
  Dialog,
  EmptyState,
  Flex,
  InputField,
  Skeleton,
  Text
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
import type { PAT } from '@raystack/proton/frontier';
import { useFrontier } from '../../contexts/FrontierContext';
import { useTerminology } from '../../hooks/useTerminology';
import { ViewContainer } from '../../components/view-container';
import { ViewHeader } from '../../components/view-header';
import { DEFAULT_DATE_FORMAT } from '../../utils/constants';
import { timestampToDayjs, isNullTimestamp } from '../../../utils/timestamp';
import { TokenCell } from './components/token-cell';
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

  const [debouncedSearch, setDebouncedSearch] = useDebouncedState('', 300);

  const orgId = organization?.id ?? '';

  const rqlQuery = useMemo(
    () => create(RQLRequestSchema, { search: debouncedSearch || '' }),
    [debouncedSearch]
  );

  const {
    data: patsData,
    isLoading: isPatsLoading,
    refetch
  } = useQuery(
    FrontierServiceQueries.searchCurrentUserPATs,
    create(SearchCurrentUserPATsRequestSchema, {
      orgId,
      query: rqlQuery
    }),
    {
      enabled: Boolean(orgId)
    }
  );

  const pats = useMemo(() => patsData?.pats ?? [], [patsData]);

  const isLoading = !organization?.id || isActiveOrganizationLoading || isPatsLoading;
  const hasNoPats = !isLoading && pats.length === 0 && !debouncedSearch.trim();

  const dateFormat = config?.dateFormat || DEFAULT_DATE_FORMAT;

  const formatExpiry = (pat: PAT): string => {
    const d = timestampToDayjs(pat.expiresAt);
    return d ? `Exp: ${d.format(dateFormat)}` : '';
  };

  const formatLastUsed = (pat: PAT): string => {
    if (!pat.lastUsedAt || isNullTimestamp(pat.lastUsedAt)) return '';
    const d = timestampToDayjs(pat.lastUsedAt);
    return d ? `Last used ${d.fromNow()}` : '';
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

      {isLoading ? (
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
        <>
          <InputField
            placeholder="Search by name."
            size="large"
            leadingIcon={<MagnifyingGlassIcon />}
            onChange={e => setDebouncedSearch(e.target.value)}
            className={styles.searchInput}
            data-test-id="frontier-sdk-pat-search-input"
          />

          <div className={styles.tokenList}>
            <Flex
              className={styles.tokenListHeader}
              justify="between"
              align="center"
            >
              <Text size="regular" weight="medium">
                {pats.length}{' '}
                {pats.length === 1 ? 'Token' : 'Tokens'}
              </Text>
              <Button
                variant="text"
                color="neutral"
                size="small"
                onClick={() => createPATDialogHandle.open(null)}
                data-test-id="frontier-sdk-create-pat-btn"
              >
                Create new PAT
              </Button>
            </Flex>

            {pats.length === 0 ? (
              <Flex
                className={styles.tokenCell}
                align="center"
                justify="center"
              >
                <Text size="regular" variant="secondary">
                  No tokens matching your search
                </Text>
              </Flex>
            ) : (
              pats.map(pat => (
                <TokenCell
                  key={pat.id}
                  title={pat.title}
                  expiry={formatExpiry(pat)}
                  lastUsed={formatLastUsed(pat)}
                  onClick={() => onPATClick?.(pat.id)}
                  onRevoke={() =>
                    revokePATDialogHandle.openWithPayload(pat.id)
                  }
                />
              ))
            )}
          </div>
        </>
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
