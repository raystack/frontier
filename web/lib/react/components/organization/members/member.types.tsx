import { MemberWithInvite } from '~/react/hooks/useOrganizationMembers';
import { User, Role } from '@raystack/proton/frontier';

export type MembersType = {
  users: User[];
};

export enum MemberActionmethods {
  InviteMember = 'invite'
}

export type MembersTableType = {
  isLoading?: boolean;
  users: MemberWithInvite[];
  organizationId: string;
  canCreateInvite?: boolean;
  canDeleteUser?: boolean;
  memberRoles: Record<string, Role[]>;
  roles: Role[];
  refetch?: () => void;
};
