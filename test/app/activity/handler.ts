import Code from 'code';
import Lab from '@hapi/lab';
import Hapi from '@hapi/hapi';
import Sinon from 'sinon';
import { lab } from '../../setup';
import * as Config from '../../../config/config';
import * as activityPlugin from '../../../app/activity';
import * as Resource from '../../../app/activity/resource';

exports.lab = Lab.script();
let server: Hapi.Server;
const Sandbox = Sinon.createSandbox();

const TEST_AUTH = {
  strategy: 'test',
  credentials: { id: 'dev.test' }
};

lab.before(async () => {
  const plugins = [activityPlugin];
  server = new Hapi.Server({ port: Config.get('/port/web'), debug: false });
  await server.register(plugins);
});

lab.after(async () => {
  await server.stop();
});

lab.afterEach(() => {
  Sandbox.restore();
});

lab.experiment('Activity::Handler', () => {
  lab.experiment('get all activities', () => {
    let request: any, getStub: any, activities: any;

    lab.beforeEach(async () => {
      activities = [
        {
          createdAt: '2021-03-18T07:18:16.947Z',
          diff: {
            created: ['new test created']
          },
          id: '1149e63d-2c93-4ea7-acf7-809803536bcb',
          reason: 'Cheese Union PNG',
          user: 'dev.test'
        }
      ];
      getStub = Sandbox.stub(Resource, 'get');
      request = {
        method: 'GET',
        url: `/api/activities`,
        auth: TEST_AUTH
      };
    });
    lab.afterEach(() => {
      getStub.restore();
    });

    lab.test('should get user by id', async () => {
      getStub.resolves(activities);
      const response = await server.inject(request);
      Sandbox.assert.calledWithExactly(getStub);
      Code.expect(response.result).to.equal(activities);
      Code.expect(response.statusCode).to.equal(200);
    });
  });
});
