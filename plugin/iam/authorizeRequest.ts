import Hapi from '@hapi/hapi';
import * as R from 'ramda';
import Boom from '@hapi/boom';
import CasbinSingleton from '../../lib/casbin';
import { IAMRouteOptionsApp, IAMAuthorizeList, IAMAuthorize } from './types';
import { constructIAMResourceFromConfig } from './utils';

const getRequestData = async (request: Hapi.Request) => {
  const { payload } = <any>request;
  let parsedPayload = payload;

  if (Buffer.isBuffer(payload)) {
    try {
      parsedPayload = JSON.parse(payload.toString());
      // eslint-disable-next-line no-empty
    } catch (e) {}
  }

  const requestData = R.assoc(
    'payload',
    parsedPayload,
    R.pick(['query', 'params', 'headers'], request)
  );
  return requestData;
};

const authorizeRequest = async (
  server: Hapi.Server,
  request: Hapi.Request,
  h: Hapi.ResponseToolkit
) => {
  const { enforcer } = CasbinSingleton;
  const route = server.match(request.method, request.path);
  const { iam } = <IAMRouteOptionsApp>route?.settings?.app || {};

  const requestDataForIam = await getRequestData(request);

  const createEnforcerPromise = (authorizeConfig: IAMAuthorize) => {
    const { attributes: attributesTransformConfig, action } = authorizeConfig;

    const resource = constructIAMResourceFromConfig(
      attributesTransformConfig,
      requestDataForIam
    );
    const { id: userId } = request.auth.credentials;
    return enforcer?.enforceJson({ user: userId }, resource, {
      action
    });
  };

  const checkAuthorization = async (authorizeConfigList: IAMAuthorizeList) => {
    const enforcerPromiseList = authorizeConfigList.map(createEnforcerPromise);

    // user should have access to all authConfigs to get access to an endpoint
    const result = await Promise.all(enforcerPromiseList);
    return result.every((hasAccess) => hasAccess === true);
  };

  if (iam?.permissions) {
    const hasAccess = await checkAuthorization(iam.permissions);
    if (!hasAccess) {
      return Boom.forbidden("Sorry you don't have access");
    }
  }
  return h.continue;
};

export default authorizeRequest;
