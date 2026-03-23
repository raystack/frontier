'use client';

import { useCallback, useMemo, useState } from 'react';
import {
    Tooltip,
    Skeleton,
    EmptyState,
    Flex,
    Button,
    Select,
    DataTable
} from '@raystack/apsara';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { useFrontier } from '~/react/contexts/FrontierContext';

import { useOrganizationTeams } from '~/react/hooks/useOrganizationTeams';
import { usePermissions } from '~/react/hooks/usePermissions';
import type { Group } from '@raystack/proton/frontier';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { getColumns } from './teams-columns';
import { AuthTooltipMessage } from '~/react/utils';
import { PageHeader } from '~/react/components/common/page-header';
import { AddTeamDialog } from './add-team-dialog';
import { DeleteTeamDialog } from '../details/delete-team-dialog';
import sharedStyles from '../../../components/organization/styles.module.css';
import styles from './teams.module.css';
import { useTerminology } from '~/react/hooks/useTerminology';

const teamsSelectOptions = [
    { value: 'my-teams', label: 'My Teams' },
    { value: 'all-teams', label: 'All Teams' }
];

interface TeamsTableProps {
    teams: Group[];
    isLoading?: boolean;
    canCreateGroup?: boolean;
    userAccessOnTeam: Record<string, string[]>;
    canListOrgGroups?: boolean;
    onOrgTeamsFilterChange: (value: string) => void;
    onTeamClick?: (teamId: string) => void;
    onDeleteTeamClick?: (teamId: string) => void;
    onAddTeamClick?: () => void;
}

export interface TeamsListPageProps {
    title?: string;
    description?: string;
    onTeamClick?: (teamId: string) => void;
}

export function TeamsListPage({
    title,
    description,
    onTeamClick
}: TeamsListPageProps = {}) {
    const [showOrgTeams, setShowOrgTeams] = useState(false);
    const t = useTerminology();

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

    const isLoading = isPermissionsFetching || isTeamsLoading;

    const [showAddTeamDialog, setShowAddTeamDialog] = useState(false);
    const [deleteTeamState, setDeleteTeamState] = useState<{
        open: boolean;
        teamId: string;
    }>({ open: false, teamId: '' });

    const handleAddTeamOpenChange = (value: boolean) => {
        setShowAddTeamDialog(value);
        refetch();
    };

    const handleDeleteTeamOpenChange = (value: boolean) => {
        setDeleteTeamState({ open: value, teamId: '' });
        refetch();
    };

    const handleDeleteTeamClick = (teamId: string) => {
        setDeleteTeamState({ open: true, teamId });
    };

    return (
        <Flex direction="column" className={sharedStyles.pageWrapper}>
            <Flex
                direction="column"
                className={`${sharedStyles.container} ${sharedStyles.containerFlex}`}
            >
                <Flex
                    direction="row"
                    justify="between"
                    align="center"
                    className={sharedStyles.header}
                >
                    <PageHeader
                        title={title || 'Teams'}
                        description={description || `Manage teams in this ${t.organization({
                            case: 'lower'
                        })}.`}
                    />
                </Flex>
                <Flex
                    direction="column"
                    gap={9}
                    className={sharedStyles.contentWrapper}
                >
                    <TeamsTable
                        teams={teams}
                        isLoading={isLoading}
                        canCreateGroup={canCreateGroup}
                        userAccessOnTeam={userAccessOnTeam}
                        canListOrgGroups={canListOrgGroups}
                        onOrgTeamsFilterChange={onOrgTeamsFilterChange}
                        onTeamClick={onTeamClick}
                        onDeleteTeamClick={handleDeleteTeamClick}
                        onAddTeamClick={() => setShowAddTeamDialog(true)}
                    />
                </Flex>
            </Flex>
            <AddTeamDialog
                open={showAddTeamDialog}
                onOpenChange={handleAddTeamOpenChange}
            />
            <DeleteTeamDialog
                open={deleteTeamState.open}
                onOpenChange={handleDeleteTeamOpenChange}
                teamId={deleteTeamState.teamId}
                onDeleteSuccess={() => refetch()}
            />
        </Flex>
    );
}

const TeamsTable = ({
    teams,
    isLoading,
    canCreateGroup,
    userAccessOnTeam,
    canListOrgGroups,
    onOrgTeamsFilterChange,
    onTeamClick,
    onDeleteTeamClick,
    onAddTeamClick
}: TeamsTableProps) => {
    const columns = useMemo(
        () => getColumns(userAccessOnTeam, onTeamClick, onDeleteTeamClick),
        [userAccessOnTeam, onTeamClick, onDeleteTeamClick]
    );

  return (
    <DataTable
      data={teams ?? []}
      isLoading={isLoading}
      columns={columns}
      defaultSort={{ name: 'name', order: 'asc' }}
      onRowClick={row => onTeamClick?.(row.id)}
    >
      <Flex direction="column" gap={7} className={styles.tableWrapper}>
        <Flex justify="between" gap={3}>
          <Flex gap={3} justify="start" className={styles.tableSearchWrapper}>
            {isLoading ? (
              <Skeleton height="34px" width="500px" />
            ) : (
              <DataTable.Search placeholder="Search by title" />
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
                onClick={onAddTeamClick}
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

