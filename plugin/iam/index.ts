import Hapi from '@hapi/hapi';
import CasbinSingleton from '../../lib/casbin';
import { ConnectionConfig } from '../postgres';
import manageResourceAttributesMapping from './manageResourceAttributesMapping';
import authorizeRequest from './authorizeRequest';

export const plugin = {
  name: 'iam',
  dependencies: ['postgres', 'iap'],
  async register(server: Hapi.Server, options: ConnectionConfig) {
    await CasbinSingleton.create(options.uri);
    authorizeRequest(server);

    server.ext({
      type: 'onPreResponse',
      async method(request, h) {
        manageResourceAttributesMapping(server, request, h);
      }
    });
  }
};
