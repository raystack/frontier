import * as R from 'ramda';
import {
  RequestToIAMTransformConfig,
  RequestKeysForIAM,
  IAMAuthorizeAction,
  IAMAuthorizeActionConfig
} from './types';

export const constructIAMResourceFromConfig = (
  requestToIAMTransformObj: RequestToIAMTransformConfig[],
  request: Partial<Record<RequestKeysForIAM, unknown>>
) => {
  // ? This will contain query, params, payload, and response
  const httpRequestKeys = Object.keys(request);

  return requestToIAMTransformObj.reduce(
    (resource, transformConfig: RequestToIAMTransformConfig) => {
      const { requestKey, iamKey } = transformConfig;

      const valueInRequest = httpRequestKeys.reduce(
        (val, httpKey) => R.pathOr(val, [httpKey, requestKey], request),
        undefined
      );

      if (valueInRequest) {
        return R.assocPath([iamKey], valueInRequest, resource);
      }

      return resource;
    },
    {}
  );
};

const getIAMActionOperationByMethod = (method: string) =>
  R.propOr('get', method.toLowerCase(), {
    get: 'read',
    post: 'create',
    put: 'manage',
    patch: 'manage',
    delete: 'delete'
  });

export const contructIAMActionFromConfig = (
  actionConfig: IAMAuthorizeActionConfig,
  method: string
) => {
  if (R.has('operation', actionConfig)) {
    return `${actionConfig.baseName.toLowerCase()}.${actionConfig.operation?.toLowerCase()}`;
  }

  return `${actionConfig.baseName.toLowerCase()}.${getIAMActionOperationByMethod(
    method
  )}`;
};

export const getIAMAction = (action: IAMAuthorizeAction, method: string) => {
  if (typeof action === 'string') return action;

  return contructIAMActionFromConfig(action, method);
};
