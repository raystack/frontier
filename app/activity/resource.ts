import { getManager } from 'typeorm';
import { Activity } from '../../model/activity';
import Constants from '../../utils/constant';
import { delta } from '../../utils/deep-diff';
import { User } from '../../model/user';
import { Role } from '../../model/role';
import { Group } from '../../model/group';

type ActivityType = {
  id: string;
  reason: string;
  createdAt: string;
  diff: {
    [key: string]: string[] | undefined;
  };
  user: string;
};

const excludeFields = ['createdAt', 'updatedAt'];

export const actions = {
  CREATE: 'create',
  EDIT: 'edit',
  DELETE: 'delete'
};

const titleMap = {
  ASSIGNED_ROLE: 'Assigned a role',
  ASSIGNED_USER: 'Assigned a user',
  ADD_ATTRIBUTE_TO_GROUP: 'Added attribute to a team',
  REMOVED_ROLE: 'Removed a role',
  REMOVED_USER: 'Removed a user',
  REMOVED_ATTRIBUTE_FROM_GROUP: 'Removed attribute from a team'
};

const mapData = (input: any[] = [], key: string) => {
  return input.reduce((output, row) => {
    if (!Object.prototype.hasOwnProperty.call(output, row[key])) {
      // eslint-disable-next-line no-param-reassign
      output[row[key]] = row;
    }
    return output;
  }, {});
};

const activityResponsePayload = (activity: Activity) => {
  const activityResponse: ActivityType = {
    createdAt: activity.createdAt,
    diff: { created: undefined, edited: undefined, removed: undefined },
    id: activity.id,
    reason: activity.title,
    user: activity.createdBy.username
  };
  return activityResponse;
};

const calcDiff = (input: Record<string, string>[], key: string) => {
  return input.filter((diff) => {
    return diff.path[0] === key;
  });
};

const relationType = (diffs: Record<string, string>[]) => {
  const relation = {
    isRole: false,
    isUser: false
  };
  const pType = calcDiff(diffs, 'ptype');
  let value = '';
  if (pType.length > 0) {
    if (Object.prototype.hasOwnProperty.call(pType[0], 'rhs')) {
      value = pType[0].rhs;
    } else {
      value = pType[0].lhs;
    }

    if (value === 'p') {
      relation.isRole = true;
    }

    if (value === 'g') {
      relation.isUser = true;
    }
  }
  return relation;
};

const parseGroupActivity = async (activity: Activity) => {
  const output = activityResponsePayload(activity);
  const displayName = calcDiff(activity.diffs, 'displayname');
  const metadata = calcDiff(activity.diffs, 'metadata');

  if (activity.documentId === '0') {
    // created
    output.diff.created = [displayName[0].rhs];
    if (metadata.length > 0) {
      Object.keys(metadata[0].rhs).forEach((key) => {
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore
        output.diff[key] = [metadata[0].rhs[key]];
      });
    }
  } else {
    if (metadata.length > 0) {
      metadata.forEach((meta) => {
        output.diff[meta.path[1] || ''] = [meta.lhs, meta.rhs];
      });
    }

    if (
      displayName.length > 0 &&
      Object.prototype.hasOwnProperty.call(displayName[0], 'lhs')
    ) {
      if (Object.prototype.hasOwnProperty.call(displayName[0], 'rhs')) {
        output.diff.edited = [displayName[0].lhs, displayName[0].rhs];
      } else {
        output.diff.removed = [displayName[0].lhs];
      }
    }
  }
  return output;
};

const parseCasbinActivity = async (activity: Activity) => {
  const [groups, roles] = await Promise.all([
    await Group.find(),
    await Role.find()
  ]);
  const groupMap = mapData(groups, 'id');
  const roleMap = mapData(roles, 'id');
  const output = activityResponsePayload(activity);
  const relation = relationType(activity.diffs);
  const isDocumentEmpty = Object.keys(activity.document).length === 0;
  if (isDocumentEmpty) {
    if (relation.isRole) {
      const userDiff: any = calcDiff(activity.diffs, 'subject');
      const groupDiff: any = calcDiff(activity.diffs, 'resource');
      const roleDiff: any = calcDiff(activity.diffs, 'action');
      const role = roleMap[roleDiff[0]?.rhs?.role || ''];
      const group = groupMap[groupDiff[0]?.rhs?.group || ''];
      const user = await User.findOne({
        select: ['displayname'],
        where: {
          id: userDiff[0]?.rhs?.user
        }
      });
      output.diff.created = [
        `Assigned a role ${role?.displayname || ''} ${
          user?.displayname ? `to user ${user?.displayname}` : ''
        } for team ${group?.displayname || ''}`
      ];
    } else if (relation.isUser) {
      const userDiff: any = calcDiff(activity.diffs, 'subject');
      const groupDiff: any = calcDiff(activity.diffs, 'resource');
      const group = groupMap[groupDiff[0]?.rhs?.group || ''];
      const user = await User.findOne({
        select: ['displayname'],
        where: {
          id: userDiff[0]?.rhs?.user
        }
      });
      output.diff.created = [
        `Assigned a user ${user?.displayname || ''} to team ${
          group?.displayname || ''
        }`
      ];
    }
  } else if (relation.isRole) {
    const { role } = activity.document.action;
    const { group } = activity.document.subject;
    output.diff.removed = [
      `Removed a role ${roleMap[role]?.displayname || ''} from team ${
        groupMap[group]?.displayname || ''
      }`,
      ''
    ];
  } else if (relation.isUser) {
    const user = await User.findOne({
      select: ['displayname'],
      where: {
        id: activity.document?.subject?.user
      }
    });
    const { group } = activity.document?.resource;
    output.diff.removed = [
      `Remove a user ${user?.displayname || ''} from team ${
        groupMap[group]?.displayname || ''
      }`,
      ''
    ];
  }
  return output;
};

