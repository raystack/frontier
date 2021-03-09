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
import { Group } from '../model/group';

@EventSubscriber()
export class GroupSubscriber implements EntitySubscriberInterface<Group> {
  listenTo = () => {
    return Group;
  };

  afterInsert = async (event: InsertEvent<any>) => {
    await ActivityLog(event, ActivityActions.CREATE);
  };

  afterUpdate = async (event: UpdateEvent<any>) => {
    await ActivityLog(event, ActivityActions.EDIT);
  };
}
