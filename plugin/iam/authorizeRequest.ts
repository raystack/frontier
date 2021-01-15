import Hapi from '@hapi/hapi';
import Boom from '@hapi/boom';
import CasbinSingleton from '../../lib/casbin';
import { IAMRouteOptionsApp, IAMAuthorizeList, IAMAuthorize } from './types';
import { getIAMAction, constructIAMResourceFromConfig } from './utils';

const authorizeRequest = (server: Hapi.Server) => {
  server.ext({
    type: 'onPreHandler',
    async method(request, h) {
      const { enforcer } = CasbinSingleton;
      const route = server.match(request.method, request.path);
      const { iam } = <IAMRouteOptionsApp>route?.settings?.app || {};

      const createEnforcerPromise = (authorizeConfig: IAMAuthorize) => {
        const { resource: resourceTransformConfig } = authorizeConfig;
        const resource = constructIAMResourceFromConfig(
          resourceTransformConfig,
          request
        );
        const { username } = request.auth.credentials;
        const action = getIAMAction(authorizeConfig.action, request.method);

        return enforcer?.enforceJson({ username }, resource, {
          action
        });
      };

      const checkAuthorization = async (
        authorizeConfigList: IAMAuthorizeList
      ) => {
        const enforcerPromiseList = authorizeConfigList.map(
          createEnforcerPromise
        );

        // user should have access to all authConfigs to get access to an endpoint
        const result = await Promise.all(enforcerPromiseList);
        return result.every((hasAccess) => hasAccess === true);
      };

      if (iam?.authorize) {
        const hasAccess = await checkAuthorization(iam.authorize);
        if (!hasAccess) {
          return Boom.forbidden("Sorry you don't have access");
        }
      }
      return h.continue;
    }
  });
};

export default authorizeRequest;
