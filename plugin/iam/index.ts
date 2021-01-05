import Hapi from '@hapi/hapi';
import Boom from '@hapi/boom';
import CasbinSingleton from '../../lib/casbin';
import { ConnectionConfig } from '../postgres';
import { IAMConfig } from '../../app/proxy/routes';
import { constructResource } from './utils';

export const plugin = {
  name: 'iam',
  dependencies: ['postgres', 'iap'],
  async register(server: Hapi.Server, options: ConnectionConfig) {
    const enforcer = await CasbinSingleton.create(options.uri);

    server.ext({
      type: 'onPreHandler',
      async method(request, h) {
        const route = server.match(request.method, request.path);
        const { iam } = <IAMConfig>route?.settings?.app;

        const checkAuthorization = async () => {
          const { action, resourceTransformConfig } = iam;
          const { username } = request.auth.credentials;
          const resource = constructResource(request, resourceTransformConfig);
          return enforcer.enforceJson({ username }, resource, {
            action
          });
        };

        if (iam) {
          const hasAccess = await checkAuthorization();
          if (!hasAccess) {
            return Boom.forbidden("Sorry you don't have access");
          }
        }
        return h.continue;
      }
    });
  }
};
