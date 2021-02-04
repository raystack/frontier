import * as Handler from './handler';
import { IAMRoute } from '../../../plugin/iam/types';

export const GROUP_USER_ROUTES: IAMRoute[] = [
  {
    method: ['POST', 'PUT', 'DELETE'],
    path: '/api/groups/{groupIdentifier}/users/{userIdentifier}',
    options: Handler.operation
  }
];
