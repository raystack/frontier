import {
  EntitySubscriberInterface,
  EventSubscriber,
  InsertEvent,
  UpdateEvent
} from 'typeorm';
import {
  log as ActivityLog,
  actions as ActivityActions
} from '../app/activity/resource';
import Constants from '../utils/constant';

@EventSubscriber()
export class ModelSubscriber implements EntitySubscriberInterface {
  afterInsert = async (event: InsertEvent<any>) => {
    if (
      event.metadata.tableName === Constants.MODEL.Activity ||
      event.metadata.tableName === Constants.MODEL.Role ||
      event.metadata.tableName === Constants.MODEL.User ||
      event.metadata.tableName === Constants.MODEL.CasbinRule
    ) {
      return;
    }
    await ActivityLog(event, ActivityActions.CREATE);
  };

  afterUpdate = async (event: UpdateEvent<any>) => {
    if (
      event.metadata.tableName === Constants.MODEL.Activity ||
      event.metadata.tableName === Constants.MODEL.Role ||
      event.metadata.tableName === Constants.MODEL.User ||
      event.metadata.tableName === Constants.MODEL.CasbinRule
    ) {
      return;
    }
    await ActivityLog(event, ActivityActions.EDIT);
  };
}
