import * as Handler from './handler';
import { IAMRoute } from '../../../plugin/iam/types';

export const GROUP_USER_ROUTES: IAMRoute[] = [
  {
    method: ['GET', 'POST', 'PUT', 'DELETE'],
    path: '/api/groups/{groupId}/users/{userId}',
    options: Handler.operation
  }
];
