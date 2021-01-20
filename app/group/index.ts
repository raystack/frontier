import Hapi from '@hapi/hapi';
import * as Handler from './handler';

export const plugin = {
  name: 'group',
  dependencies: 'postgres',
  register(server: Hapi.Server) {
    server.route([
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
      }
    ]);
  }
};
