import Lab from '@hapi/lab';
import Sinon from 'sinon';
import Code from 'code';
import { factory } from 'typeorm-seeding';
import { lab } from '../../setup';
import { User } from '../../../model/user';
import * as Resource from '../../../app/user/resource';

exports.lab = Lab.script();
const Sandbox = Sinon.createSandbox();

lab.afterEach(() => {
  Sandbox.restore();
});

lab.experiment('User::resource', () => {
  lab.experiment('create user', () => {
    let user;

    lab.beforeEach(async () => {
      user = await factory(User)().create();
    });

    lab.afterEach(() => {
      Sandbox.restore();
    });

    lab.test('should update user by email', async () => {
      const response = await Resource.create(user);
      Code.expect(response.displayName).to.equal(user.displayName);
      Code.expect(response.metadata.email).to.equal(user.metadata.email);
    });
  });
});
