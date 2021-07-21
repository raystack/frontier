import * as Handler from './handler';
import { IAMRoute } from '../../../plugin/iam/types';

export const USER_GROUP_ROUTES: IAMRoute[] = [
  {
    method: 'GET',
    path: '/api/users/self/groups',
    options: Handler.selfList
  },
  {
    method: 'GET',
    path: '/api/users/{userId}/groups',
    options: Handler.list
  }
];
