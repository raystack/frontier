import Hapi from '@hapi/hapi';
import * as Handler from './handler';
import { IAMRoute } from '../../plugin/iam/types';

export const plugin = {
  name: 'access',
  dependencies: 'postgres',
  register(server: Hapi.Server) {
    const ROUTES: IAMRoute[] = [
      {
        method: 'POST',
        path: '/api/check-access',
        options: Handler.checkAccess
      }
    ];

    server.route(ROUTES);
  }
};
