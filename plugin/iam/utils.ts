import _ from 'lodash';
import {
  RequestToIAMTransformConfig,
  RequestToIAMTransformObj,
  RequestKeysForIAM
} from '../../app/proxy/types';

export const constructIAMObjFromRequest = (
  request: Record<RequestKeysForIAM, unknown>,
  requestToIAMTransformConfig: RequestToIAMTransformConfig
) => {
  return _.reduce(
    <Record<RequestKeysForIAM, RequestToIAMTransformObj[]>>(
      requestToIAMTransformConfig
    ),
    (finalResource, transformConfig, key) => {
      const requestData = _.get(request, key, {});

      return transformConfig?.reduce((transformedObj, config) => {
        const { requestKey, iamKey } = config;
        const valueInRequestData = requestData[requestKey];

        return {
          ...transformedObj,
          ...(valueInRequestData && { [iamKey]: valueInRequestData })
        };
      }, finalResource);
    },
    {}
  );
};
