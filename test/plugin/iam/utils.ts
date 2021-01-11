import Code from 'code';
import Lab from '@hapi/lab';
import { lab } from '../../setup';
import { constructIAMObjFromRequest } from '../../../plugin/iam/utils';

exports.lab = Lab.script();

lab.experiment('IAM Plugin Utils::constructIAMObjFromRequest', () => {
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
        },
        response: {}
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

      const resource = constructIAMObjFromRequest(
        request,
        resourceTransformConfig
      );

      Code.expect(expectedResult).to.equal(resource);
    }
  );
});
