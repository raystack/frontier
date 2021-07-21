import Hapi from '@hapi/hapi';
import * as Handler from './handler';
import { IAMRoute } from '../../plugin/iam/types';
import { GROUP_USER_ROUTES } from './user';

export const plugin = {
  name: 'group',
  dependencies: 'postgres',
  register(server: Hapi.Server) {
    const ROUTES: IAMRoute[] = [
      {
        method: 'GET',
        path: '/api/groups',
        options: Handler.list
      },
      {
        method: 'POST',
        path: '/api/groups',
        options: Handler.create
      },
      {
        method: 'GET',
        path: '/api/groups/{id}',
        options: Handler.get
      },
      {
        method: 'PUT',
        path: '/api/groups/{id}',
        options: Handler.update
      },
      {
        method: 'GET',
        path: '/api/groups/{id}/activities',
        options: Handler.getActivitiesByGroup
      }
    ];

    server.route([...ROUTES, ...GROUP_USER_ROUTES]);
  }
};
