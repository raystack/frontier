import {
  EntitySubscriberInterface,
  EventSubscriber,
  InsertEvent,
  RemoveEvent,
  UpdateEvent
} from 'typeorm';
import { delta } from '../utils/deep-diff';
import Constants from '../utils/constant';
import { create } from '../app/activity/resource';
import { getPolicies, setPolicies } from '../app/policy/resource';

interface Policy {
  [key: string]: any;
}

const excludeFields = ['createdAt', 'updatedAt'];
const actions = {
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

const storeActivityPayload = async (event: any, type: string) => {
  if (
    event.metadata.tableName === Constants.MODEL.Activity ||
    event.metadata.tableName === Constants.MODEL.Role ||
    event.metadata.tableName === Constants.MODEL.User
  ) {
    return;
  }
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

@EventSubscriber()
export class ModelSubscriber implements EntitySubscriberInterface {
  afterInsert = async (event: InsertEvent<any>) => {
    await storeActivityPayload(event, actions.CREATE);
  };

  afterUpdate = async (event: UpdateEvent<any>) => {
    await storeActivityPayload(event, actions.EDIT);
  };

  beforeRemove = async (event: RemoveEvent<any>) => {
    const policies = await event.queryRunner.query('select * from casbin_rule');
    setPolicies(policies);
  };

  afterRemove = async (event: RemoveEvent<any>) => {
    const previousPolicies = getPolicies();
    const currentPolicies = await event.queryRunner.query(
      'select * from casbin_rule'
    );
    const currentPoliciesMap: Policy = {};
    currentPolicies.forEach((policy: any) => {
      if (
        !Object.prototype.hasOwnProperty.call(currentPoliciesMap, policy.id)
      ) {
        currentPoliciesMap[policy.id] = policy;
      }
    });
    await Promise.all(
      previousPolicies
        .filter((policy: any) => {
          return !Object.prototype.hasOwnProperty.call(
            currentPoliciesMap,
            policy.id
          );
        })
        .map((policy: any) => {
          // eslint-disable-next-line no-param-reassign
          event.databaseEntity = policy;
          return storeActivityPayload(event, actions.DELETE);
        })
    );
  };
}
