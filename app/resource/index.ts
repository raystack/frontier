import Hapi from '@hapi/hapi';
import * as Handler from './handler';

export const plugin = {
  name: 'resource',
  dependencies: 'postgres',
  register(server: Hapi.Server) {
    server.route([
      {
        method: 'POST',
        path: '/api/resources',
        options: Handler.create
      }
    ]);
  }
};
