import Hapi from '@hapi/hapi';
import * as Handler from './handler';
import { USER_GROUP_ROUTES } from './group';

export const plugin = {
  name: 'user',
  dependencies: 'postgres',
  register(server: Hapi.Server) {
    const USER_ROUTES = [
      {
        method: 'POST',
        path: '/api/users',
        options: Handler.create
      },
      {
        method: 'GET',
        path: '/api/users',
        options: Handler.list
      },
      {
        method: 'GET',
        path: '/api/users/{userId}',
        options: Handler.get
      }
    ];

    server.route([...USER_ROUTES, ...USER_GROUP_ROUTES]);
  }
};
