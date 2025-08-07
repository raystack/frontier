'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Tooltip,
  Skeleton,
  EmptyState,
  Text,
  Flex,
  Button,
  Select,
  DataTable
} from '@raystack/apsara';
import { Outlet, useNavigate, useRouterState } from '@tanstack/react-router';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { useFrontier } from '~/react/contexts/FrontierContext';

import { useOrganizationTeams } from '~/react/hooks/useOrganizationTeams';
import { usePermissions } from '~/react/hooks/usePermissions';
import type { V1Beta1Group } from '~/src';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { getColumns } from './teams.columns';
import { AuthTooltipMessage } from '~/react/utils';
import styles from './teams.module.css';

const teamsSelectOptions = [
  { value: 'my-teams', label: 'My Teams' },
  { value: 'all-teams', label: 'All Teams' }
];

interface WorkspaceTeamProps {
  teams: V1Beta1Group[];
  isLoading?: boolean;
  canCreateGroup?: boolean;
  userAccessOnTeam: Record<string, string[]>;
  canListOrgGroups?: boolean;
  onOrgTeamsFilterChange: (value: string) => void;
}

export default function WorkspaceTeams() {
  const [showOrgTeams, setShowOrgTeams] = useState(false);

  const routerState = useRouterState();

  const isListRoute = useMemo(() => {
    return routerState.location.pathname === '/teams';
  }, [routerState.location.pathname]);

  const {
    isFetching: isTeamsLoading,
    teams,
    userAccessOnTeam,
    refetch
  } = useOrganizationTeams({
    withPermissions: ['update', 'delete'],
    showOrgTeams,
    withMemberCount: true
  });
  const { activeOrganization: organization } = useFrontier();

  const resource = `app/organization:${organization?.id}`;
  const listOfPermissionsToCheck = useMemo(
    () => [
      {
        permission: PERMISSIONS.GroupCreatePermission,
        resource
      },
      {
        permission: PERMISSIONS.GroupListPermission,
        resource
      }
    ],
    [resource]
  );

  const { permissions, isFetching: isPermissionsFetching } = usePermissions(
    listOfPermissionsToCheck,
    !!organization?.id
  );

  const { canCreateGroup, canListOrgGroups } = useMemo(() => {
    return {
      canCreateGroup: shouldShowComponent(
        permissions,
        `${PERMISSIONS.GroupCreatePermission}::${resource}`
      ),
      canListOrgGroups: shouldShowComponent(
        permissions,
        `${PERMISSIONS.GroupCreatePermission}::${resource}`
      )
    };
  }, [permissions, resource]);

  const onOrgTeamsFilterChange = useCallback((value: string) => {
    if (value === 'all-teams') {
      setShowOrgTeams(true);
    } else {
      setShowOrgTeams(false);
    }
  }, []);

  useEffect(() => {
    if (isListRoute) {
      refetch();
    }
  }, [isListRoute, refetch, routerState.location.state.key]);

  const isLoading = isPermissionsFetching || isTeamsLoading;

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex className={styles.header}>
        <Text size={6}>Teams</Text>
      </Flex>
      <Flex direction="column" gap={9} className={styles.container}>
        <TeamsTable
          teams={teams}
          isLoading={isLoading}
          canCreateGroup={canCreateGroup}
          userAccessOnTeam={userAccessOnTeam}
          canListOrgGroups={canListOrgGroups}
          onOrgTeamsFilterChange={onOrgTeamsFilterChange}
        />
      </Flex>
      <Outlet />
    </Flex>
  );
}

const TeamsTable = ({
  teams,
  isLoading,
  canCreateGroup,
  userAccessOnTeam,
  canListOrgGroups,
  onOrgTeamsFilterChange
}: WorkspaceTeamProps) => {
  const navigate = useNavigate({ from: '/teams' });

  const columns = useMemo(
    () => getColumns(userAccessOnTeam),
    [userAccessOnTeam]
  );

  return (
    <DataTable
      data={teams ?? []}
      isLoading={isLoading}
      columns={columns}
      defaultSort={{ name: 'name', order: 'asc' }}
    >
      <Flex direction="column" gap={7} className={styles.tableWrapper}>
        <Flex justify="between" gap={3}>
          <Flex gap={3} justify="start" className={styles.tableSearchWrapper}>
            {isLoading ? (
              <Skeleton height="34px" width="500px" />
            ) : (
              <DataTable.Search placeholder="Search by title" size="medium" />
            )}
            {canListOrgGroups ? (
              <Select
                defaultValue={teamsSelectOptions[0].value}
                onValueChange={onOrgTeamsFilterChange}
              >
                <Select.Trigger style={{ minWidth: '140px' }}>
                  <Select.Value />
                </Select.Trigger>
                <Select.Content>
                  {teamsSelectOptions.map(opt => (
                    <Select.Item value={opt.value} key={opt.value}>
                      {opt.label}
                    </Select.Item>
                  ))}
                </Select.Content>
              </Select>
            ) : null}
          </Flex>
          {isLoading ? (
            <Skeleton height="34px" width="64px" />
          ) : (
            <Tooltip
              message={AuthTooltipMessage}
              side="left"
              disabled={canCreateGroup}
            >
              <Button
                disabled={!canCreateGroup}
                onClick={() => navigate({ to: '/teams/modal' })}
                data-test-id="frontier-sdk-add-team-btn"
              >
                Add team
              </Button>
            </Tooltip>
          )}
        </Flex>
        <DataTable.Content
          emptyState={noDataChildren}
          classNames={{
            root: styles.tableRoot,
            header: styles.tableHeader
          }}
        />
      </Flex>
    </DataTable>
  );
};

const noDataChildren = (
  <EmptyState
    icon={<ExclamationTriangleIcon />}
    heading="No teams found"
    subHeading="Get started by creating your first team."
  />
);
