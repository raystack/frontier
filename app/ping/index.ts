import Hapi from '@hapi/hapi';
import * as Handler from './handler';

export const plugin = {
  name: 'ping',
  version: '1.0.0',
  dependencies: 'postgres',
  register(server: Hapi.Server) {
    server.route([
      {
        method: 'GET',
        path: '/ping',
        options: Handler.ping
      }
    ]);
  }
};
