import Code from 'code';
import Lab from '@hapi/lab';
import Hapi from '@hapi/hapi';
import Sinon from 'sinon';
import { lab } from '../../setup';
import * as iapPlugin from '../../../plugin/iap';
import * as profilePlugin from '../../../app/profile';
import * as Config from '../../../config/config';

exports.lab = Lab.script();
let server: Hapi.Server;
const Sandbox = Sinon.createSandbox();

lab.before(async () => {
  const plugins = [iapPlugin, profilePlugin];
  server = new Hapi.Server({ port: Config.get('/port/web'), debug: false });
  await server.register(plugins);
});

lab.after(async () => {
  await server.stop();
});

lab.afterEach(() => {
  Sandbox.restore();
});

lab.experiment('IAPPlugin', () => {
  let request;

  lab.test('should reject /api/profile on missing iap headers', async () => {
    request = {
      method: 'GET',
      url: `/api/profile`
    };
    const response = await server.inject(request);
    Code.expect(response.statusCode).to.equal(401);
  });

  lab.test('should allow /api/profile on providing iap headers', async () => {
    request = {
      method: 'GET',
      url: `/api/profile`,
      headers: {
        'x-goog-authenticated-user-email': 'praveen.yadav@gojek.com'
      }
    };
    const response = await server.inject(request);
    Code.expect(response.statusCode).to.equal(200);
  });
});
