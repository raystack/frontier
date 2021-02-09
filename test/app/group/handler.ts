import Code from 'code';
import Lab from '@hapi/lab';
import Hapi from '@hapi/hapi';
import Sinon from 'sinon';
import { factory } from 'typeorm-seeding';
import { lab } from '../../setup';
import { Group } from '../../../model/group';
import * as Config from '../../../config/config';
import * as groupPlugin from '../../../app/group';
import * as Resource from '../../../app/group/resource';

exports.lab = Lab.script();
let server;
const Sandbox = Sinon.createSandbox();

const TEST_AUTH = {
  strategy: 'test',
  credentials: { id: 'dev.test' }
};

lab.before(async () => {
  const plugins = [groupPlugin];
  server = new Hapi.Server({ port: Config.get('/port/web'), debug: false });
  await server.register(plugins);
});

lab.after(async () => {
  await server.stop();
});

lab.afterEach(() => {
  Sandbox.restore();
});

lab.experiment('Group::Handler', () => {
  lab.experiment('create group', () => {
    let request, createStub, payload;

    lab.beforeEach(async () => {
      createStub = Sandbox.stub(Resource, 'create');
      payload = {
        displayName: 'test title',
        metadata: {
          name: 'test',
          email: 'test@go-jek.com',
          privacy: 'public'
        },
        attributes: [{ entity: 'gojek' }],
        policies: []
      };
      request = {
        method: 'POST',
        url: `/api/groups`,
        payload,
        auth: TEST_AUTH
      };
    });

    lab.afterEach(() => {
      createStub.restore();
    });

    lab.test('should create group', async () => {
      createStub.resolves(payload);

      const response = await server.inject(request);

      Sandbox.assert.calledWithExactly(
        createStub,
        payload,
        TEST_AUTH.credentials.id
      );
      Code.expect(response.result).to.equal(payload);
      Code.expect(response.statusCode).to.equal(200);
    });
  });

  lab.experiment('get group by id', () => {
    let request, group, getStub;

    lab.beforeEach(async () => {
      group = await factory(Group)().create();
      getStub = Sandbox.stub(Resource, 'get');
      request = {
        method: 'GET',
        url: `/api/groups/${group.id}?test=123`,
        auth: TEST_AUTH
      };
    });

    lab.afterEach(() => {
      getStub.restore();
    });

    lab.test('should return current profile', async () => {
      getStub.resolves(group);
      const response = await server.inject(request);

      Sandbox.assert.calledWithExactly(getStub, group.id.toString(), {
        test: '123'
      });
      Code.expect(response.result).to.equal(group);
      Code.expect(response.statusCode).to.equal(200);
    });
  });

  lab.experiment('update group by id', () => {
    let request, group, updateStub, payload;

    lab.beforeEach(async () => {
      group = await factory(Group)().create();
      updateStub = Sandbox.stub(Resource, 'update');
      payload = {
        displayName: 'test title',
        metadata: {
          name: 'test',
          email: 'test@go-jek.com',
          privacy: 'public'
        },
        attributes: [{ entity: 'gojek' }],
        policies: []
      };
      request = {
        method: 'PUT',
        url: `/api/groups/${group.id}`,
        payload,
        auth: TEST_AUTH
      };
    });

    lab.afterEach(() => {
      updateStub.restore();
    });

    lab.test('should update group by id', async () => {
      updateStub.resolves(group);

      const response = await server.inject(request);

      Sandbox.assert.calledWithExactly(
        updateStub,
        group.id.toString(),
        payload,
        TEST_AUTH.credentials.id
      );
      Code.expect(response.result).to.equal(group);
      Code.expect(response.statusCode).to.equal(200);
    });
  });
});
