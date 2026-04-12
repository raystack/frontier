'use client';

import { useCallback, useEffect, useMemo } from 'react';
import { DotsHorizontalIcon } from '@radix-ui/react-icons';
import {
  AlertDialog,
  Breadcrumb,
  Button,
  Chip,
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
import { useFrontier } from '../../contexts/FrontierContext';
import { ViewContainer } from '../../components/view-container';
import { ViewHeader } from '../../components/view-header';
import { DEFAULT_DATE_FORMAT } from '../../utils/constants';
import { PERMISSIONS } from '../../../utils';
import { timestampToDayjs } from '../../../utils/timestamp';
import { PATCreatedDialog } from './components/pat-created-dialog';
import { PATFormDialog } from './components/pat-form-dialog';
import {
  RegeneratePATDialog,
  type RegeneratePayload
} from './components/regenerate-pat-dialog';
import { RevokePATDialog } from './components/revoke-pat-dialog';
import styles from './pat-details-view.module.css';

const updatePATDialogHandle = Dialog.createHandle();
const regenerateDialogHandle = Dialog.createHandle<RegeneratePayload>();
const patCreatedDialogHandle = Dialog.createHandle<string>();
const revokePATDialogHandle = AlertDialog.createHandle<string>();

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
    return projects.filter(p =>
      projectScope.resourceIds.includes(p.id || '')
    );
  }, [projectScope, projects]);

  const createdOn = useMemo(() => {
    const d = timestampToDayjs(pat?.createdAt);
    return d ? d.format(dateFormat) : '';
  }, [pat, dateFormat]);

  const { expiryInfo, expiryDays } = useMemo(() => {
    const created = timestampToDayjs(pat?.createdAt);
    const expires = timestampToDayjs(pat?.expiresAt);
    if (!created || !expires) return { expiryInfo: '', expiryDays: '' };
    const days = expires.diff(created, 'day');
    return {
      expiryInfo: `${days} Days (Exp: ${expires.format(dateFormat)})`,
      expiryDays: String(days)
    };
  }, [pat, dateFormat]);

  const handleRegenerated = useCallback(
    (token: string) => {
      patCreatedDialogHandle.openWithPayload(token);
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
          <Skeleton height="120px" width="100%" />
          <Skeleton height="80px" width="100%" />
        </Flex>
      ) : (
        <>
          <div className={styles.section}>
            <Flex
              className={styles.sectionHeader}
              justify="between"
              align="center"
            >
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
            <Flex className={styles.sectionBody} direction="column" gap={5}>
              {orgRoleName && (
                <div className={styles.detailRow}>
                  <Text size="small" className={styles.detailLabel}>
                    Organization role :
                  </Text>
                  <Text size="small" weight="medium">
                    {orgRoleName}
                  </Text>
                </div>
              )}
              {projectRoleName && (
                <div className={styles.detailRow}>
                  <Text size="small" className={styles.detailLabel}>
                    Project role:
                  </Text>
                  <Text size="small" weight="medium">
                    {projectRoleName}
                  </Text>
                </div>
              )}
              {scopeProjects.length > 0 && (
                <div className={styles.detailRow}>
                  <Text size="small" className={styles.detailLabel}>
                    Projects
                  </Text>
                  <div className={styles.chipGroup}>
                    {scopeProjects.map(project => (
                      <Chip
                        key={project.id}
                      >
                        {project.title}
                      </Chip>
                    ))}
                  </div>
                </div>
              )}
            </Flex>
          </div>

          <div className={styles.section}>
            <Flex
              className={styles.sectionHeader}
              direction="column"
              gap={3}
            >
              <Flex justify="between" align="center">
                <Text size="regular" weight="medium">
                  Expiry Date
                </Text>
                <Button
                  variant="outline"
                  color="neutral"
                  size="small"
                  onClick={() =>
                    regenerateDialogHandle.openWithPayload({
                      patId,
                      currentExpiryDays: expiryDays
                    })
                  }
                  data-test-id="frontier-sdk-pat-regenerate-btn"
                >
                  Regenerate
                </Button>
              </Flex>
              {createdOn && (
                <Text size="small">Created on: {createdOn}</Text>
              )}
              {expiryInfo && <Text size="small">{expiryInfo}</Text>}
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
