import Hapi from '@hapi/hapi';
import * as Handler from './handler';

export const plugin = {
  name: 'profile',
  dependencies: 'postgres',
  register(server: Hapi.Server) {
    server.route([
      {
        method: 'GET',
        path: '/api/profile',
        options: Handler.get
      },
      {
        method: 'PUT',
        path: '/api/profile',
        options: Handler.update
      }
    ]);
  }
};
