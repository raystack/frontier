import Code from 'code';
import Lab from '@hapi/lab';
import Hapi from '@hapi/hapi';
import Sinon from 'sinon';
import { factory } from 'typeorm-seeding';
import { lab } from '../../setup';
import { User } from '../../../model/user';
import * as Config from '../../../config/config';
import * as userPlugin from '../../../app/profile';
import * as iapPlugin from '../../../plugin/iap';
import * as Resource from '../../../app/profile/resource';

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

lab.experiment('Profile::Handler', () => {
  lab.experiment('get current user', () => {
    let request, user, getUserByMetadataStub;

    lab.beforeEach(async () => {
      request = {
        method: 'GET',
        url: `/api/profile`,
        headers: {
          'x-goog-authenticated-user-email': 'praveen.yadav@gojek.com'
        }
      };
      user = await factory(User)().create();
      getUserByMetadataStub = Sandbox.stub(Resource, 'getUserByMetadata');
    });

    lab.afterEach(() => {
      Sandbox.restore();
    });

    lab.test('should return current profile', async () => {
      const email = 'demo@demo.com';

      await Resource.getUserByMetadata({ email });
      Sandbox.assert.calledWithExactly(getUserByMetadataStub, {
        email
      });

      getUserByMetadataStub.resolves(user);
      const response = await server.inject(request);
      Code.expect(response.statusCode).to.equal(200);
    });
  });

  lab.experiment('update current user', () => {
    let updateUserByMetadataStub;

    lab.beforeEach(async () => {
      updateUserByMetadataStub = Sandbox.stub(Resource, 'updateUserByMetadata');
    });

    lab.afterEach(() => {
      Sandbox.restore();
    });

    lab.test('should update current profile', async () => {
      const email = 'demo@demo.com';
      const payload = {
        displayname: 'demo',
        metadata: {
          username: 'demo',
          email: 'demo@demo.com'
        }
      };

      await Resource.updateUserByMetadata({ email }, payload);
      Sandbox.assert.calledWithExactly(
        updateUserByMetadataStub,
        { email },
        payload
      );
    });
  });
});
