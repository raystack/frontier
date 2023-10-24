import { Button, DataTable, EmptyState, Flex, Tooltip } from '@raystack/apsara';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useCallback, useMemo } from 'react';

import { usePermissions } from '~/react/hooks/usePermissions';
import { V1Beta1User } from '~/src';
import { Role } from '~/src/types';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { getColumns } from './member.columns';
import Skeleton from 'react-loading-skeleton';
import { AuthTooltipMessage } from '~/react/utils';

export type MembersProps = {
  members: V1Beta1User[];
  organizationId: string;
  memberRoles?: Record<string, Role[]>;
  isLoading?: boolean;
};

export const Members = ({
  members,
  organizationId,
  memberRoles = {},
  isLoading: isMemberLoading
}: MembersProps) => {
  let { teamId } = useParams({ from: '/teams/$teamId' });
  const navigate = useNavigate({ from: '/teams/$teamId' });

  const membersCount = members?.length || 0;

  const tableStyle = membersCount
    ? { width: '100%' }
    : { width: '100%', height: '100%' };

  const resource = `app/group:${teamId}`;
  const listOfPermissionsToCheck = useMemo(
    () => [
      {
        permission: PERMISSIONS.UpdatePermission,
        resource
      }
    ],
    [resource]
  );

  const { permissions, isFetching: isPermissionsFetching } = usePermissions(
    listOfPermissionsToCheck,
    !!teamId
  );
  const { canUpdateGroup } = useMemo(() => {
    return {
      canUpdateGroup: shouldShowComponent(
        permissions,
        `${PERMISSIONS.UpdatePermission}::${resource}`
      )
    };
  }, [permissions, resource]);

  const isLoading = isPermissionsFetching || isMemberLoading;

  const columns = useMemo(
    () =>
      getColumns({
        organizationId,
        canUpdateGroup,
        memberRoles,
        isLoading,
        membersCount
      }),
    [organizationId, canUpdateGroup, memberRoles, isLoading, membersCount]
  );

  const updatedUsers = useMemo(() => {
    return isLoading
      ? ([{ id: 1 }, { id: 2 }, { id: 3 }] as any)
      : members?.length
      ? members
      : [];
  }, [members, isLoading]);

  return (
    <Flex direction="column" style={{ paddingTop: '32px' }}>
      <DataTable
        data={updatedUsers}
        // @ts-ignore
        columns={columns}
        emptyState={noDataChildren}
        parentStyle={{ height: 'calc(100vh - 212px)' }}
        style={tableStyle}
      >
        <DataTable.Toolbar style={{ padding: 0, border: 0 }}>
          <Flex justify="between" gap="small">
            <Flex style={{ maxWidth: '360px', width: '100%' }}>
              <DataTable.GloabalSearch
                placeholder="Search by name or email"
                size="medium"
              />
            </Flex>
            {isLoading ? (
              <Skeleton height={'32px'} width={'64px'} />
            ) : (
              <Tooltip
                message={AuthTooltipMessage}
                side="left"
                disabled={canUpdateGroup}
              >
                <Button
                  variant="primary"
                  style={{ width: 'fit-content' }}
                  onClick={() =>
                    navigate({
                      to: '/teams/$teamId/invite',
                      params: { teamId: teamId }
                    })
                  }
                  disabled={!canUpdateGroup}
                >
                  Add Members
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
    <h3>0 members in your team</h3>
    <div className="pera">Try adding new members.</div>
  </EmptyState>
);
