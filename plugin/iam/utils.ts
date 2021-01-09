import * as R from 'ramda';
import _ from 'lodash';
import {
  RequestToIAMTransformConfig,
  RequestToIAMTransformObj,
  RequestKeysForIAM
} from './types';

export const constructIAMObjFromRequest = (
  request: Record<RequestKeysForIAM, unknown>,
  requestToIAMTransformConfig: RequestToIAMTransformConfig
) => {
  return _.reduce(
    <Record<RequestKeysForIAM, RequestToIAMTransformObj[]>>(
      requestToIAMTransformConfig
    ),
    (finalResource, transformConfig, key) => {
      const requestData = R.pathOr({}, [key], request);

      return transformConfig?.reduce((transformedObj, config) => {
        const { requestKey, iamKey } = config;
        const valueInRequestData = R.path([requestKey], requestData);

        if (valueInRequestData) {
          return R.assocPath([iamKey], valueInRequestData, transformedObj);
        }
        return transformedObj;
      }, finalResource);
    },
    {}
  );
};
