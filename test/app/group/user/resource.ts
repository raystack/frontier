import Lab from '@hapi/lab';
import Sinon from 'sinon';
import Code from 'code';
import * as R from 'ramda';
import { factory } from 'typeorm-seeding';
import { lab } from '../../../setup';
import * as Resource from '../../../../src/app/group/user/resource';
import CasbinSingleton from '../../../../src/lib/casbin';
import * as PolicyResource from '../../../../src/app/policy/resource';
import { User } from '../../../../src/model/user';
import { Group } from '../../../../src/model/group';
import { Role } from '../../../../src/model/role';
import * as Config from '../../../../src/config/config';

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

        const user = await factory(User)().create();
        const groupId = 'test_group';
        const userId = user.id;
        const loggedInUserId = user.id;
        const payload = {
          policies: [
            {
              operation: 'create',
              subject: { user: user.id },
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
        const user = await factory(User)().create();
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

        const response = await Resource.remove(groupId, userId, user.id);

        Sandbox.assert.calledWithExactly(
          getPoliciesBySubjectStub,
          { user: userId },
          { group: groupId }
        );
        Sandbox.assert.calledWithExactly(
          removeSubjectGroupingJsonPolicyStub,
          { user: userId },
          { group: groupId },
          { created_by: user }
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
        const user = await factory(User)().create();
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

        const response = await Resource.remove(groupId, userId, user.id);

        Sandbox.assert.calledWithExactly(
          getPoliciesBySubjectStub,
          { user: userId },
          { group: groupId }
        );
        Sandbox.assert.calledWithExactly(
          removeSubjectGroupingJsonPolicyStub,
          { user: userId },
          { group: groupId },
          { created_by: user }
        );
        Sandbox.assert.notCalled(removeJsonPolicyStub);
        Code.expect(response).to.equal(true);
      }
    );
  });

  lab.experiment('list users of a group user', () => {
    let users, groups, enforcer;
    const roles = [];

    const ABC_TAG = 'abc';
    const DEF_TAG = 'def';

    lab.before(async () => {
      const dbUri = Config.get('/postgres').uri;
      enforcer = await CasbinSingleton.create(dbUri);
    });

    lab.afterEach(() => {
      Sandbox.restore();
    });

    lab.beforeEach(async () => {
      // setup 2 users
      users = await factory(User)().createMany(3);
      // set up a group
      groups = await factory(Group)().createMany(2);
      // setup 2 roles with tags and 1 without
      roles[0] = await factory(Role)().create({ tags: [ABC_TAG, 'test'] });
      roles[1] = await factory(Role)().create({ tags: [DEF_TAG] });
      roles[2] = await factory(Role)().create({ tags: [] });

      // setup policies for users

      // add users[0] and users[1] to groups[0]
      await enforcer.addSubjectGroupingJsonPolicy(
        { user: users[0].id },
        { group: groups[0].id },
        { created_by: {} }
      );
      await enforcer.addSubjectGroupingJsonPolicy(
        { user: users[1].id },
        { group: groups[0].id },
        { created_by: {} }
      );

      // add users[2] to groups[1], just to check whether filtering is working
      await enforcer.addSubjectGroupingJsonPolicy(
        { user: users[2].id },
        { group: groups[1].id },
        { created_by: {} }
      );

      // assign users[0] roles[0]
      await enforcer.addJsonPolicy(
        { user: users[0].id },
        { group: groups[0].id, test: 123 },
        { role: roles[0].id },
        { created_by: {} }
      );

      // assign users[1] roles[1]
      await enforcer.addJsonPolicy(
        { user: users[1].id },
        { group: groups[0].id, test: 123 },
        { role: roles[1].id },
        { created_by: {} }
      );

      // assign users[2] roles[2]
      await enforcer.addJsonPolicy(
        { user: users[2].id },
        { group: groups[1].id, test: 3453 },
        { role: roles[2].id },
        { created_by: {} }
      );
    });

    lab.test('should return users of a group with policies', async () => {
      const groupId = groups[0].id;
      const result = await Resource.list(groupId);

      Code.expect(result.length).to.equal(2);
      Code.expect(result[0].policies.length).to.equal(1);
      Code.expect(result[1].policies.length).to.equal(1);
    });

    lab.test(
      'should return users of a group with role filtered policies for abc tag',
      async () => {
        const groupId = groups[0].id;
        const path = ['fields', 'policies', '$filter', 'role', 'tags'];

        const filter = R.assocPath(path, ABC_TAG, {});
        const result = await Resource.list(groupId, filter);
        const abcUser = result.find((r: any) => r.id === users[0].id);
        const defUser = result.find((r: any) => r.id === users[1].id);

        Code.expect(result.length).to.equal(2);
        Code.expect(abcUser.policies.length).to.equal(1);
        Code.expect(defUser.policies.length).to.equal(0);
      }
    );

    lab.test(
      'should return users of a group with role filtered policies for def tag',
      async () => {
        const groupId = groups[0].id;
        const path = ['fields', 'policies', '$filter', 'role', 'tags'];

        const filter = R.assocPath(path, DEF_TAG, {});
        const result = await Resource.list(groupId, filter);
        const abcUser = result.find((r: any) => r.id === users[0].id);
        const defUser = result.find((r: any) => r.id === users[1].id);

        Code.expect(result.length).to.equal(2);
        Code.expect(abcUser.policies.length).to.equal(0);
        Code.expect(defUser.policies.length).to.equal(1);
      }
    );
  });
});
