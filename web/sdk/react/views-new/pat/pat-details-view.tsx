'use client';

import { ReactNode, useCallback, useEffect, useMemo } from 'react';
import { DotsHorizontalIcon } from '@radix-ui/react-icons';
import {
  AlertDialog,
  Breadcrumb,
  Button,
  Dialog,
  Flex,
  IconButton,
  Menu,
  Skeleton,
  Text,
  toastManager
} from '@raystack/apsara-v1';
import { useQuery } from '@connectrpc/connect-query';
import { create } from '@bufbuild/protobuf';
import {
  FrontierServiceQueries,
  GetCurrentUserPATRequestSchema,
  ListRolesForPATRequestSchema,
  ListOrganizationProjectsRequestSchema
} from '@raystack/proton/frontier';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import { useFrontier } from '../../contexts/FrontierContext';
import { ViewContainer } from '../../components/view-container';
import { ViewHeader } from '../../components/view-header';
import { DEFAULT_DATE_FORMAT } from '../../utils/constants';
import { PERMISSIONS } from '../../../utils';
import { isNullTimestamp, timestampToDayjs } from '~/utils/timestamp';
import {
  PATCreatedDialog,
  type PATCreatedPayload
} from './components/pat-created-dialog';
import { PATFormDialog } from './components/pat-form-dialog';
import { PATProjectChips } from './components/pat-project-chips';
import {
  RegeneratePATDialog,
  type RegeneratePayload
} from './components/regenerate-pat-dialog';
import { RevokePATDialog } from './components/revoke-pat-dialog';
import { getExpiryOptionValue, getExpiryReferenceDayjs } from './utils';
import styles from './pat-details-view.module.css';

dayjs.extend(relativeTime);

const updatePATDialogHandle = Dialog.createHandle();
const regenerateDialogHandle = Dialog.createHandle<RegeneratePayload>();
const patCreatedDialogHandle = Dialog.createHandle<PATCreatedPayload>();
const revokePATDialogHandle = AlertDialog.createHandle<string>();

interface DetailRowProps {
  label: string;
  children: ReactNode;
}

function DetailRow({ label, children }: DetailRowProps) {
  return (
    <div className={styles.detailRow}>
      <Text size="small">{label}</Text>
      {typeof children === 'string' ? (
        <Text size="small" weight="medium">
          {children}
        </Text>
      ) : (
        children
      )}
    </div>
  );
}

export interface PATDetailsViewProps {
  patId: string;
  onNavigateToPats?: () => void;
  onDeleteSuccess?: () => void;
}

