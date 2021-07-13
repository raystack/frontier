import Code from 'code';
import Lab from '@hapi/lab';
import { lab } from '../../setup';
import * as IAMPluginUtils from '../../../src/plugin/iam/utils';

exports.lab = Lab.script();

const { constructIAMResourceFromConfig } = IAMPluginUtils;

lab.experiment('IAMPluginUtils::constructIAMResourceFromConfig', () => {
  lab.test(
    'should construct the resource object for the given transformConfig and request',
    async () => {
      const request = {
        query: {
          entity: 'platform',
          landscape: 'id'
        },
        params: {
          name: 'p-firehose-123'
        },
        payload: {
          environment: 'integration'
        }
      };

      const resourceTransformConfig = [
        {
          newEntity: {
            key: 'entity',
            type: 'query'
          }
        },
        {
          resource: {
            key: 'name',
            type: 'params'
          }
        },
        {
          environment: {
            key: 'environment',
            type: 'payload'
          }
        }
      ];

      const expectedResult = {
        newEntity: 'platform',
        resource: 'p-firehose-123',
        environment: 'integration'
      };

      const resource = constructIAMResourceFromConfig(
        resourceTransformConfig,
        request
      );

      Code.expect(resource).to.equal(expectedResult);
    }
  );

  lab.test(
    'should skip keys that are not present in request data',
    async () => {
      const request = {
        query: {
          entity: 'platform'
        },
        params: {
          name: 'p-firehose-123'
        }
      };

      const resourceTransformConfig = [
        {
          newEntity: {
            key: 'entity',
            type: 'query'
          }
        },
        {
          resource: {
            key: 'name',
            type: 'params'
          }
        },
        {
          environment: {
            key: 'environment',
            type: 'payload'
          }
        }
      ];

      const expectedResult = {
        newEntity: 'platform',
        resource: 'p-firehose-123'
      };

      const resource = constructIAMResourceFromConfig(
        resourceTransformConfig,
        request
      );

      Code.expect(resource).to.equal(expectedResult);
    }
  );
});
