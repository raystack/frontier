import Hapi from '@hapi/hapi';
import Wreck from '@hapi/wreck';
import * as R from 'ramda';
import CasbinSingleton from '../../lib/casbin';
import { IAMRouteOptionsApp, IAMUpsertConfig } from './types';
import { constructIAMResourceFromConfig } from './utils';

export const upsertResourceAttributesMapping = async (
  iamUpsertConfigList: IAMUpsertConfig[],
  requestData: Record<string, unknown>,
  user: Hapi.UserCredentials | undefined
) => {
  const { enforcer } = CasbinSingleton;

  const iamPolicyUpsertOperationList = iamUpsertConfigList?.map(
    async (iamUpsertConfig: IAMUpsertConfig) => {
      const resource = constructIAMResourceFromConfig(
        iamUpsertConfig.resources,
        requestData
      );
      const resourceAttributes = constructIAMResourceFromConfig(
        iamUpsertConfig.attributes,
        requestData
      );
      if (!R.isEmpty(resource) && !R.isEmpty(resourceAttributes)) {
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore
        return enforcer?.upsertResourceGroupingJsonPolicy(
          resource,
          resourceAttributes,
          { created_by: JSON.parse(JSON.stringify(user)) }
        );
      }

      return Promise.resolve();
    }
  );

  await Promise.all(iamPolicyUpsertOperationList);
};

export const checkIfShouldUpsertResourceAttributes = (
  route: Hapi.RequestRoute | null,
  request: Hapi.Request
) => {
  const { iam } = <IAMRouteOptionsApp>route?.settings?.app || {};

  // TODO: remove <any> if there is a better approach here. Not able to access response.source without this
  const { response } = <any>request || {};
  const ALLOWED_METHODS = ['post', 'put', 'patch'];
  const hasUpsertConfig = iam?.hooks;
  const shouldUpsertResourceAttributes =
    hasUpsertConfig &&
    response?.source &&
    ALLOWED_METHODS.includes(request.method.toLowerCase());

  return !!shouldUpsertResourceAttributes;
};

export const getRequestData = async (request: Hapi.Request) => {
  const { response } = <any>request;
  let body = {};

  if (response?.source) {
    if (typeof response?.source === 'object') {
      body = response?.source;
    } else {
      body = await Wreck.read(response?.source, {
        json: 'force',
        gunzip: true
      });
    }
  }

  const requestData = R.assoc(
    'response',
    body,
    R.pick(['query', 'params', 'payload'], request)
  );
  return requestData;
};

const manageResourceAttributesMapping = async (
  server: Hapi.Server,
  request: Hapi.Request,
  h: Hapi.ResponseToolkit
) => {
  const route = server.match(request.method, request.path);
  const { user } = request.auth.credentials;
  const shouldUpsertResourceAttributes = checkIfShouldUpsertResourceAttributes(
    route,
    request
  );
  if (shouldUpsertResourceAttributes) {
    const { iam } = <IAMRouteOptionsApp>route?.settings?.app || {};
    const iamUpsertConfigList = <IAMUpsertConfig[]>iam?.hooks || [];

    const requestData = await getRequestData(request);

    await upsertResourceAttributesMapping(
      iamUpsertConfigList,
      requestData,
      user
    );
    if (!R.isEmpty(requestData.response)) {
      return h.response(requestData.response);
    }
  }
  return h.continue;
};

export default manageResourceAttributesMapping;
