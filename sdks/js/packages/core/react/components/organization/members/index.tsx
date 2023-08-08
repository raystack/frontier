'use client';

import {
  Button,
  DataTable,
  EmptyState,
  Flex,
  Separator,
  Text
} from '@raystack/apsara';
import { useState } from 'react';
import { styles } from '../styles';
import { columns } from './member.columns';
import type { MembersTableType, MembersType } from './member.types';

export default function WorkspaceMembers({ users }: MembersType) {
  const [_, setOpenInviteDialog] = useState(false);
  return (
    <Flex direction="column" gap="large" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size={6}>Members</Text>
      </Flex>
      <Flex direction="column" gap="large" style={styles.container}>
        <AllowedEmailDomains />
        <Separator></Separator>
        <ManageMembers />
        <MembersTable users={users} setOpenInviteDialog={setOpenInviteDialog} />
      </Flex>
    </Flex>
  );
}

const MembersHeading = () => (
  <Flex direction="column" gap="small" style={styles.container}>
    <Text size={10}>Members</Text>
    <Text size={4} style={{ color: 'var(--foreground-muted)' }}>
      Manage who has access to this workspace
    </Text>
  </Flex>
);

const AllowedEmailDomains = () => (
  <Flex direction="row" justify="between" align="center">
    <Flex direction="column" gap="small">
      <Text size={6}>Allowed email domains</Text>
      <Text size={4} style={{ color: 'var(--foreground-muted)' }}>
        Anyone with an email address at these domains is allowed to sign up for
        this workspace.
      </Text>
    </Flex>
    <Button size="medium" variant="primary">
      Add Domain
    </Button>
  </Flex>
);
const ManageMembers = () => (
  <Flex direction="row" justify="between" align="center">
    <Flex direction="column" gap="small">
      <Text size={6}>Manage members</Text>
      <Text size={4} style={{ color: 'var(--foreground-muted)' }}>
        Manage members for this domain.
      </Text>
    </Flex>
  </Flex>
);

const MembersTable = ({ users, setOpenInviteDialog }: MembersTableType) => {
  const tableStyle = users?.length
    ? { width: '100%' }
    : { width: '100%', height: '100%' };

  return (
    <Flex direction="row">
      <DataTable
        data={users ?? []}
        // @ts-ignore
        columns={columns}
        emptyState={noDataChildren}
        parentStyle={{ height: 'calc(100vh - 400px)' }}
        style={tableStyle}
      >
        <DataTable.Toolbar style={{ padding: 0, border: 0 }}>
          <Flex justify="between" gap="small">
            <DataTable.GloabalSearch
              placeholder="Search by name or email"
              type="small"
            />
            <Button
              variant="primary"
              onClick={() => setOpenInviteDialog(true)}
              style={{ width: 'fit-content' }}
            >
              Invite member
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
    <h3>0 members in your workspace</h3>
    <div className="pera">Try adding new members.</div>
  </EmptyState>
);
