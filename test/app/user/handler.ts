import Code from 'code';
import Lab from '@hapi/lab';
import Hapi from '@hapi/hapi';
import Sinon from 'sinon';
import { factory } from 'typeorm-seeding';
import { lab } from '../../setup';
import { User } from '../../../model/user';
import * as Config from '../../../config/config';
import * as userPlugin from '../../../app/user';
import * as iapPlugin from '../../../plugin/iap';
import * as Resource from '../../../app/user/resource';

exports.lab = Lab.script();
let server;
const Sandbox = Sinon.createSandbox();

lab.before(async () => {
  const plugins = [iapPlugin, userPlugin];
  server = new Hapi.Server({ port: Config.get('/port/web'), debug: false });
  await server.register(plugins);
});

lab.after(async () => {
  await server.stop();
});

lab.experiment('User::Handler', () => {
  lab.experiment('get user', () => {
    let request, user, getUserByEmailStub;

    lab.beforeEach(async () => {
      request = {
        method: 'GET',
        url: `/api/profile`,
        headers: {
          'x-goog-authenticated-user-email': 'praveen.yadav@gojek.com'
        }
      };
      user = await factory(User)().create();
      getUserByEmailStub = Sandbox.stub(Resource, 'getUserByEmail');
    });

    lab.afterEach(() => {
      Sandbox.restore();
    });

    lab.test('should return current profile', async () => {
      const credentials = {
        id: 1,
        username: 'demo',
        email: 'demo@demo.com'
      };
      await Resource.getUserByEmail(credentials.email);
      Sandbox.assert.calledWithExactly(getUserByEmailStub, credentials.email);

      getUserByEmailStub.resolves(user);
      const response = await server.inject(request);
      Code.expect(response.statusCode).to.equal(200);
    });
  });

  lab.experiment('update user', () => {
    let updateUserByEmailStub;

    lab.beforeEach(async () => {
      updateUserByEmailStub = Sandbox.stub(Resource, 'updateUserByEmail');
    });

    lab.afterEach(() => {
      Sandbox.restore();
    });

    lab.test('should update current profile', async () => {
      const credentials = {
        id: 1,
        username: 'demo',
        email: 'demo@demo.com'
      };
      const payload = {
        name: 'User1',
        email: 'demo@demo.com'
      };

      await Resource.updateUserByEmail(credentials.email, payload);
      Sandbox.assert.calledWithExactly(
        updateUserByEmailStub,
        credentials.email,
        payload
      );
    });
  });
});
