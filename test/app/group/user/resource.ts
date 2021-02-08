import Lab from '@hapi/lab';
import Sinon from 'sinon';
import Code from 'code';
import { factory } from 'typeorm-seeding';
import { lab } from '../../../setup';
import * as Resource from '../../../../app/group/user/resource';
import CasbinSingleton from '../../../../lib/casbin';
import * as PolicyResource from '../../../../app/policy/resource';

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

    lab.test.only(
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
    const addSubjectGroupingJsonPolicyStub = Sandbox.stub();
    Sandbox.stub(CasbinSingleton, 'enforcer').value({
      addSubjectGroupingJsonPolicy: addSubjectGroupingJsonPolicyStub
    });

    lab.afterEach(() => {
      Sandbox.restore();
    });

    lab.test(
      'should create group user mapping based on group_id and user_id',
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
});
