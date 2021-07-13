import Hapi from '@hapi/hapi';
import * as Handler from './handler';

export const plugin = {
  name: 'activity',
  dependencies: 'postgres',
  register(server: Hapi.Server) {
    server.route([
      {
        method: 'GET',
        path: '/api/activities',
        options: Handler.get
      }
    ]);
  }
};
