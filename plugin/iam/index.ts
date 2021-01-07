import Hapi from '@hapi/hapi';
import Boom from '@hapi/boom';
import Wreck from '@hapi/wreck';
import _ from 'lodash';
import CasbinSingleton from '../../lib/casbin';
import { ConnectionConfig } from '../postgres';
import {
  IAMRouteOptionsApp,
  IAMAuthorizeList,
  IAMAuthorize
} from '../../app/proxy/types';
import { constructIAMObjFromRequest } from './utils';

export const plugin = {
  name: 'iam',
  dependencies: ['postgres', 'iap'],
  async register(server: Hapi.Server, options: ConnectionConfig) {
    const enforcer = await CasbinSingleton.create(options.uri);

    server.ext({
      type: 'onPreHandler',
      method: async function authorizeRequest(request, h) {
        const route = server.match(request.method, request.path);
        const { iam } = <IAMRouteOptionsApp>route?.settings?.app || {};

        const createEnforcerPromise = (authorizeConfig: IAMAuthorize) => {
          const { action, resource: resourceTransformConfig } = authorizeConfig;
          const { username } = request.auth.credentials;
          const resource = constructIAMObjFromRequest(
            request,
            resourceTransformConfig
          );

          return enforcer.enforceJson({ username }, resource, {
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
          if (hasAccess) {
            return Boom.forbidden("Sorry you don't have access");
          }
        }
        return h.continue;
      }
    });

    server.ext({
      type: 'onPreResponse',
      method: async function manageIAMPolicy(request, h) {
        const route = server.match(request.method, request.path);
        const { iam } = <IAMRouteOptionsApp>route?.settings?.app || {};

        // TODO: remove <any> if there is a better approach here. Not able to access response.source without this
        const { response } = <any>request;
        const iamUpsertConfigList = iam?.manage?.upsert;
        const shouldUpsertIAMPolicy = iamUpsertConfigList && response.source;

        if (shouldUpsertIAMPolicy) {
          const iamPolicyUpsertOperationList = _.map(
            iamUpsertConfigList,
            async function upsertIAMConfig(iamUpsertConfig) {
              if (
                iamUpsertConfig.resource &&
                iamUpsertConfig.resourceAttributes
              ) {
                const body = await Wreck.read(response.source, {
                  json: 'force',
                  gunzip: true
                });
                const requestData = _.assign(
                  _.pick(request, ['query', 'params', 'payload']),
                  { response: body }
                );

                const resource = constructIAMObjFromRequest(
                  requestData,
                  iamUpsertConfig.resource
                );
                const resourceAttributes = constructIAMObjFromRequest(
                  requestData,
                  iamUpsertConfig.resourceAttributes
                );

                await enforcer.upsertResourceGroupingJsonPolicy(
                  resource,
                  resourceAttributes
                );
              }
            }
          );

          await Promise.all(iamPolicyUpsertOperationList);
        }
        return h.continue;
      }
    });
  }
};
