import Lab from '@hapi/lab';
import * as R from 'ramda';
import Sinon from 'sinon';
import Code from 'code';
import Faker from 'faker';
import { factory } from 'typeorm-seeding';
import { lab } from '../../setup';
import { Group } from '../../../model/group';
import * as Resource from '../../../app/group/resource';
import * as PolicyResource from '../../../app/policy/resource';

exports.lab = Lab.script();
const Sandbox = Sinon.createSandbox();

lab.afterEach(() => {
  Sandbox.restore();
});

lab.experiment('Group::resource', () => {
  lab.experiment('get group by id', () => {
    let group;

    lab.beforeEach(async () => {
      group = await factory(Group)().create();
    });

    lab.afterEach(() => {
      Sandbox.restore();
    });

    lab.test(
      'should get group by id along with attributes and policies',
      async () => {
        const policies = <any>[{ policy: '' }];
        const getPoliciesBySubjectStub = Sandbox.stub(
          PolicyResource,
          'getPoliciesBySubject'
        ).returns(policies);

        const attributes = <any>[{ entity: 'gojek' }];
        const getAttributesForGroupStub = Sandbox.stub(
          PolicyResource,
          'getAttributesForGroup'
        ).returns(attributes);

        const response = await Resource.get(group.id, {});
        Code.expect(response).to.equal({ ...group, policies, attributes });
        Sandbox.assert.calledWithExactly(
          getPoliciesBySubjectStub,
          { group: group.id },
          {}
        );
        Sandbox.assert.calledWithExactly(getAttributesForGroupStub, group.id);
      }
    );

    lab.test(
      'should return undefined response if group is not found',
      async () => {
        try {
          await Resource.get(Faker.random.uuid());
        } catch (e) {
          Code.expect(e.output.statusCode).to.equal(404);
        }
      }
    );
  });

  lab.experiment('create group', () => {
    lab.afterEach(() => {
      Sandbox.restore();
    });

    lab.test('should create group by id', async () => {
      const payload = <any>{
        displayName: 'Data Engineering',
        metadata: {
          name: 'Data Engineering',
          privacy: 'private'
        },
        policies: [{ operation: 'create' }],
        attributes: [{ entity: 'gojek' }]
      };
      const groupId = Faker.random.uuid();

      const checkSubjectHasAccessToCreateAttributesMappingStub = Sandbox.stub(
        Resource,
        'checkSubjectHasAccessToCreateAttributesMapping'
      ).returns(<any>true);

      const upsertGroupAndAttributesMappingStub = Sandbox.stub(
        Resource,
        'upsertGroupAndAttributesMapping'
      ).returns(<any>true);

      const groupSaveStub = Sandbox.stub(Group, 'save').returns(<any>{
        id: groupId,
        ...payload
      });

      const bulkUpsertPoliciesForGroupStub = Sandbox.stub(
        Resource,
        'bulkUpsertPoliciesForGroup'
      ).returns(<any>[
        {
          operation: 'create',
          success: true
        }
      ]);

      const getStub = Sandbox.stub(Resource, 'get').returns(<any>{
        id: groupId,
        ...payload
      });

      const loggedInUserId = Faker.random.uuid();
      const response = await Resource.create(payload, loggedInUserId);

      Sandbox.assert.calledWithExactly(
        checkSubjectHasAccessToCreateAttributesMappingStub,
        { user: loggedInUserId },
        payload.attributes
      );
      Sandbox.assert.calledWithExactly(
        upsertGroupAndAttributesMappingStub,
        groupId,
        payload.attributes
      );
      Sandbox.assert.calledWithExactly(
        groupSaveStub,
        <any>R.omit(['attributes', 'policies'], payload)
      );
      Sandbox.assert.calledWithExactly(
        bulkUpsertPoliciesForGroupStub,
        groupId,
        payload.policies,
        loggedInUserId
      );
      Sandbox.assert.calledWithExactly(getStub, groupId);
      Code.expect(response).to.equal({
        id: groupId,
        ...payload,
        policyOperationResult: [{ operation: 'create', success: true }]
      });
    });
  });

  lab.experiment('update group by id', () => {
    lab.afterEach(() => {
      Sandbox.restore();
    });

    lab.test('should create group by id', async () => {
      const payload = <any>{
        displayName: 'Data Engineering',
        metadata: {
          name: 'Data Engineering',
          privacy: 'private'
        },
        policies: [{ operation: 'create' }],
        attributes: [{ entity: 'gojek' }]
      };
      const groupId = Faker.random.uuid();

      const checkSubjectHasAccessToCreateAttributesMappingStub = Sandbox.stub(
        Resource,
        'checkSubjectHasAccessToCreateAttributesMapping'
      ).returns(<any>true);

      const upsertGroupAndAttributesMappingStub = Sandbox.stub(
        Resource,
        'upsertGroupAndAttributesMapping'
      ).returns(<any>true);

      const groupSaveStub = Sandbox.stub(Group, 'save').returns(<any>{
        id: groupId,
        ...payload
      });

      const bulkUpsertPoliciesForGroupStub = Sandbox.stub(
        Resource,
        'bulkUpsertPoliciesForGroup'
      ).returns(<any>[
        {
          operation: 'create',
          success: true
        }
      ]);

      const getStub = Sandbox.stub(Resource, 'get').returns(<any>{
        id: groupId,
        ...payload
      });

      const loggedInUserId = Faker.random.uuid();
      const response = await Resource.update(groupId, payload, loggedInUserId);

      Sandbox.assert.calledWithExactly(
        checkSubjectHasAccessToCreateAttributesMappingStub,
        { user: loggedInUserId },
        [...payload.attributes, ...payload.attributes]
      );
      Sandbox.assert.calledWithExactly(
        upsertGroupAndAttributesMappingStub,
        groupId,
        payload.attributes
      );
      Sandbox.assert.calledWithExactly(
        groupSaveStub,
        <any>R.omit(['attributes', 'policies'], { ...payload, id: groupId })
      );
      Sandbox.assert.calledWithExactly(
        bulkUpsertPoliciesForGroupStub,
        groupId,
        payload.policies,
        loggedInUserId
      );
      Sandbox.assert.calledWithExactly(getStub, groupId);
      Sandbox.assert.callCount(getStub, 2);
      Code.expect(response).to.equal({
        id: groupId,
        ...payload,
        policyOperationResult: [{ operation: 'create', success: true }]
      });
    });
  });
});
