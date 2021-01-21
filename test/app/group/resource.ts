import Lab from '@hapi/lab';
import Sinon from 'sinon';
import Code from 'code';
import { factory } from 'typeorm-seeding';
import { lab } from '../../setup';
import { Group } from '../../../model/group';
import * as Resource from '../../../app/group/resource';

exports.lab = Lab.script();
const Sandbox = Sinon.createSandbox();

lab.afterEach(() => {
  Sandbox.restore();
});

lab.experiment('Group::resource', () => {
  lab.experiment('get group by id', () => {
    let group;

    lab.beforeEach(async () => {
      group = await factory(Group)().create();
    });

    lab.afterEach(() => {
      Sandbox.restore();
    });

    lab.test('should get group by id', async () => {
      const response = await Resource.get(group.id);
      Code.expect(response).to.equal(group);
    });

    lab.test(
      'should return undefined response if group is not found',
      async () => {
        const response = await Resource.get(55);
        Code.expect(response).to.undefined();
      }
    );
  });

  lab.experiment('update group by id', () => {
    let group;

    lab.beforeEach(async () => {
      group = await factory(Group)().create();
    });

    lab.afterEach(() => {
      Sandbox.restore();
    });

    lab.test('should update group by id', async () => {
      const payload = {
        title: 'Updated Title'
      };
      const response = await Resource.update(group.id, payload);
      Code.expect(response.title).to.equal(payload.title);
    });
  });

  lab.experiment('create group', () => {
    lab.afterEach(() => {
      Sandbox.restore();
    });

    lab.test('should update group by id', async () => {
      const payload = {
        name: 'de',
        title: 'Data Engineering',
        privacy: 'private'
      };
      const response = await Resource.create(payload);
      Code.expect(response.name).to.equal(payload.name);
    });
  });
});
