'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import { ExclamationTriangleIcon, Pencil1Icon } from '@radix-ui/react-icons';
import {
  Button,
  Tooltip,
  Skeleton,
  Flex,
  EmptyState,
  DataTable,
  Dialog,
  AlertDialog,
  Image,
  Menu,
  Select
} from '@raystack/apsara-v1';
import deleteIcon from '../../assets/delete.svg';
import { toastManager } from '@raystack/apsara-v1';
import { useFrontier } from '../../contexts/FrontierContext';
import { useOrganizationTeams } from '../../hooks/useOrganizationTeams';
import { usePermissions } from '../../hooks/usePermissions';
import { AuthTooltipMessage } from '../../utils';
import { PERMISSIONS, shouldShowComponent } from '../../../utils';
import { ViewContainer } from '../../components/view-container';
import { ViewHeader } from '../../components/view-header';
import { getColumns, type TeamMenuPayload } from './components/team-columns';
import { AddTeamDialog } from './components/add-team-dialog';
import { EditTeamDialog, type EditTeamPayload } from './components/edit-team-dialog';
import { DeleteTeamDialog, type DeleteTeamPayload } from './components/delete-team-dialog';
import styles from './teams-view.module.css';
import { useTerminology } from '~/react/hooks/useTerminology';

const teamsFilterOptions = [
  { value: 'my-teams', label: 'My Teams' },
  { value: 'all-teams', label: 'All Teams' }
];

const teamMenuHandle = Menu.createHandle<TeamMenuPayload>();
const addTeamDialogHandle = Dialog.createHandle();
const editTeamDialogHandle = Dialog.createHandle<EditTeamPayload>();
const deleteTeamDialogHandle = AlertDialog.createHandle<DeleteTeamPayload>();

export interface TeamsViewProps {
  title?: string;
  description?: string;
  onTeamClick?: (teamId: string) => void;
}

export function TeamsView({
  title = 'Teams',
  description,
  onTeamClick
}: TeamsViewProps) {
  const [showOrgTeams, setShowOrgTeams] = useState(false);
  const t = useTerminology();

  const {
    isFetching: isTeamsLoading,
    teams,
    userAccessOnTeam,
    refetch,
    error: teamsError
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
        `${PERMISSIONS.GroupListPermission}::${resource}`
      )
    };
  }, [permissions, resource]);

  useEffect(() => {
    if (teamsError) {
      toastManager.add({
        title: 'Something went wrong',
        description:
          teamsError instanceof Error
            ? teamsError.message
            : 'Failed to load teams',
        type: 'error'
      });
    }
  }, [teamsError]);

  const onFilterChange = useCallback((value: string) => {
    setShowOrgTeams(value === 'all-teams');
  }, []);

  const isLoading = !organization?.id || isPermissionsFetching || isTeamsLoading;

  const columns = useMemo(
    () =>
      getColumns({
        userAccessOnTeam,
        menuHandle: teamMenuHandle
      }),
    [userAccessOnTeam]
  );

  return (
    <ViewContainer>
      <ViewHeader
        title={title}
        description={
          description ??
          `Manage teams in this ${t.organization({ case: 'lower' })}`
        }
      />

      <DataTable
        data={teams ?? []}
        columns={columns}
        isLoading={isLoading}
        defaultSort={{ name: 'title', order: 'asc' }}
        mode="client"
        onRowClick={row => onTeamClick?.(row.id)}
      >
        <Flex direction="column" gap={7}>
          <Flex justify="between" gap={3}>
            <Flex gap={3} align="center">
              {isLoading ? (
                <Skeleton height="34px" width="360px" />
              ) : (
                <>
                  <DataTable.Search
                    placeholder="Search by title"
                    size="large"
                    width={360}
                  />
                  {canListOrgGroups && (
                    <Select
                      defaultValue={teamsFilterOptions[0].value}
                      onValueChange={onFilterChange}
                    >
                      <Select.Trigger className={styles.teamsFilter}>
                        <Select.Value />
                      </Select.Trigger>
                      <Select.Content>
                        {teamsFilterOptions.map(opt => (
                          <Select.Item value={opt.value} key={opt.value}>
                            {opt.label}
                          </Select.Item>
                        ))}
                      </Select.Content>
                    </Select>
                  )}
                </>
              )}
            </Flex>
            {isLoading ? (
              <Skeleton height="34px" width="120px" />
            ) : (
              <Tooltip>
                <Tooltip.Trigger
                  disabled={canCreateGroup}
                  render={<span />}
                >
                  <Button
                    variant="solid"
                    color="accent"
                    onClick={() => addTeamDialogHandle.open(null)}
                    disabled={!canCreateGroup}
                    data-test-id="frontier-sdk-add-team-button"
                  >
                    Add team
                  </Button>
                </Tooltip.Trigger>
                {!canCreateGroup && (
                  <Tooltip.Content>{AuthTooltipMessage}</Tooltip.Content>
                )}
              </Tooltip>
            )}
          </Flex>
          <DataTable.Content
            emptyState={
              <EmptyState
                icon={<ExclamationTriangleIcon />}
                heading="No teams found"
                subHeading="Get started by creating your first team."
              />
            }
            classNames={{
              root: styles.tableRoot
            }}
          />
        </Flex>
      </DataTable>

      <Menu handle={teamMenuHandle} modal={false}>
        {({ payload: rawPayload }) => {
          const payload = rawPayload as TeamMenuPayload | undefined;
          return (
            <Menu.Content align="end" className={styles.menuContent}>
              {payload?.canUpdate && (
                <Menu.Item
                  leadingIcon={<Pencil1Icon />}
                  onClick={() =>
                    editTeamDialogHandle.openWithPayload({
                      teamId: payload.teamId,
                      title: payload.title,
                      name: ''
                    })
                  }
                  data-test-id="frontier-sdk-edit-team-dropdown-item"
                >
                  Edit
                </Menu.Item>
              )}
              {payload?.canDelete && (
                <Menu.Item
                  leadingIcon={
                    <Image
                      src={deleteIcon as unknown as string}
                      alt="Delete"
                      width={16}
                      height={16}
                    />
                  }
                  onClick={() =>
                    deleteTeamDialogHandle.openWithPayload({
                      teamId: payload.teamId
                    })
                  }
                  data-test-id="frontier-sdk-delete-team-dropdown-item"
                  style={{
                    color: 'var(--rs-color-foreground-danger-primary)'
                  }}
                >
                  Delete team
                </Menu.Item>
              )}
            </Menu.Content>
          );
        }}
      </Menu>

      <AddTeamDialog handle={addTeamDialogHandle} refetch={refetch} />
      <EditTeamDialog handle={editTeamDialogHandle} refetch={refetch} />
      <DeleteTeamDialog handle={deleteTeamDialogHandle} refetch={refetch} />
    </ViewContainer>
  );
}
