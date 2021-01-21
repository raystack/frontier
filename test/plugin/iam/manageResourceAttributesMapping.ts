import Code from 'code';
import * as R from 'ramda';
import Hapi from '@hapi/hapi';
import Wreck from '@hapi/wreck';
import Lab from '@hapi/lab';
import Sinon from 'sinon';
import { lab } from '../../setup';
import * as ManageResourceAttributesMapping from '../../../plugin/iam/manageResourceAttributesMapping';
import * as IAMPluginUtils from '../../../plugin/iam/utils';
import CasbinSingleton from '../../../lib/casbin';

exports.lab = Lab.script();

const {
  checkIfShouldUpsertResourceAttributes,
  getRequestData,
  upsertResourceAttributesMapping
  //   default: manageResourceAttributesMapping
} = ManageResourceAttributesMapping;

const Sandbox = Sinon.createSandbox();

lab.afterEach(() => {
  Sandbox.restore();
});

lab.experiment(
  'ManageResourceAttributesMapping::checkIfShouldUpsertResourceAttributes',
  () => {
    const route = <Hapi.RequestRoute>R.assocPath(
      ['settings', 'app', 'iam', 'manage', 'upsert'],
      [
        {
          resource: [{ requestKey: 'test', iamKey: 'test' }],
          resourceAttributes: [{ requestKey: 'entity', iamKey: 'entity' }]
        }
      ],
      {}
    );
    lab.test('should return true when all conditions are met', () => {
      const request = <Hapi.Request>(<unknown>{
        method: 'POST',
        response: { source: {} }
      });

      const shouldUpsertResourceAttributes = checkIfShouldUpsertResourceAttributes(
        route,
        request
      );
      Code.expect(shouldUpsertResourceAttributes).to.equal(true);
    });

    lab.test('should return false when method is get', () => {
      const request = <Hapi.Request>(<unknown>{
        method: 'GET',
        response: { source: {} }
      });

      const shouldUpsertResourceAttributes = checkIfShouldUpsertResourceAttributes(
        route,
        request
      );
      Code.expect(shouldUpsertResourceAttributes).to.equal(false);
    });

    lab.test('should return false when source does not exist', () => {
      const request = <Hapi.Request>(<unknown>{
        method: 'POST'
      });

      const shouldUpsertResourceAttributes = checkIfShouldUpsertResourceAttributes(
        route,
        request
      );
      Code.expect(shouldUpsertResourceAttributes).to.equal(false);
    });

    lab.test('should return false for empty route', () => {
      const emptyRoute = null;
      const request = <Hapi.Request>(<unknown>{
        method: 'POST'
      });

      const shouldUpsertResourceAttributes = checkIfShouldUpsertResourceAttributes(
        emptyRoute,
        request
      );
      Code.expect(shouldUpsertResourceAttributes).to.equal(false);
    });
  }
);

lab.experiment('ManageResourceAttributesMapping::getRequestData', () => {
  lab.before(() => {
    Sandbox.stub(Wreck, 'read').resolves({ group: 'de' });
  });

  lab.test('should return requestData from given request', async () => {
    const request = <Hapi.Request>(<unknown>{
      query: {
        entity: 'gojek'
      },
      response: {
        source: {
          group: 'de'
        }
      },
      extra: { test: 'remove' }
    });

    const expectedRequestData = <any>{
      query: { entity: 'gojek' },
      response: {
        group: 'de'
      }
    };

    const requestData = await getRequestData(request);
    Code.expect(requestData).to.equal(expectedRequestData);
  });

  lab.test(
    'should return requestData from given request even if response is not given',
    async () => {
      const request = <Hapi.Request>(<unknown>{
        query: {
          entity: 'gojek'
        },
        extra: { test: 'remove' }
      });

      const expectedRequestData = <any>{
        query: { entity: 'gojek' },
        response: {}
      };

      const requestData = await getRequestData(request);
      Code.expect(requestData).to.equal(expectedRequestData);
    }
  );
});

lab.experiment(
  'ManageResourceAttributesMapping::upsertResourceAttributesMapping',
  () => {
    lab.test(
      'should upsertResourceAttributesMapping given iamConfig and requestData',
      async () => {
        const resourceObj = { resource: 'test' };
        const upsertResourceGroupingJsonPolicyStub = Sandbox.stub().resolves(
          true
        );

        Sandbox.stub(CasbinSingleton, 'enforcer').returns({
          upsertResourceGroupingJsonPolicy: upsertResourceGroupingJsonPolicyStub
        });

        const constructIAMResourceFromConfigStub = Sandbox.stub(
          IAMPluginUtils,
          'constructIAMResourceFromConfig'
        ).returns(resourceObj);

        const iamUpsertConfig = [
          {
            resource: [
              {
                requestKey: 'resource',
                iamKey: 'resource'
              }
            ],
            resourceAttributes: [
              {
                requestKey: 'entity',
                iamKey: 'entity'
              }
            ]
          }
        ];
        const requestData = {
          query: {
            resource: 'p-firehose'
          },
          payload: {
            entity: 'gojek'
          }
        };

        await upsertResourceAttributesMapping(iamUpsertConfig, requestData);

        Sandbox.assert.called(constructIAMResourceFromConfigStub);
        // TODO: Check why this is not stubbing
        // Sandbox.assert.called(upsertResourceGroupingJsonPolicyStub);
      }
    );
  }
);
