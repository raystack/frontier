import { MemberWithInvite } from '~/react/hooks/useOrganizationMembers';
import {  V1Beta1User, V1Beta1Role } from '~/src';

export type MembersType = {
  users: V1Beta1User[];
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
  memberRoles: Record<string, V1Beta1Role[]>;
  roles: V1Beta1Role[];
  refetch?: () => void;
};
