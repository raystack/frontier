import { Button, DataTable, EmptyState, Flex } from '@raystack/apsara';
import { useNavigate } from 'react-router-dom';
import { V1Beta1User } from '~/src';
import { columns } from './member.columns';

export type MembersProps = {
  members?: V1Beta1User[];
};

export const Members = ({ members }: MembersProps) => {
  let navigate = useNavigate();

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
        parentStyle={{ height: 'calc(100vh - 400px)' }}
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

            <Button
              variant="primary"
              style={{ width: 'fit-content' }}
              onClick={() => {}}
            >
              Invite people
            </Button>
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
