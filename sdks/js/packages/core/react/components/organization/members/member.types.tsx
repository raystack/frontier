import { User } from '~/src/types';

export type MembersType = {
  users: User[];
};

export enum MemberActionmethods {
  InviteMember = 'invite'
}

export type MembersTableType = {
  isLoading?: boolean;
  users: User[];
  organizationId: string;
  canCreateInvite?: boolean;
  canDeleteUser?: boolean;
};
