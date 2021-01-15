import Lab from '@hapi/lab';
import Sinon from 'sinon';
import { lab } from '../../setup';
import * as Resource from '../../../app/user/resource';

exports.lab = Lab.script();
const Sandbox = Sinon.createSandbox();

lab.afterEach(() => {
  Sandbox.restore();
});

lab.experiment('User::resource', () => {
  lab.experiment('get user', () => {
    let getUserByEmailStub;

    lab.beforeEach(async () => {
      getUserByEmailStub = Sandbox.stub(Resource, 'getUserByEmail');
    });

    lab.afterEach(() => {
      Sandbox.restore();
    });

    lab.test('should get user by email', async () => {
      const email = 'demo@demo.com';
      await Resource.getUserByEmail(email);
      Sandbox.assert.calledWithExactly(getUserByEmailStub, email);
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

    lab.test('should update user by email', async () => {
      const email = 'demo@demo.com';
      const payload = {
        name: 'User1',
        email: 'demo@demo.com'
      };

      await Resource.updateUserByEmail(email, payload);
      Sandbox.assert.calledWithExactly(updateUserByEmailStub, email, payload);
    });
  });
});
