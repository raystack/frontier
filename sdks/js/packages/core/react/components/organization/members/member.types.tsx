import { User } from '~/src/types';

export type MembersType = {
  users: User[];
};

export enum MemberActionmethods {
  InviteMember = 'invite'
}

export type MembersTableType = {
  users: User[];
  setOpenInviteDialog: React.Dispatch<React.SetStateAction<boolean>>;
};