export function PATDetailsView({
  patId,
  onNavigateToPats,
  onDeleteSuccess
}: PATDetailsViewProps) {
  const { activeOrganization: organization, config } = useFrontier();
  const orgId = organization?.id || '';
  const dateFormat = config?.dateFormat || DEFAULT_DATE_FORMAT;

  const {
    data: pat,
    isLoading: isPatLoading,
    error: patError,
    refetch: refetchPat
  } = useQuery(
    FrontierServiceQueries.getCurrentUserPAT,
    create(GetCurrentUserPATRequestSchema, { id: patId }),
    {
      enabled: Boolean(patId),
      select: d => d?.pat
    }
  );

  useEffect(() => {
    if (patError) {
      toastManager.add({
        title: 'Something went wrong',
        description: patError.message,
        type: 'error'
      });
    }
  }, [patError]);

  const { data: orgRolesData, isLoading: isOrgRolesLoading } = useQuery(
    FrontierServiceQueries.listRolesForPAT,
    create(ListRolesForPATRequestSchema, { scopes: [PERMISSIONS.OrganizationNamespace] }),
    { enabled: Boolean(orgId) }
  );
  const orgRoles = useMemo(() => orgRolesData?.roles ?? [], [orgRolesData]);

  const { data: projectRolesData, isLoading: isProjectRolesLoading } =
    useQuery(
      FrontierServiceQueries.listRolesForPAT,
      create(ListRolesForPATRequestSchema, { scopes: [PERMISSIONS.ProjectNamespace] }),
      { enabled: Boolean(orgId) }
    );
  const projectRoles = useMemo(
    () => projectRolesData?.roles ?? [],
    [projectRolesData]
  );

  const { data: projectsData, isLoading: isProjectsLoading } = useQuery(
    FrontierServiceQueries.listOrganizationProjects,
    create(ListOrganizationProjectsRequestSchema, {
      id: orgId,
      state: '',
      withMemberCount: false
    }),
    { enabled: Boolean(orgId) }
  );
  const projects = useMemo(
    () => projectsData?.projects ?? [],
    [projectsData]
  );

  const isLoading =
    !organization?.id ||
    isPatLoading ||
    isOrgRolesLoading ||
    isProjectRolesLoading ||
    isProjectsLoading;

  const orgScope = useMemo(
    () => pat?.scopes?.find(s => s.resourceType === PERMISSIONS.OrganizationNamespace),
    [pat]
  );

  const projectScope = useMemo(
    () => pat?.scopes?.find(s => s.resourceType === PERMISSIONS.ProjectNamespace),
    [pat]
  );

  const orgRoleName = useMemo(() => {
    if (!orgScope) return '';
    const role = orgRoles.find(r => r.id === orgScope.roleId);
    return role?.title || role?.name || '';
  }, [orgScope, orgRoles]);

  const projectRoleName = useMemo(() => {
    if (!projectScope) return '';
    const role = projectRoles.find(r => r.id === projectScope.roleId);
    return role?.title || role?.name || '';
  }, [projectScope, projectRoles]);

  const scopeProjects = useMemo(() => {
    if (!projectScope?.resourceIds?.length) return [];
    return projects
      .filter(p => projectScope.resourceIds.includes(p.id || ''))
      .map(p => ({ id: p.id || '', title: p.title || p.id || '' }));
  }, [projectScope, projects]);

  const isAllProjects =
    !projectScope?.resourceIds || projectScope.resourceIds.length === 0;

  const createdOn = useMemo(() => {
    const d = timestampToDayjs(pat?.createdAt);
    return d ? d.format(dateFormat) : '';
  }, [pat, dateFormat]);

  const lastUsed = useMemo(() => {
    if (!pat?.usedAt || isNullTimestamp(pat.usedAt)) return '';
    const d = timestampToDayjs(pat.usedAt);
    return d ? d.fromNow() : '';
  }, [pat]);

  const regeneratedOn = useMemo(() => {
    if (!pat?.regeneratedAt || isNullTimestamp(pat.regeneratedAt)) return '';
    const d = timestampToDayjs(pat.regeneratedAt);
    return d ? d.format(dateFormat) : '';
  }, [pat, dateFormat]);

  const { expiryInfo, currentExpiryValue } = useMemo(() => {
    const reference = getExpiryReferenceDayjs(pat);
    const expires = timestampToDayjs(pat?.expiresAt);
    if (!reference || !expires)
      return { expiryInfo: '', currentExpiryValue: '' };
    const days = expires.diff(reference, 'day');
    return {
      expiryInfo: `${expires.format(dateFormat)} (${days} Days)`,
      currentExpiryValue: getExpiryOptionValue(reference, expires)
    };
  }, [pat, dateFormat]);

  const handleRegenerated = useCallback(
    (token: string) => {
      patCreatedDialogHandle.openWithPayload({ token, isRegenerated: true });
    },
    []
  );

  const handleTokenDialogClose = () => {
    refetchPat();
  };

  const patTitle = pat?.title || '';

  return (
    <ViewContainer>
      <ViewHeader
        title={isPatLoading ? '' : patTitle}
        breadcrumb={
          <Breadcrumb size="small">
            <Breadcrumb.Item
              href="#"
              onClick={e => {
                e.preventDefault();
                onNavigateToPats?.();
              }}
            >
              Personal access token
            </Breadcrumb.Item>
            <Breadcrumb.Separator />
            <Breadcrumb.Item current>
              {isPatLoading ? (
                <Skeleton height="16px" width="100px" />
              ) : (
                patTitle
              )}
            </Breadcrumb.Item>
          </Breadcrumb>
        }
      >
        <Menu modal={false}>
          <Menu.Trigger
            render={
              <IconButton
                size={2}
                aria-label="PAT actions"
                data-test-id="frontier-sdk-pat-details-actions-btn"
              />
            }
          >
            <DotsHorizontalIcon />
          </Menu.Trigger>
          <Menu.Content align="start" className={styles.menuContent}>
            <Menu.Item
              onClick={() => updatePATDialogHandle.open(null)}
              data-test-id="frontier-sdk-pat-update-menu-btn"
            >
              Update
            </Menu.Item>
            <Menu.Item
              onClick={() =>
                regenerateDialogHandle.openWithPayload({
                  patId,
                  currentExpiryValue
                })
              }
              data-test-id="frontier-sdk-pat-regenerate-menu-btn"
            >
              Regenerate
            </Menu.Item>
            <Menu.Item
              onClick={() => revokePATDialogHandle.openWithPayload(patId)}
              style={{ color: 'var(--rs-color-foreground-danger-primary)' }}
              data-test-id="frontier-sdk-pat-revoke-menu-btn"
            >
              Revoke
            </Menu.Item>
          </Menu.Content>
        </Menu>
      </ViewHeader>

      {isLoading ? (
        <Flex direction="column" gap={7}>
          <Skeleton height="200px" width="100%" />
          <Skeleton height="120px" width="100%" />
        </Flex>
      ) : (
        <>
          <div className={styles.section}>
            <Flex justify="between" align="center">
              <Text size="regular" weight="medium">
                General
              </Text>
              <Button
                variant="outline"
                color="neutral"
                size="small"
                onClick={() => updatePATDialogHandle.open(null)}
                data-test-id="frontier-sdk-pat-update-btn"
              >
                Update
              </Button>
            </Flex>
            <Flex direction="column" gap={5}>
              {createdOn && <DetailRow label="Created on:">{createdOn}</DetailRow>}
              {lastUsed && <DetailRow label="Last used:">{lastUsed}</DetailRow>}
              {orgRoleName && (
                <DetailRow label="Organization role:">{orgRoleName}</DetailRow>
              )}
              {projectRoleName && (
                <DetailRow label="Project role:">{projectRoleName}</DetailRow>
              )}
              <DetailRow label="Projects:">
                {isAllProjects || scopeProjects.length === 0 ? (
                  <Text size="small" weight="medium">
                    All projects
                  </Text>
                ) : (
                  <PATProjectChips projects={scopeProjects} />
                )}
              </DetailRow>
            </Flex>
          </div>

          <div className={styles.section}>
            <Flex justify="between" align="center">
              <Text size="regular" weight="medium">
                Expiry Details
              </Text>
              <Button
                variant="outline"
                color="neutral"
                size="small"
                onClick={() =>
                  regenerateDialogHandle.openWithPayload({
                    patId,
                    currentExpiryValue
                  })
                }
                data-test-id="frontier-sdk-pat-regenerate-btn"
              >
                Regenerate
              </Button>
            </Flex>
            <Flex direction="column" gap={5}>
              {expiryInfo && (
                <DetailRow label="Expiry on:">{expiryInfo}</DetailRow>
              )}
              {regeneratedOn && (
                <DetailRow label="Regenerated on:">{regeneratedOn}</DetailRow>
              )}
            </Flex>
          </div>
        </>
      )}

      <RevokePATDialog
        handle={revokePATDialogHandle}
        onRevoked={onDeleteSuccess}
      />
      <PATFormDialog
        handle={updatePATDialogHandle}
        initialData={pat}
        onUpdated={() => refetchPat()}
      />
      <RegeneratePATDialog
        handle={regenerateDialogHandle}
        onRegenerated={handleRegenerated}
      />
      <PATCreatedDialog
        handle={patCreatedDialogHandle}
        onClose={handleTokenDialogClose}
      />
    </ViewContainer>
  );
}
