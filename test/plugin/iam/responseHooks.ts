import Code from 'code';
import * as R from 'ramda';
import Hapi from '@hapi/hapi';
import Wreck from '@hapi/wreck';
import Lab from '@hapi/lab';
import Sinon from 'sinon';
import { lab } from '../../setup';
import * as ResponseHooks from '../../../plugin/iam/responseHooks';
import * as IAMPluginUtils from '../../../plugin/iam/utils';
import CasbinSingleton from '../../../lib/casbin';
import * as PolicyResource from '../../../app/policy/resource';

exports.lab = Lab.script();

const {
  checkIfShouldTriggerHooks,
  getRequestData,
  upsertResourceAttributesMapping,
  mergeResourceListWithAttributes
} = ResponseHooks;

const Sandbox = Sinon.createSandbox();

lab.afterEach(() => {
  Sandbox.restore();
});

lab.experiment('ResponseHooks::checkIfShouldTriggerHooks', () => {
  const route = <Hapi.RequestRoute>R.assocPath(
    ['settings', 'app', 'iam', 'hooks'],
    [
      {
        resources: [
          {
            test: {
              key: 'test',
              type: 'params'
            }
          }
        ],
        attributes: [
          {
            entity: {
              key: 'entity',
              type: 'payload'
            }
          }
        ]
      }
    ],
    {}
  );
  lab.test('should return true when all conditions are met', () => {
    const request = <Hapi.Request>(<unknown>{
      method: 'POST',
      response: { source: {} }
    });

    const shouldTriggerHooks = checkIfShouldTriggerHooks(route, request);
    Code.expect(shouldTriggerHooks).to.equal(true);
  });

  lab.test('should return false when source does not exist', () => {
    const request = <Hapi.Request>(<unknown>{
      method: 'POST'
    });

    const shouldTriggerHooks = checkIfShouldTriggerHooks(route, request);
    Code.expect(shouldTriggerHooks).to.equal(false);
  });

  lab.test('should return false for empty route', () => {
    const emptyRoute = null;
    const request = <Hapi.Request>(<unknown>{
      method: 'POST'
    });

    const shouldTriggerHooks = checkIfShouldTriggerHooks(emptyRoute, request);
    Code.expect(shouldTriggerHooks).to.equal(false);
  });
});

lab.experiment('ResponseHooks::getRequestData', () => {
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

lab.experiment('ResponseHooks::upsertResourceAttributesMapping', () => {
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
          resources: [
            {
              resource: {
                key: 'resource',
                type: 'query'
              }
            }
          ],
          attributes: [
            {
              entity: {
                key: 'entity',
                type: 'payload'
              }
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
});

lab.experiment('ResponseHooks::mergeResourceListWithAttributes', () => {
  let getResourceAttributeMappingsByResourcesStub;
  const hook = {
    resources: [{ name: { key: 'name', type: 'response' } }],
    attributes: [
      { entity: { key: 'entity', type: 'response' } },
      { environment: { key: 'env', type: 'response' } }
    ]
  };

  const resourcesWithAttributes = [
    {
      resource: { name: 'test-resource' },
      attributes: { entity: 'gojek', environment: 'production' }
    },
    {
      resource: { name: 'test-resource-1' },
      attributes: { entity: 'gofing', environment: 'integration' }
    }
  ];

  lab.beforeEach(async () => {
    getResourceAttributeMappingsByResourcesStub = Sandbox.stub(
      PolicyResource,
      'getResourceAttributeMappingsByResources'
    ).resolves(resourcesWithAttributes);
  });

  lab.afterEach(() => {
    Sandbox.restore();
  });

  lab.test('should merge for list of resource', async () => {
    const resourceList = [
      {
        name: 'test-resource',
        urn: '1314143'
      }
    ];
    const mergedResourceList = await mergeResourceListWithAttributes(
      resourceList,
      hook
    );

    const expectedResult = [
      {
        name: 'test-resource',
        urn: '1314143',
        entity: 'gojek',
        env: 'production'
      }
    ];
    Code.expect(mergedResourceList).to.equal(expectedResult);

    Sandbox.assert.calledWithExactly(
      getResourceAttributeMappingsByResourcesStub,
      [{ name: 'test-resource' }]
    );
  });

  lab.test('should return even if attributes are not found', async () => {
    const resourceList = [
      {
        name: 'test-resource-314134',
        urn: '1314143'
      }
    ];
    const mergedResourceList = await mergeResourceListWithAttributes(
      resourceList,
      hook
    );

    Code.expect(resourceList).to.equal(mergedResourceList);

    Sandbox.assert.calledWithExactly(
      getResourceAttributeMappingsByResourcesStub,
      [{ name: 'test-resource-314134' }]
    );
  });
});
