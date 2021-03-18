import Code from 'code';
import Lab from '@hapi/lab';
import Hapi from '@hapi/hapi';
import Sinon from 'sinon';
import { lab } from '../../../setup';
import * as Config from '../../../../config/config';
import * as userPlugin from '../../../../app/user';
import * as UserGroupResource from '../../../../app/user/group/resource';

exports.lab = Lab.script();
let server: Hapi.Server;
const Sandbox = Sinon.createSandbox();

const TEST_AUTH = {
  strategy: 'test',
  credentials: { id: 'dev.test' }
};

lab.before(async () => {
  const plugins = [userPlugin];
  server = new Hapi.Server({ port: Config.get('/port/web'), debug: false });
  await server.register(plugins);
});

lab.after(async () => {
  await server.stop();
});

lab.afterEach(() => {
  Sandbox.restore();
});

lab.experiment('UserGroup::Handler', () => {
  lab.experiment('get groups of a user by id', () => {
    let getStub: any;

    lab.afterEach(() => {
      getStub.restore();
    });

    lab.test('should get groups of a user by id', async () => {
      const USER_ID = 'test-user';
      const request: any = {
        method: 'GET',
        url: `/api/users/${USER_ID}/groups?action=123`,
        auth: TEST_AUTH
      };

      const result: any = [
        {
          policies: [
            {
              subject: { user: 'user' },
              resource: { group: 'group' },
              action: { role: 'role' }
            }
          ],
          attributes: { group: 'group' }
        }
      ];
      getStub = Sandbox.stub(UserGroupResource, 'list').returns(result);

      const response = await server.inject(request);

      Sandbox.assert.calledWithExactly(getStub, USER_ID, { action: '123' });
      Code.expect(response.result).to.equal(result);
      Code.expect(response.statusCode).to.equal(200);
    });

    lab.test('should get groups of logged in user', async () => {
      const USER_ID = TEST_AUTH.credentials.id;
      const request: any = {
        method: 'GET',
        url: `/api/users/self/groups`,
        auth: TEST_AUTH
      };

      const result: any = [
        {
          policies: [
            {
              subject: { user: 'user' },
              resource: { group: 'group' },
              action: { role: 'role' }
            }
          ],
          attributes: { group: 'group' }
        }
      ];
      getStub = Sandbox.stub(UserGroupResource, 'list').returns(result);

      const response = await server.inject(request);

      Sandbox.assert.calledWithExactly(getStub, USER_ID, {});
      Code.expect(response.result).to.equal(result);
      Code.expect(response.statusCode).to.equal(200);
    });
  });
});
