import Lab from '@hapi/lab';
import Sinon from 'sinon';
import Code from 'code';
import { lab } from '../../../setup';
import * as Resource from '../../../../app/group/user/resource';
import CasbinSingleton from '../../../../lib/casbin';
import * as PolicyResource from '../../../../app/policy/resource';
import { User } from '../../../../model/user';

exports.lab = Lab.script();
const Sandbox = Sinon.createSandbox();

lab.afterEach(() => {
  Sandbox.restore();
});

lab.experiment('Group:User:Mapping::resource', () => {
  lab.experiment('update group user mapping', () => {
    lab.afterEach(() => {
      Sandbox.restore();
    });

    lab.test(
      'should update group user mapping based on group_id and user_id',
      async () => {
        const groupId = 'test_group';
        const userId = 'test_user';
        const loggedInUserId = 'test_logged_in_user';
        const payload = {
          policies: [
            {
              operation: 'create',
              subject: { user: 'test_user' },
              resource: {
                group: 'test_group',
                entity: 'gojek',
                landscape: 'id',
                environment: 'production'
              },
              action: { permission: 'test_permission' }
            },
            {
              operation: 'delete',
              subject: { user: 'delete_test_user' },
              resource: {
                group: 'delete_test_group',
                entity: 'gojek',
                landscape: 'id',
                environment: 'production'
              },
              action: { permission: 'delete_test_permission' }
            }
          ]
        };

        const result: any = payload.policies.map((policy: any) => {
          return { ...policy, success: true };
        });
        Sandbox.stub(PolicyResource, 'bulkOperation').returns(result);
        const response = await Resource.update(
          groupId,
          userId,
          loggedInUserId,
          payload
        );

        response.forEach((row, index) => {
          Code.expect(row.subject).to.equal(result[index].subject);
          Code.expect(row.resource).to.equal(result[index].resource);
          Code.expect(row.action).to.equal(result[index].action);
          Code.expect(row.success).to.equal(true);
        });
      }
    );
  });

  lab.experiment('create group user mapping', () => {
    lab.afterEach(() => {
      Sandbox.restore();
    });

    lab.test(
      'should create group user mapping based on group_id and user_id',
      async () => {
        const addSubjectGroupingJsonPolicyStub = Sandbox.stub();
        Sandbox.stub(CasbinSingleton, 'enforcer').value({
          addSubjectGroupingJsonPolicy: addSubjectGroupingJsonPolicyStub
        });

        const groupId = 'test_group';
        const userId = 'test_user';
        const loggedInUserId = 'test_logged_in_user';
        const payload = {
          policies: [
            {
              operation: 'create',
              subject: { user: 'test_user' },
              resource: {
                group: 'test_group',
                entity: 'gojek',
                landscape: 'id',
                environment: 'production'
              },
              action: { permission: 'test_permission' }
            }
          ]
        };
        const result: any = payload.policies.map((policy) => {
          return { ...policy, success: true };
        });
        Sandbox.stub(PolicyResource, 'bulkOperation').returns(result);
        const response = await Resource.create(
          groupId,
          userId,
          loggedInUserId,
          payload
        );
        Code.expect(response[0].subject).to.equal(result[0].subject);
        Code.expect(response[0].resource).to.equal(result[0].resource);
        Code.expect(response[0].action).to.equal(result[0].action);
        Code.expect(response[0].success).to.equal(true);
      }
    );
  });

  lab.experiment('get group user mapping', () => {
    const userId = 'test-user';
    const groupId = 'test-group';

    lab.afterEach(() => {
      Sandbox.restore();
    });

    lab.test('should return user group mapping', async () => {
      const getGroupUserMappingStub = Sandbox.stub(
        PolicyResource,
        'getGroupUserMapping'
      ).returns(<any>{ user: userId, group: groupId });
      const getUserStub = Sandbox.stub(User, 'findOne').returns(<any>{
        id: userId
      });
      const getPoliciesBySubjectStub = Sandbox.stub(
        PolicyResource,
        'getPoliciesBySubject'
      ).returns(<any>[]);

      const result = await Resource.get(groupId, userId);
      const expectedResult: any = { id: userId, policies: [] };

      Sandbox.assert.calledWithExactly(
        getGroupUserMappingStub,
        groupId,
        userId
      );
      Sandbox.assert.calledWithExactly(getUserStub, <any>userId);
      Sandbox.assert.calledWithExactly(
        getPoliciesBySubjectStub,
        { user: userId },
        { group: groupId }
      );
      Code.expect(result).to.equal(expectedResult);
    });

    lab.test('should return Boom 404 if user is not found', async () => {
      const getGroupUserMappingStub = Sandbox.stub(
        PolicyResource,
        'getGroupUserMapping'
      ).returns(null);
      const getUserStub = Sandbox.stub(User, 'findOne').returns(<any>{
        id: userId
      });
      const getPoliciesBySubjectStub = Sandbox.stub(
        PolicyResource,
        'getPoliciesBySubject'
      ).returns(<any>[]);

      const result: any = await Resource.get(groupId, userId);

      Sandbox.assert.calledWithExactly(
        getGroupUserMappingStub,
        groupId,
        userId
      );
      Sandbox.assert.notCalled(getUserStub);
      Sandbox.assert.notCalled(getPoliciesBySubjectStub);
      Code.expect(result?.isBoom).to.equal(true);
    });
  });

  lab.experiment('remove group user mapping', () => {
    lab.afterEach(() => {
      Sandbox.restore();
    });

    lab.test(
      'should remove group user mapping based on group_id and user_id',
      async () => {
        const groupId = 'test_group';
        const userId = 'test_user';
        const payload: any = {
          policies: [
            {
              operation: 'delete',
              subject: { user: 'test_user' },
              resource: {
                group: 'test_group',
                entity: 'gojek',
                landscape: 'id',
                environment: 'production'
              },
              action: { permission: 'test_permission' }
            }
          ]
        };
        const removeSubjectGroupingJsonPolicyStub = Sandbox.stub();
        const removeJsonPolicyStub = Sandbox.stub();
        Sandbox.stub(CasbinSingleton, 'enforcer').value({
          removeSubjectGroupingJsonPolicy: removeSubjectGroupingJsonPolicyStub,
          removeJsonPolicy: removeJsonPolicyStub
        });

        const getPoliciesBySubjectStub = Sandbox.stub(
          PolicyResource,
          'getPoliciesBySubject'
        ).returns(payload.policies);

        const response = await Resource.remove(groupId, userId);

        Sandbox.assert.calledWithExactly(
          getPoliciesBySubjectStub,
          { user: userId },
          { group: groupId }
        );
        Sandbox.assert.calledWithExactly(
          removeSubjectGroupingJsonPolicyStub,
          { user: userId },
          { group: groupId }
        );
        Sandbox.assert.callCount(removeJsonPolicyStub, payload.policies.length);
        Code.expect(response).to.equal(true);
      }
    );

    lab.test(
      'should remove group user mapping based on group_id and user_id even if no policies are present',
      async () => {
        const groupId = 'test_group';
        const userId = 'test_user';
        const payload: any = {
          policies: []
        };
        const removeSubjectGroupingJsonPolicyStub = Sandbox.stub();
        const removeJsonPolicyStub = Sandbox.stub();
        Sandbox.stub(CasbinSingleton, 'enforcer').value({
          removeSubjectGroupingJsonPolicy: removeSubjectGroupingJsonPolicyStub,
          removeJsonPolicy: removeJsonPolicyStub
        });

        const getPoliciesBySubjectStub = Sandbox.stub(
          PolicyResource,
          'getPoliciesBySubject'
        ).returns(payload.policies);

        const response = await Resource.remove(groupId, userId);

        Sandbox.assert.calledWithExactly(
          getPoliciesBySubjectStub,
          { user: userId },
          { group: groupId }
        );
        Sandbox.assert.calledWithExactly(
          removeSubjectGroupingJsonPolicyStub,
          { user: userId },
          { group: groupId }
        );
        Sandbox.assert.notCalled(removeJsonPolicyStub);
        Code.expect(response).to.equal(true);
      }
    );
  });
});
