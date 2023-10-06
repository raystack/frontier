import { Button, DataTable, EmptyState, Flex } from '@raystack/apsara';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useMemo } from 'react';

import { usePermissions } from '~/react/hooks/usePermissions';
import { V1Beta1User } from '~/src';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { getColumns } from './member.columns';

export type MembersProps = {
  members: V1Beta1User[];
  organizationId?: string;
};

export const Members = ({ members, organizationId }: MembersProps) => {
  let { teamId } = useParams({ from: '/teams/$teamId' });
  const navigate = useNavigate({ from: '/teams/$teamId' });

  const tableStyle = members?.length
    ? { width: '100%' }
    : { width: '100%', height: '100%' };

  const resource = `app/group:${teamId}`;
  const listOfPermissionsToCheck = [
    {
      permission: PERMISSIONS.UpdatePermission,
      resource
    }
  ];

  const { permissions } = usePermissions(listOfPermissionsToCheck, !!teamId);
  const { canUpdateGroup } = useMemo(() => {
    return {
      canUpdateGroup: shouldShowComponent(
        permissions,
        `${PERMISSIONS.UpdatePermission}::${resource}`
      )
    };
  }, [permissions, resource]);

  return (
    <Flex direction="column" style={{ paddingTop: '32px' }}>
      <DataTable
        data={members ?? []}
        // @ts-ignore
        columns={getColumns(organizationId, canUpdateGroup)}
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
            {canUpdateGroup ? (
              <Button
                variant="primary"
                style={{ width: 'fit-content' }}
                onClick={() =>
                  navigate({
                    to: '/teams/$teamId/invite',
                    params: { teamId: teamId }
                  })
                }
              >
                Add Members
              </Button>
            ) : null}
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
