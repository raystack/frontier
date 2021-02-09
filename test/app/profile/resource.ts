import Lab from '@hapi/lab';
import Sinon from 'sinon';
import Code from 'code';
import { factory } from 'typeorm-seeding';
import { lab } from '../../setup';
import { User } from '../../../model/user';
import * as Resource from '../../../app/profile/resource';

exports.lab = Lab.script();
const Sandbox = Sinon.createSandbox();

lab.afterEach(() => {
  Sandbox.restore();
});

lab.experiment('User::resource', () => {
  lab.experiment('get user', () => {
    let user;
    lab.beforeEach(async () => {
      user = await factory(User)().create();
    });

    lab.afterEach(() => {
      Sandbox.restore();
    });

    lab.test('should get user by email', async () => {
      const response = await Resource.getUserByEmail(user.email);
      Code.expect(response.username).to.equal(user.username);
      Code.expect(response.email).to.equal(user.email);
      Code.expect(response.designation).to.equal(user.designation);
    });
  });

  lab.experiment('update user', () => {
    let user;

    lab.beforeEach(async () => {
      user = await factory(User)().create();
    });

    lab.afterEach(() => {
      Sandbox.restore();
    });

    lab.test('should update user by email', async () => {
      const payload = {
        username: 'Demo username',
        name: 'Demo user',
        company: 'Demo company'
      };

      const response = await Resource.updateUserByEmail(user.email, payload);
      Code.expect(response.username).to.equal(payload.username);
      Code.expect(response.name).to.equal(payload.name);
      Code.expect(response.company).to.equal(payload.company);
    });
  });
});
