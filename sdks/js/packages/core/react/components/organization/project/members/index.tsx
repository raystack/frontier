import { Button, DataTable, EmptyState, Flex } from '@raystack/apsara';
import { useNavigate, useParams } from '@tanstack/react-router';
import { V1Beta1User } from '~/src';
import { columns } from './member.columns';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { usePermissions } from '~/react/hooks/usePermissions';
import { useMemo } from 'react';

export type MembersProps = {
  members?: V1Beta1User[];
};

export const Members = ({ members }: MembersProps) => {
  const navigate = useNavigate({ from: '/projects/$projectId' });
  const { projectId } = useParams({ from: '/projects/$projectId' });

  const resource = `app/project:${projectId}`;
  const listOfPermissionsToCheck = [
    {
      permission: PERMISSIONS.UpdatePermission,
      resource
    }
  ];

  const { permissions } = usePermissions(listOfPermissionsToCheck, !!projectId);

  const { canUpdateProject } = useMemo(() => {
    return {
      canUpdateProject: shouldShowComponent(
        permissions,
        `${PERMISSIONS.UpdatePermission}::${resource}`
      )
    };
  }, [permissions, resource]);

  const tableStyle = members?.length
    ? { width: '100%' }
    : { width: '100%', height: '100%' };

  return (
    <Flex direction="column" style={{ paddingTop: '32px' }}>
      <DataTable
        data={members ?? []}
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
            {canUpdateProject ? (
              <Button
                variant="primary"
                style={{ width: 'fit-content' }}
                onClick={() =>
                  navigate({
                    to: '/projects/$projectId/invite',
                    params: { projectId: projectId }
                  })
                }
              >
                Add Team
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
