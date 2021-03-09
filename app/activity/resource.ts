import { Activity } from '../../model/activity';
import Constants from '../../utils/constant';
import { delta } from '../../utils/deep-diff';

export const get = async (team = '') => {
  let criteria: any = {
    order: {
      createdAt: 'DESC'
    }
  };

  if (team.length !== 0) {
    // fetch activities based on team
    criteria = Object.assign(criteria, {
      where: {
        team
      }
    });
  }

  return Activity.find(criteria);
};

export const create = async (payload: any) => {
  return await Activity.save({ ...payload });
};

const excludeFields = ['createdAt', 'updatedAt'];
export const actions = {
  CREATE: 'create',
  EDIT: 'edit',
  DELETE: 'delete'
};

const getTitle = (event: any, type: string) => {
  let title = '';
  switch (type) {
    case actions.CREATE:
      if (event.metadata.tableName === Constants.MODEL.Group) {
        title = `Created ${event.entity?.displayname} Team `;
      } else if (event.metadata.tableName === Constants.MODEL.CasbinRule) {
        title = `Created ${event.entity?.ptype} Casbin Rule `;
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
        title = `Deleted ${event.entity?.displayname} Team`;
      } else if (event.metadata.tableName === Constants.MODEL.CasbinRule) {
        title = `Deleted ${event.databaseEntity?.ptype} Casbin Rule `;
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
        documentId: event.databaseEntity.id,
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
        documentId: event.databaseEntity.id,
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
