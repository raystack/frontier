import Hapi from '@hapi/hapi';
import * as Handler from './handler';

export const plugin = {
  name: 'role',
  dependencies: 'postgres',
  register(server: Hapi.Server) {
    server.route([
      {
        method: 'GET',
        path: '/api/roles',
        options: Handler.get
      },
      {
        method: 'POST',
        path: '/api/roles',
        options: Handler.create
      },
      {
        method: 'PUT',
        path: '/api/roles/{id}',
        options: Handler.update
      }
    ]);
  }
};
