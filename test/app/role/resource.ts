import Lab from '@hapi/lab';
import Sinon from 'sinon';
import Code from 'code';
import { factory } from 'typeorm-seeding';
import { lab } from '../../setup';
import { Role } from '../../../model/role';
import * as Resource from '../../../app/role/resource';

exports.lab = Lab.script();
const Sandbox = Sinon.createSandbox();

lab.afterEach(() => {
  Sandbox.restore();
});

lab.experiment('Role::resource', () => {
  lab.experiment(
    'get role by attributes ( entity , environment , landscape etc. )',
    () => {
      let role: Role;
      lab.beforeEach(async () => {
        role = await factory(Role)().create();
      });

      lab.afterEach(() => {
        Sandbox.restore();
      });

      lab.test('should get role by attributes', async () => {
        const response = await Resource.get(['entity']);
        Code.expect(response).to.equal([role]);
      });

      lab.test(
        'should return all roles saved in system if no attribute is passed',
        async () => {
          const response = await Resource.get([]);
          Code.expect(response).to.equal([role]);
        }
      );

      lab.test(
        'should return empty array response if passed attributes not found',
        async () => {
          const response = await Resource.get(['attribute not found']);
          Code.expect(response).to.equal([]);
        }
      );
    }
  );
});
