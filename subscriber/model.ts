import {
  EntitySubscriberInterface,
  EventSubscriber,
  InsertEvent,
  UpdateEvent
} from 'typeorm';
import { delta } from '../utils/deep-diff';
import Constants from '../utils/constant';
import { create } from '../app/activity/resource';

const excludeFields = ['id', 'createdAt', 'updatedAt'];
const actions = {
  CREATE: 'create',
  EDIT: 'edit'
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
    default:
      title = '';
  }

  return title;
};

const getDiff = (event: any, type: string) => {
  switch (type) {
    case actions.CREATE:
      return delta({}, event.entity, {
        exclude: excludeFields
      });
    case actions.EDIT:
      return delta(event.databaseEntity, event.entity, {
        exclude: excludeFields
      });
    default:
      return [];
  }
};

const storeActivityPayload = async (event: any, type: string) => {
  // console.log(
  //   'storeActivityPayload event -> ',
  //   event.entity,
  //   event.databaseEntity,
  //   event.metadata.tableName
  // );
  if (
    event.metadata.tableName === Constants.MODEL.Activity ||
    event.metadata.tableName === Constants.MODEL.Role ||
    event.metadata.tableName === Constants.MODEL.User
  ) {
    return;
  }
  const title = getTitle(event, type);
  await create({
    document: event.entity,
    title,
    documentId: event.entity.id,
    model: event.metadata.tableName,
    diffs: getDiff(event, type)
  });
};

@EventSubscriber()
export class ModelSubscriber implements EntitySubscriberInterface {
  afterInsert = async (event: InsertEvent<any>) => {
    await storeActivityPayload(event, actions.CREATE);
  };

  afterUpdate = async (event: UpdateEvent<any>) => {
    await storeActivityPayload(event, actions.EDIT);
  };

  /**
   * Called before entity removal.
   */
  // beforeRemove = async (event: RemoveEvent<any>) => {
  // console.log(
  //   `BEFORE ENTITY WITH metadata.tableName ${JSON.stringify(
  //     event.metadata.tableName
  //   )} REMOVED: `,
  //   event.entity
  // );
  // };

  /**
   * Called after entity removal.
   */
  // afterRemove = async (event: RemoveEvent<any>) => {
  // Object.keys(event.queryRunner.data).forEach((key) => {
  //   console.log(`key => ${key} and value => ${event.queryRunner.data[key]}`);
  // });
  //
  // console.log(`AFTER ENTITY WITH queryRunner REMOVED: `, event.entity);
  // };
}