export const get = async (groupId = '') => {
  let whereClause =
    '( activity.title != :addAttributeToGroup AND activity.title != :removeAttributeFromGroup )';
  const whereParameter = {
    addAttributeToGroup: titleMap.ADD_ATTRIBUTE_TO_GROUP,
    removeAttributeFromGroup: titleMap.REMOVED_ATTRIBUTE_FROM_GROUP,
    createGroup: '',
    userRoleGroupMap: '',
    groupId: ''
  };

  const ActivityRepository = getManager().getRepository(Activity);
  if (groupId) {
    whereClause +=
      ' AND ( activity.diffs @> :createGroup OR activity.diffs @> :userRoleGroupMap ) ';
    whereParameter.createGroup = JSON.stringify([{ rhs: groupId }]);
    whereParameter.userRoleGroupMap = JSON.stringify([
      {
        rhs: { group: groupId }
      }
    ]);
  }

  const activities = await ActivityRepository.createQueryBuilder('activity')
    .where(whereClause, whereParameter)
    .orderBy('activity.created_at', 'DESC')
    .skip(0)
    .take(50)
    .getMany();

  return await Promise.all(
    activities.map(async (activity) => {
      let output: ActivityType = {
        createdAt: '',
        diff: { created: undefined, edited: undefined, removed: undefined },
        id: '',
        reason: '',
        user: ''
      };

      if (activity.model === Constants.MODEL.Group) {
        output = await parseGroupActivity(activity);
      } else if (activity.model === Constants.MODEL.CasbinRule) {
        output = await parseCasbinActivity(activity);
      }
      return output;
    })
  );
};

export const create = async (payload: any) => {
  if (payload?.diffs && payload.diffs.length > 0) {
    return await Activity.save({ ...payload });
  }
  return null;
};

const getTitle = (event: any, type: string) => {
  let title = '';
  switch (type) {
    case actions.CREATE:
      if (event.metadata.tableName === Constants.MODEL.Group) {
        title = `Created ${event.entity?.displayname} team `;
      } else if (event.metadata.tableName === Constants.MODEL.CasbinRule) {
        if (event.entity?.ptype === 'p') {
          title = titleMap.ASSIGNED_ROLE;
        } else if (event.entity?.ptype === 'g') {
          title = titleMap.ASSIGNED_USER;
        } else if (event.entity?.ptype === 'g2') {
          title = titleMap.ADD_ATTRIBUTE_TO_GROUP;
        }
      }
      break;
    case actions.EDIT:
      if (event.metadata.tableName === Constants.MODEL.Group) {
        title = `Edited ${event.entity?.displayname}`;
      } else if (event.metadata.tableName === Constants.MODEL.CasbinRule) {
        title = `Edited ${event.entity?.ptype} Casbin Rule `;
      }
      break;
    case actions.DELETE:
      if (event.metadata.tableName === Constants.MODEL.Group) {
        title = `Deleted ${event.entity?.displayname} team`;
      } else if (event.metadata.tableName === Constants.MODEL.CasbinRule) {
        if (event.databaseEntity?.ptype === 'p') {
          title = titleMap.REMOVED_ROLE;
        } else if (event.databaseEntity?.ptype === 'g') {
          title = titleMap.REMOVED_USER;
        } else if (event.databaseEntity?.ptype === 'g2') {
          title = titleMap.REMOVED_ATTRIBUTE_FROM_GROUP;
        }
      }
      break;
    default:
      title = '';
  }

  return title;
};

export const log = async (event: any, type: string) => {
  let promise = null;
  const title = getTitle(event, type);
  switch (type) {
    case actions.CREATE:
      promise = create({
        document: {},
        title,
        documentId: '0',
        model: event.metadata.tableName,
        diffs: delta({}, event.entity || {}, {
          exclude: excludeFields
        }),
        createdBy: event.queryRunner.data.user
      });
      break;
    case actions.EDIT:
      promise = create({
        document: event.databaseEntity,
        title,
        documentId: event.databaseEntity?.id || '0',
        model: event.metadata.tableName,
        diffs: delta(event.databaseEntity || {}, event.entity || {}, {
          exclude: excludeFields
        }),
        createdBy: event.queryRunner.data.user
      });
      break;
    case actions.DELETE:
      promise = create({
        document: event.databaseEntity,
        title,
        documentId: event.databaseEntity?.id || '0',
        model: event.metadata.tableName,
        diffs: delta(event.databaseEntity || {}, event.entity || {}, {
          exclude: excludeFields
        }),
        createdBy: event.queryRunner.data.user
      });
      break;
    default:
      promise = Promise.resolve();
  }
  await promise;
};
