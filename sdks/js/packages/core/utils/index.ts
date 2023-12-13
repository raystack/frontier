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
  return arr
    .filter(user => !excludeIds.includes(user.id))
    .sort((a, b) =>
      (a.title || a.email || '').localeCompare(b.title || b.email || '')
    );
};

type Predicate<T> = (a: T, b: T) => boolean;
export const isEqualById = (a: any, b: any) => a.id === b.id;
export function differenceWith<T>(
  pred: Predicate<T>,
  list1: T[],
  list2: T[]
): T[] {
  return list1.filter(item1 => !list2.some(item2 => pred(item1, item2)));
}
  
export const PERMISSIONS = {
  // namespace
  PlatformNamespace: 'app/platform',
  OrganizationNamespace: 'app/organization',
  ProjectNamespace: 'app/project',
  GroupNamespace: 'app/group',
  RoleBindingNamespace: 'app/rolebinding',
  RoleNamespace: 'app/role',
  InvitationNamespace: 'app/invitation',
  UserNamespace: 'app/user',

  // relations
  PlatformRelationName: 'platform',
  AdminRelationName: 'admin',
  OrganizationRelationName: 'org',
  UserRelationName: 'user',
  ProjectRelationName: 'project',
  GroupRelationName: 'group',
  MemberRelationName: 'member',
  OwnerRelationName: 'owner',
  RoleRelationName: 'role',
  RoleGrantRelationName: 'granted',
  RoleBearerRelationName: 'bearer',

  // permissions
  ListPermission: 'list',
  GetPermission: 'get',
  CreatePermission: 'create',
  UpdatePermission: 'update',
  DeletePermission: 'delete',
  SudoPermission: 'superuser',
  RoleManagePermission: 'rolemanage',
  PolicyManagePermission: 'policymanage',
  ProjectListPermission: 'projectlist',
  GroupListPermission: 'grouplist',
  ProjectCreatePermission: 'projectcreate',
  GroupCreatePermission: 'groupcreate',
  ResourceListPermission: 'resourcelist',
  InvitationListPermission: 'invitationlist',
  InvitationCreatePermission: 'invitationcreate',
  AcceptPermission: 'accept',
  ServiceUserManagePermission: 'serviceusermanage',
  ManagePermission: 'manage',

  // synthetic permission
  MembershipPermission: 'membership',

  // principals
  UserPrincipal: 'app/user',
  ServiceUserPrincipal: 'app/serviceuser',
  GroupPrincipal: 'app/group',
  SuperUserPrincipal: 'app/superuser',

  // Roles
  RoleProjectOwner: 'app_project_owner',
  RoleGroupMember: 'app_group_member',
  RoleProjectViewer: 'app_project_viewer'
};

export const formatPermissions = (
  permisions: { body: any; status: boolean }[] = []
): Record<string, boolean> =>
  permisions.reduce((acc: any, p: any) => {
    const { body, status } = p;
    acc[`${body.permission}::${body.resource}`] = status;
    return acc;
  }, {});

export const shouldShowComponent = (
  permissions: Record<string, boolean> = {},
  permisionsRequired: string
) => {
  return permissions[permisionsRequired] === true;
};
