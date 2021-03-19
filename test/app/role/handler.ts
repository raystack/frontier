import Code from 'code';
import Lab from '@hapi/lab';
import * as R from 'ramda';
import Hapi from '@hapi/hapi';
import Sinon from 'sinon';
import { factory } from 'typeorm-seeding';
import { lab } from '../../setup';
import { Role } from '../../../model/role';
import * as Config from '../../../config/config';
import * as rolePlugin from '../../../app/role';
import * as Resource from '../../../app/role/resource';

exports.lab = Lab.script();
let server: Hapi.Server;
const Sandbox = Sinon.createSandbox();

const TEST_AUTH = {
  strategy: 'test',
  credentials: { id: 'dev.test' }
};

lab.before(async () => {
  const plugins = [rolePlugin];
  server = new Hapi.Server({ port: Config.get('/port/web'), debug: false });
  await server.register(plugins);
});

lab.after(async () => {
  await server.stop();
});

lab.afterEach(() => {
  Sandbox.restore();
});

lab.experiment('Role::Handler', () => {
  lab.experiment('get role by attributes', () => {
    let request: any, role: Role, getStub: any;

    lab.beforeEach(async () => {
      role = await factory(Role)().create();
      getStub = Sandbox.stub(Resource, 'get');
    });

    lab.afterEach(() => {
      getStub.restore();
    });

    lab.test('should return roles based on attributes', async () => {
      const roles: Role[] = [role];
      request = {
        method: 'GET',
        url: `/api/roles?attributes=entity&attributes=landscape`
      };
      getStub.resolves(roles);
      const response = await server.inject(request);
      Sandbox.assert.calledWithExactly(getStub, ['entity', 'landscape']);
      Code.expect(response.result).to.equal(roles);
      Code.expect(response.statusCode).to.equal(200);
    });
  });

  lab.experiment('create role with actions', () => {
    let request: any, role: Role, createStub: any;

    lab.beforeEach(async () => {
      role = await factory(Role)().create();
      createStub = Sandbox.stub(Resource, 'create');
    });

    lab.afterEach(() => {
      createStub.restore();
    });

    lab.test('should return roles based on attributes', async () => {
      request = {
        method: 'POST',
        url: `/api/roles`,
        payload: {
          displayname: role.displayname,
          attributes: role.attributes,
          metadata: role.metadata,
          actions: [
            { operation: 'create', action: 'test1' },
            { operation: 'create', action: 'test2' }
          ]
        },
        auth: TEST_AUTH
      };
      createStub.resolves(role);
      const response = await server.inject(request);
      Sandbox.assert.calledWithExactly(
        createStub,
        request.payload,
        request.auth.credentials
      );
      Code.expect(response.result).to.equal(role);
      Code.expect(response.statusCode).to.equal(200);
    });
  });

  lab.experiment('update role by id with actions', () => {
    let request: any, role: Role, updateStub: any;

    lab.beforeEach(async () => {
      role = await factory(Role)().create();
      updateStub = Sandbox.stub(Resource, 'update');
    });

    lab.afterEach(() => {
      updateStub.restore();
    });

    lab.test('should return roles based on attributes', async () => {
      request = {
        method: 'PUT',
        url: `/api/roles/${role.id}`,
        payload: {
          displayname: `${role.displayname}1`,
          actions: [
            { operation: 'create', action: 'test1' },
            { operation: 'create', action: 'test2' }
          ]
        },
        auth: TEST_AUTH
      };
      updateStub.resolves({
        ...role,
        displayname: request.payload.displayname
      });
      const response = await server.inject(request);
      Sandbox.assert.calledWithExactly(
        updateStub,
        role.id,
        request.payload,
        request.auth.credentials
      );
      Code.expect(R.path(['displayname'], response.result)).to.equal(
        request.payload.displayname
      );
      Code.expect(response.statusCode).to.equal(200);
    });
  });
});
