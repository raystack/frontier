import Code from 'code';
import Lab from '@hapi/lab';
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
});
