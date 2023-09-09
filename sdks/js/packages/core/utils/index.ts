import { V1Beta1User } from '~/src';

export const hasWindow = (): boolean => typeof window !== 'undefined';

export function capitalize(str: string) {
  return str.charAt(0).toUpperCase() + str.slice(1).toLowerCase();
}

export const getInitials = function (name: string = '') {
  var names = name.split(' '),
    initials = names[0].substring(0, 1).toUpperCase();

  if (names.length > 1) {
    initials += names[names.length - 1].substring(0, 1).toUpperCase();
  }
  return initials;
};

export const filterUsersfromUsers = (
  arr: V1Beta1User[] = [],
  exclude: V1Beta1User[] = []
) => {
  const excludeIds = exclude.map(e => e.id);
  return arr.filter(user => !excludeIds.includes(user.id));
};


export const PERMISSIONS = {
  ADMINISTER: 'administer',
  GROUPCREATE: 'groupcreate',
  GROUPLIST: 'grouplist',
  INVITATIONCREATE: 'invitationcreate',
  INVITATIONLIST: 'invitationlist',
  POLICYMANAGE: 'policymanage',
  PROJECTCREATE: 'projectcreate',
  PROJECTLIST: 'projectlist',
  RESOURCELIST: 'resourcelist',
  ROLEMANAGE: 'rolemanage',
  SERVICEUSERMANAGE: 'serviceusermanage',
  GET: 'get',
  PUT: 'put',
  POST: 'post',
  UPDATE: 'update',
  DELETE: 'delete'
};

export const formatPermissions = (
  permisions: { body: any; status: boolean }[] = []
): Record<string, string> =>
  permisions.reduce((acc: any, p: any) => {
    const { body, status } = p;
    acc[`${body.permission}::${body.resource}`] = status;
    return acc;
  }, {});