import _ from 'lodash';
import {
  ResourceTransformConfig,
  RequestToIAMTransformConfig
} from '../../app/proxy/routes';

export const constructResource = (
  request: Record<keyof ResourceTransformConfig, unknown>,
  resourceTransformConfig: ResourceTransformConfig
) => {
  return _.reduce(
    <Record<'string', RequestToIAMTransformConfig[]>>resourceTransformConfig,
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
