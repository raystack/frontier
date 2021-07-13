import Lab from '@hapi/lab';
import Sinon from 'sinon';
import Code from 'code';
import { factory } from 'typeorm-seeding';
import { lab } from '../../setup';
import { Role } from '../../../src/model/role';
import * as Resource from '../../../src/app/role/resource';

exports.lab = Lab.script();
const Sandbox = Sinon.createSandbox();

lab.afterEach(() => {
  Sandbox.restore();
});

lab.experiment('Role::resource', () => {
  lab.experiment(
    'get role by attributes ( entity , environment , landscape etc. )',
    () => {
      let role: Role;
      lab.beforeEach(async () => {
        role = await factory(Role)().create();
      });

      lab.afterEach(() => {
        Sandbox.restore();
      });

      lab.test('should get role by attributes', async () => {
        const response = await Resource.get(['entity']);
        Code.expect(response).to.equal([role]);
      });

      lab.test(
        'should return all roles saved in system if no attribute is passed',
        async () => {
          const response = await Resource.get([]);
          Code.expect(response).to.equal([role]);
        }
      );

      lab.test(
        'should return empty array response if passed attributes not found',
        async () => {
          const response = await Resource.get(['attribute not found']);
          Code.expect(response).to.equal([]);
        }
      );
    }
  );

  lab.experiment('get role by tags', () => {
    let role: Role;

    lab.afterEach(() => {
      Sandbox.restore();
    });

    lab.test('should get role by tags', async () => {
      role = await factory(Role)().create({ tags: ['abc'] });
      const response = await Resource.get([], { tags: ['abc'] });
      Code.expect(response).to.equal([role]);
    });

    lab.test('should ignore unmatching tags', async () => {
      role = await factory(Role)().create({ tags: ['abc'] });
      const response = await Resource.get([], { tags: ['def'] });
      Code.expect(response).to.not.equal([role]);
    });

    lab.test('should match multiple tags', async () => {
      const roles = await factory(Role)().createMany(5, {
        tags: ['abc', 'def']
      });
      const response = await Resource.get([], { tags: ['def', 'abc'] });
      Code.expect(response).to.equal(roles);
    });
  });

  lab.experiment('create role along with action mapping', () => {
    const loggedInUser = { username: 'test' };
    let mapActionRoleInBulkStub = null;

    lab.beforeEach(() => {
      mapActionRoleInBulkStub = Sandbox.stub(
        Resource,
        'mapActionRoleInBulk'
      ).resolves([]);
    });

    lab.afterEach(() => {
      Sandbox.restore();
    });

    lab.test('should create role even without actions', async () => {
      const payload = {
        displayname: 'role',
        attributes: ['test'],
        metadata: {},
        tags: ['tag']
      };

      const result = await Resource.create(payload, loggedInUser);

      Code.expect(result.displayname).to.equal(payload.displayname);
      Sandbox.assert.notCalled(mapActionRoleInBulkStub);
    });

    lab.test('should create role with action', async () => {
      const payload = {
        displayname: 'role',
        attributes: ['test'],
        metadata: {},
        actions: [{ operation: 'create', action: 'test' }]
      };

      const result = await Resource.create(<any>payload, loggedInUser);

      Code.expect(result.displayname).to.equal(payload.displayname);
      Sandbox.assert.calledOnceWithExactly(
        mapActionRoleInBulkStub,
        result.id,
        payload.actions,
        loggedInUser
      );
    });
  });

  lab.experiment('update role by id along with action mapping', () => {
    const loggedInUser = { username: 'test' };
    let role,
      mapActionRoleInBulkStub = null;

    lab.beforeEach(async () => {
      mapActionRoleInBulkStub = Sandbox.stub(
        Resource,
        'mapActionRoleInBulk'
      ).resolves([]);
      role = await factory(Role)().create();
    });

    lab.afterEach(() => {
      Sandbox.restore();
    });

    lab.test('should update role even without actions', async () => {
      const payload = {
        displayname: 'role',
        attributes: ['test'],
        metadata: {}
      };

      const result = await Resource.update(role.id, payload, loggedInUser);

      Code.expect(result.displayname).to.equal(payload.displayname);
      Sandbox.assert.notCalled(mapActionRoleInBulkStub);
    });

    lab.test('should update role with action', async () => {
      const payload = {
        displayname: 'role',
        attributes: ['test'],
        metadata: {},
        actions: [{ operation: 'create', action: 'test' }]
      };

      const result = await Resource.update(role.id, <any>payload, loggedInUser);

      Code.expect(result.displayname).to.equal(payload.displayname);
      Sandbox.assert.calledOnceWithExactly(
        mapActionRoleInBulkStub,
        result.id,
        payload.actions,
        loggedInUser
      );
    });
  });
});
