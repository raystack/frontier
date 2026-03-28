'use client';

import { useMemo, useState } from 'react';
import { ExclamationTriangleIcon, MagnifyingGlassIcon } from '@radix-ui/react-icons';
import {
  Button,
  Tooltip,
  Skeleton,
  Flex,
  InputField,
  Select,
  EmptyState,
  toastManager
} from '@raystack/apsara-v1';
import type { Role } from '@raystack/proton/frontier';
import { useFrontier } from '../../contexts/FrontierContext';
import { useOrganizationMembers } from '../../hooks/useOrganizationMembers';
import { usePermissions } from '../../hooks/usePermissions';
import { AuthTooltipMessage } from '../../utils';
import { PERMISSIONS, shouldShowComponent } from '../../../utils';
import { ViewContainer } from '../../components/view-container';
import { ViewHeader } from '../../components/view-header';
import { MemberListItem } from './components/member-list-item';
import { InviteMemberDialog } from './components/invite-member-dialog';
import { RemoveMemberDialog } from './components/remove-member-dialog';
import { UpdateRoleDialog } from './components/update-role-dialog';
import styles from './members-view.module.css';

export interface MembersViewProps {
  showTeamField?: boolean;
}

export function MembersView({ showTeamField = true }: MembersViewProps) {
  const { activeOrganization: organization } = useFrontier();

  const resource = `app/organization:${organization?.id}`;
  const listOfPermissionsToCheck = useMemo(
    () => [
      {
        permission: PERMISSIONS.InvitationCreatePermission,
        resource
      },
      {
        permission: PERMISSIONS.UpdatePermission,
        resource
      }
    ],
    [resource]
  );

  const { permissions, isFetching: isPermissionsFetching } = usePermissions(
    listOfPermissionsToCheck,
    !!organization?.id
  );

  const { canCreateInvite, canDeleteUser } = useMemo(() => {
    return {
      canCreateInvite: shouldShowComponent(
        permissions,
        `${PERMISSIONS.InvitationCreatePermission}::${resource}`
      ),
      canDeleteUser: shouldShowComponent(
        permissions,
        `${PERMISSIONS.UpdatePermission}::${resource}`
      )
    };
  }, [permissions, resource]);

  const {
    roles,
    members,
    memberRoles,
    refetch,
    isFetching: isOrgMembersLoading
  } = useOrganizationMembers({
    showInvitations: canCreateInvite
  });

  const isLoading = isOrgMembersLoading || isPermissionsFetching;

  const [showInviteDialog, setShowInviteDialog] = useState(false);
  const [removeMemberState, setRemoveMemberState] = useState<{
    open: boolean;
    memberId: string;
    invited: string;
  }>({ open: false, memberId: '', invited: 'false' });
  const [updateRoleState, setUpdateRoleState] = useState<{
    open: boolean;
    memberId: string;
    role: Role | null;
  }>({ open: false, memberId: '', role: null });

  const [searchQuery, setSearchQuery] = useState('');
  const [roleFilter, setRoleFilter] = useState('all');

  const filteredMembers = useMemo(() => {
    let result = members;

    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      result = result.filter(
        member =>
          member.title?.toLowerCase().includes(query) ||
          member.email?.toLowerCase().includes(query) ||
          member.userId?.toLowerCase().includes(query)
      );
    }

    if (roleFilter !== 'all') {
      result = result.filter(member => {
        if (member.invited) return false;
        const userRoles = member.id ? memberRoles[member.id] : [];
        return userRoles?.some(r => r.id === roleFilter);
      });
    }

    return result;
  }, [members, searchQuery, roleFilter, memberRoles]);

  const getRoleDisplay = (member: typeof members[number]): string => {
    if (member.invited) {
      const inviteRoleIds = (member as { roleIds?: string[] }).roleIds;
      if (inviteRoleIds?.length) {
        return inviteRoleIds
          .map(id => roles.find(r => r.id === id))
          .filter(Boolean)
          .map(r => r?.title || r?.name)
          .join(', ') || 'Member';
      }
      return 'Member';
    }
    if (member.id && memberRoles[member.id]) {
      return memberRoles[member.id]
        .map((r: Role) => r.title || r.name)
        .join(', ');
    }
    return 'Inherited role';
  };

  const handleRemoveMember = (memberId: string, invited: string) => {
    setRemoveMemberState({ open: true, memberId, invited });
  };

  const handleUpdateRole = (memberId: string, role: Role) => {
    setUpdateRoleState({ open: true, memberId, role });
  };

  const handleInviteOpenChange = (value: boolean) => {
    setShowInviteDialog(value);
    if (!value) refetch();
  };

  const handleRemoveOpenChange = (value: boolean) => {
    setRemoveMemberState({ open: value, memberId: '', invited: 'false' });
    if (!value) refetch();
  };

  const handleUpdateRoleOpenChange = (value: boolean) => {
    setUpdateRoleState({ open: value, memberId: '', role: null });
    if (!value) refetch();
  };

  return (
    <ViewContainer>
      <ViewHeader
        title="Members"
        description="Manage members for this domain."
      />

      <Flex direction="column" gap={7}>
        <Flex justify="between" align="center">
          <Flex gap={3} align="center">
            {isLoading ? (
              <>
                <Skeleton height="36px" width="280px" />
                <Skeleton height="36px" width="80px" />
              </>
            ) : (
              <>
                <InputField
                  placeholder="Search by name or email"
                  size="large"
                  leadingIcon={<MagnifyingGlassIcon />}
                  value={searchQuery}
                  onChange={e => setSearchQuery(e.target.value)}
                  width="280px"
                />
                <Select
                  value={roleFilter}
                  onValueChange={setRoleFilter}
                >
                  <Select.Trigger className={styles.roleFilter}>
                    <Select.Value placeholder="All" />
                  </Select.Trigger>
                  <Select.Content>
                    <Select.Item value="all">All</Select.Item>
                    {roles.map(role => (
                      <Select.Item key={role.id} value={role.id || ''}>
                        {role.title || role.name}
                      </Select.Item>
                    ))}
                  </Select.Content>
                </Select>
              </>
            )}
          </Flex>
          {isLoading ? (
            <Skeleton height="36px" width="120px" />
          ) : (
            <Tooltip>
              <Tooltip.Trigger
                disabled={canCreateInvite}
                render={<span />}
              >
                <Button
                  variant="solid"
                  color="accent"
                  onClick={() => setShowInviteDialog(true)}
                  disabled={!canCreateInvite}
                  data-test-id="frontier-sdk-invite-member-btn"
                >
                  Invite people
                </Button>
              </Tooltip.Trigger>
              {!canCreateInvite && (
                <Tooltip.Content>{AuthTooltipMessage}</Tooltip.Content>
              )}
            </Tooltip>
          )}
        </Flex>

        {isLoading ? (
          <Flex direction="column">
            {Array.from({ length: 5 }).map((_, i) => (
              <Flex
                key={i}
                align="center"
                gap={4}
                style={{
                  padding: 'var(--rs-space-3)',
                  borderBottom: '0.5px solid var(--rs-color-border-base-primary)',
                  minHeight: '48px'
                }}
              >
                <Skeleton
                  width="32px"
                  height="32px"
                  style={{ borderRadius: '50%', flexShrink: 0 }}
                />
                <Flex direction="column" gap={1} style={{ flex: 1 }}>
                  <Skeleton height="16px" width="120px" />
                  <Skeleton height="14px" width="180px" />
                </Flex>
                <Skeleton height="16px" width="60px" />
              </Flex>
            ))}
          </Flex>
        ) : filteredMembers.length === 0 ? (
          <EmptyState
            icon={<ExclamationTriangleIcon />}
            heading="No members found"
            subHeading="Get started by adding your first member"
          />
        ) : (
          <Flex direction="column">
            {filteredMembers.map(member => {
              const memberId = member.id || '';
              const userRoles = memberId ? memberRoles[memberId] : [];
              const excludedRoles = roles.filter(
                r => !userRoles?.some(ur => ur.id === r.id)
              );

              return (
                <MemberListItem
                  key={memberId || member.userId}
                  member={member}
                  roleDisplay={getRoleDisplay(member)}
                  excludedRoles={excludedRoles}
                  canUpdateRole={canDeleteUser && !member.invited}
                  canRemove={canDeleteUser}
                  onUpdateRole={role => handleUpdateRole(memberId, role)}
                  onRemove={() =>
                    handleRemoveMember(
                      memberId,
                      String(member.invited || false)
                    )
                  }
                />
              );
            })}
          </Flex>
        )}
      </Flex>

      <InviteMemberDialog
        open={showInviteDialog}
        onOpenChange={handleInviteOpenChange}
        showTeamField={showTeamField}
      />
      <RemoveMemberDialog
        open={removeMemberState.open}
        onOpenChange={handleRemoveOpenChange}
        memberId={removeMemberState.memberId}
        invited={removeMemberState.invited}
      />
      {updateRoleState.role && organization?.id && (
        <UpdateRoleDialog
          open={updateRoleState.open}
          onOpenChange={handleUpdateRoleOpenChange}
          memberId={updateRoleState.memberId}
          role={updateRoleState.role}
          organizationId={organization.id}
          refetch={refetch}
        />
      )}
    </ViewContainer>
  );
}
