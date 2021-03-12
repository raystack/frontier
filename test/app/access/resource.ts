import Lab from '@hapi/lab';
import * as R from 'ramda';
import Sinon from 'sinon';
import Code from 'code';
import { lab } from '../../setup';
import * as Resource from '../../../app/access/resource';
import CasbinSingleton from '../../../lib/casbin';

exports.lab = Lab.script();
const Sandbox = Sinon.createSandbox();

lab.afterEach(() => {
  Sandbox.restore();
});

lab.experiment('Access::resource', () => {
  lab.experiment('check access ', () => {
    lab.afterEach(() => {
      Sandbox.restore();
    });

    lab.test('should evaluate accessList', async () => {
      const accessList: Resource.AccessList = [
        {
          subject: { user: '12f314' },
          resource: { entity: 'gojek' },
          action: { action: 'any' }
        },
        {
          subject: { user: '212314' },
          resource: { entity: 'gopay' },
          action: { action: 'any' }
        }
      ];

      const batchEnforceJsonStub = Sandbox.stub().resolves([true, true]);
      Sandbox.stub(CasbinSingleton, 'enforcer').value({
        batchEnforceJson: batchEnforceJsonStub
      });

      const response = await Resource.checkAccess(accessList);
      const expectedResponse = accessList.map(R.assoc('hasAccess', true));
      Code.expect(response).to.equal(expectedResponse);
      Sandbox.assert.calledWithExactly(batchEnforceJsonStub, accessList);
    });
  });
});
