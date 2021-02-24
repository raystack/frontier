import * as R from 'ramda';
import { RequestToIAMTransformConfig, RequestKeysForIAM } from './types';

export type RequestTransformConfig = {
  key: string;
  type: string;
};

export const constructIAMResourceFromConfig = (
  requestToIAMTransformObj: RequestToIAMTransformConfig[] = [],
  request: Partial<Record<RequestKeysForIAM, unknown>>
) => {
  // ? This will contain headers, query, params, payload, and response
  return requestToIAMTransformObj.reduce((resource, transformConfig: any) => {
    const [iamKey, requestObj]: any = Object.entries(transformConfig)[0];
    const requestKey: string = requestObj.key;
    const requestType: string = requestObj.type;

    const valueInRequest = R.pathOr(
      undefined,
      [requestType, requestKey],
      request
    );
    const resourceKey = iamKey || requestObj?.key;

    if (valueInRequest) {
      return R.assocPath([resourceKey], valueInRequest, resource);
    }

    return resource;
  }, {});
};
