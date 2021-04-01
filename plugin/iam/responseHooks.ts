import Hapi from '@hapi/hapi';
import Wreck from '@hapi/wreck';
import * as R from 'ramda';
import { getResourceAttributeMappingsByResources } from '../../app/policy/resource';
import CasbinSingleton from '../../lib/casbin';
import Logger from '../../lib/logger';
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

export const mergeResourceListWithAttributes = async (
  resourceList: any,
  hook: IAMUpsertConfig
) => {
  const {
    resources: rawResourceConfig = [],
    attributes: attributesConfig = []
  } = hook;

  const resourceConfig = rawResourceConfig.map((config) =>
    R.assocPath([...R.keys(config), 'type'], 'response', config)
  );

  // query db to fetch all attributes of the given resource list
  const iamResourceList = resourceList.map((res: any) =>
    constructIAMResourceFromConfig(resourceConfig, { response: res })
  );
  const resouceAttributeDBResults = await getResourceAttributeMappingsByResources(
    iamResourceList
  );

  // Map sample data => { "{urn: "resource"}": [{attribute1: '1', attribute2: '2'}, {attribute3: '3'}] }
  const iamResourceToAttributesMap = resouceAttributeDBResults.reduce(
    (raMap: any, { resource, attributes }: any) => {
      const resourceStrKey = JSON.stringify(resource);

      // eslint-disable-next-line no-param-reassign
      raMap[resourceStrKey] = raMap[resourceStrKey] || [];

      // eslint-disable-next-line no-param-reassign
      raMap[resourceStrKey].push(attributes);
      return raMap;
    },
    {}
  );

  // merge it with resourceList by using hook config
  const mergedResourceList = resourceList.map((res: any) => {
    const iamResource = constructIAMResourceFromConfig(resourceConfig, {
      response: res
    });
    const attributeMappingsForResource =
      iamResourceToAttributesMap[JSON.stringify(iamResource)] || [];

    return attributesConfig.reduce((mergedResource, iamAttributeConfig) => {
      const [[iamAttributeKey, { key }]] = Object.entries(iamAttributeConfig);

      const val = attributeMappingsForResource.reduce(
        (attrVal: any, attributes: any) => {
          if (!R.isEmpty(attrVal) && !R.isNil(attrVal)) return attrVal;

          return R.pathOr('', [iamAttributeKey], attributes);
        },
        ''
      );

      if (!val) return mergedResource;
      return R.assoc(key, val, mergedResource);
    }, res);
  });

  return mergedResourceList;
};

const mergeResponseWithAttributes = async (
  response: any,
  hooks: IAMUpsertConfig[]
) => {
  const firstHook = R.head(hooks) || {};
  let responseList = response;
  const isResponseList = R.is(Array, response);
  if (!isResponseList) {
    responseList = [response];
  }

  const mergedResponseList = await mergeResourceListWithAttributes(
    responseList,
    <IAMUpsertConfig>firstHook
  );

  return isResponseList ? mergedResponseList : R.head(mergedResponseList);
};

export const checkIfShouldTriggerHooks = (
  route: Hapi.RequestRoute | null,
  request: Hapi.Request
) => {
  const { iam } = <IAMRouteOptionsApp>route?.settings?.app || {};

  // TODO: remove <any> if there is a better approach here. Not able to access response.source without this
  const { response } = <any>request || {};
  const statusCode = R.pathOr(100, ['response', 'statusCode'], request);
  const isSuccessStatusCode = statusCode >= 200 && statusCode <= 299;
  const hasUpsertConfig = !R.isEmpty(iam?.hooks) && !R.isNil(iam?.hooks);
  const shouldUpsertResourceAttributes =
    hasUpsertConfig && response?.source && isSuccessStatusCode;

  return !!shouldUpsertResourceAttributes;
};

export const getRequestData = async (request: Hapi.Request) => {
  const { response } = <any>request;
  let body = response?.source;

  if (typeof body === 'object') {
    try {
      body = await Wreck.read(response?.source, {
        json: 'force',
        gunzip: true
      });
      // eslint-disable-next-line no-empty
    } catch (e) {
      Logger.error(`Failed to parse response: ${e}`);
    }
  }

  const requestData = R.assoc(
    'response',
    body || {},
    R.pick(['query', 'params', 'payload', 'headers'], request)
  );
  return requestData;
};

const isWriteMethod = (method = '') => {
  return ['put', 'patch', 'post'].includes(method.toLowerCase());
};

const manageResourceAttributesMapping = async (
  server: Hapi.Server,
  request: Hapi.Request,
  h: Hapi.ResponseToolkit
) => {
  const route = server.match(request.method, request.path);
  const user = request.auth.credentials;
  const shouldTriggerHooks = checkIfShouldTriggerHooks(route, request);
  if (shouldTriggerHooks) {
    const statusCode = R.pathOr(200, ['response', 'statusCode'], request);
    const { iam } = <IAMRouteOptionsApp>route?.settings?.app || {};
    const hooks = <IAMUpsertConfig[]>iam?.hooks || [];

    const requestData = await getRequestData(request);

    // only upsert for write based http methods
    if (isWriteMethod(request.method)) {
      await upsertResourceAttributesMapping(hooks, requestData, user);
    }

    if (!R.isEmpty(requestData.response)) {
      const mergedResponse = await mergeResponseWithAttributes(
        requestData.response,
        hooks
      );
      return h.response(mergedResponse).code(statusCode);
    }
  }
  return h.continue;
};

export default manageResourceAttributesMapping;
