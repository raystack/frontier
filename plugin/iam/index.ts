import Hapi from '@hapi/hapi';
import CasbinSingleton from '../../lib/casbin';
import { ConnectionConfig } from '../postgres';
import manageResourceAttributesMapping from './responseHooks';
import authorizeRequest from './authorizeRequest';

export const plugin = {
  name: 'iam',
  dependencies: ['postgres', 'iap'],
  async register(server: Hapi.Server, options: ConnectionConfig) {
    await CasbinSingleton.create(options.uri);

    server.ext({
      type: 'onPreHandler',
      async method(request, h) {
        return authorizeRequest(server, request, h);
      }
    });

    server.ext({
      type: 'onPreResponse',
      async method(request, h) {
        return manageResourceAttributesMapping(server, request, h);
      }
    });
  }
};
