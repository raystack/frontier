import Code from 'code';
import Lab from '@hapi/lab';
import { constructResource } from '../../../plugin/iam/utils';

const lab = Lab.script();
exports.lab = lab;

lab.experiment('IAM Plugin Utils::constructResource', () => {
  lab.test(
    'should construct the resource object for the given request and transformConfig',
    async () => {
      const request = {
        query: {
          entity: 'platform'
        },
        params: {
          name: 'p-firehose-123'
        },
        payload: {
          environment: 'integration'
        }
      };

      const resourceTransformConfig = {
        query: [
          {
            requestKey: 'entity',
            iamKey: 'newEntity'
          }
        ],
        params: [
          {
            requestKey: 'name',
            iamKey: 'resource'
          }
        ],
        payload: [
          {
            requestKey: 'environment',
            iamKey: 'environment'
          }
        ]
      };
      const expectedResult = {
        newEntity: 'platform',
        resource: 'p-firehose-123',
        environment: 'integration'
      };

      const resource = constructResource(request, resourceTransformConfig);

      Code.expect(expectedResult).to.equal(resource);
    }
  );
});
