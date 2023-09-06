import { User } from '~/src/types';

export type MembersType = {
  users: User[];
};

export enum MemberActionmethods {
  InviteMember = 'invite'
}

export type MembersTableType = {
  users: User[];
  organizationId?: string;
};
