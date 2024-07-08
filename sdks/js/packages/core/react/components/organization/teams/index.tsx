'use client';

import {
  Button,
  DataTable,
  EmptyState,
  Flex,
  Select,
  Text,
  Tooltip
} from '@raystack/apsara';
import { Outlet, useNavigate, useRouterState } from '@tanstack/react-router';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';

import { useOrganizationTeams } from '~/react/hooks/useOrganizationTeams';
import { usePermissions } from '~/react/hooks/usePermissions';
import { V1Beta1Group } from '~/src';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { styles } from '../styles';
import { getColumns } from './teams.columns';
import { AuthTooltipMessage } from '~/react/utils';
import Skeleton from 'react-loading-skeleton';

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
      <Flex style={styles.header}>
        <Text size={6}>Teams</Text>
      </Flex>
      <Flex direction="column" gap="large" style={styles.container}>
        <Flex direction="column" style={{ gap: '24px' }}>
          <TeamsTable
            teams={teams}
            isLoading={isLoading}
            canCreateGroup={canCreateGroup}
            userAccessOnTeam={userAccessOnTeam}
            canListOrgGroups={canListOrgGroups}
            onOrgTeamsFilterChange={onOrgTeamsFilterChange}
          />
        </Flex>
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
  let navigate = useNavigate({ from: '/teams' });

  const tableStyle = teams?.length
    ? { width: '100%' }
    : { width: '100%', height: '100%' };

  const columns = useMemo(
    () => getColumns(userAccessOnTeam),
    [userAccessOnTeam]
  );

  return (
    <Flex direction="row">
      <DataTable
        data={teams ?? []}
        isLoading={isLoading}
        columns={columns}
        emptyState={noDataChildren}
        parentStyle={{ height: 'calc(100vh - 180px)' }}
        style={tableStyle}
      >
        <DataTable.Toolbar
          style={{ padding: 0, border: 0, marginBottom: 'var(--pd-16)' }}
        >
          <Flex justify="between" gap="small">
            <Flex
              style={{
                maxWidth: canListOrgGroups ? '500px' : '360px',
                width: '100%'
              }}
              gap={'medium'}
            >
              <DataTable.GloabalSearch
                placeholder="Search by name"
                size="medium"
              />
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
              <Skeleton height={'32px'} width={'64px'} />
            ) : (
              <Tooltip
                message={AuthTooltipMessage}
                side="left"
                disabled={canCreateGroup}
              >
                <Button
                  variant="primary"
                  style={{ width: 'fit-content', height: '100%' }}
                  disabled={!canCreateGroup}
                  onClick={() => navigate({ to: '/teams/modal' })}
                >
                  Add team
                </Button>
              </Tooltip>
            )}
          </Flex>
        </DataTable.Toolbar>
      </DataTable>
    </Flex>
  );
};

const noDataChildren = (
  <EmptyState>
    <div className="svg-container"></div>
    <h3>0 teams in your organization</h3>
    <div className="pera">Try adding new team.</div>
  </EmptyState>
);
