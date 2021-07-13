import * as Handler from './handler';
import { IAMRoute } from '../../../plugin/iam/types';

export const GROUP_USER_ROUTES: IAMRoute[] = [
  {
    method: 'GET',
    path: '/api/groups/{groupId}/users',
    options: Handler.list
  },
  {
    method: 'GET',
    path: '/api/groups/{groupId}/users/{userId}',
    options: Handler.get
  },
  {
    method: 'POST',
    path: '/api/groups/{groupId}/users/{userId}',
    options: Handler.post
  },
  {
    method: 'PUT',
    path: '/api/groups/{groupId}/users/{userId}',
    options: Handler.put
  },
  {
    method: 'DELETE',
    path: '/api/groups/{groupId}/users/self',
    options: Handler.removeSelf
  },
  {
    method: 'DELETE',
    path: '/api/groups/{groupId}/users/{userId}',
    options: Handler.remove
  }
];
