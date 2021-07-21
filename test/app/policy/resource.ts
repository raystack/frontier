import Lab from '@hapi/lab';
import Sinon from 'sinon';
import Code from 'code';
import { lab } from '../../setup';
import * as Resource from '../../../src/app/policy/resource';
import CasbinSingleton from '../../../src/lib/casbin';

exports.lab = Lab.script();
const Sandbox = Sinon.createSandbox();

lab.afterEach(() => {
  Sandbox.restore();
});

lab.experiment('Policy::resource', () => {
  lab.experiment('Resource::bulkOperation', () => {
    lab.test(
      'should perform bulk operations on policies if access',
      async () => {
        const addJsonPolicyStub = Sandbox.stub();
        const removeJsonPolicyStub = Sandbox.stub();
        const enforceJsonStub = Sandbox.stub().resolves(true);

        Sandbox.stub(CasbinSingleton, 'enforcer').value({
          addJsonPolicy: addJsonPolicyStub,
          removeJsonPolicy: removeJsonPolicyStub,
          enforceJson: enforceJsonStub
        });

        const createPolicy = {
          subject: { username: 'shreyas' },
          resource: { entity: 'gojek', landscape: 'id' },
          action: { role: 'streaming.devs' }
        };
        const deletePolicy = {
          subject: { username: 'shreyas' },
          resource: { entity: 'gofin', landscape: 'id' },
          action: { role: 'streaming.devs' }
        };
        const policyOperations: Resource.PolicyOperation[] = [
          {
            operation: 'create',
            ...createPolicy
          },
          {
            operation: 'delete',
            ...deletePolicy
          }
        ];

        const subject = { username: 'test' };

        const expectedResult = policyOperations.map((p) => ({
          ...p,
          success: true
        }));

        const result = await Resource.bulkOperation(policyOperations, subject);

        Code.expect(result).to.equal(expectedResult);
        Sandbox.assert.called(enforceJsonStub);
        Sandbox.assert.calledWith(
          addJsonPolicyStub,
          createPolicy.subject,
          createPolicy.resource,
          createPolicy.action
        );
        Sandbox.assert.calledWith(
          removeJsonPolicyStub,
          deletePolicy.subject,
          deletePolicy.resource,
          deletePolicy.action
        );
      }
    );

    lab.test(
      'should fail bulk operations on policies if no access',
      async () => {
        const addJsonPolicyStub = Sandbox.stub();
        const removeJsonPolicyStub = Sandbox.stub();
        const enforceJsonStub = Sandbox.stub().resolves(false);

        Sandbox.stub(CasbinSingleton, 'enforcer').value({
          addJsonPolicy: addJsonPolicyStub,
          removeJsonPolicy: removeJsonPolicyStub,
          enforceJson: enforceJsonStub
        });

        const deletePolicy = {
          subject: { username: 'shreyas' },
          resource: { entity: 'gofin', landscape: 'id' },
          action: { role: 'streaming.devs' }
        };
        const policyOperations: Resource.PolicyOperation[] = [
          {
            operation: 'delete',
            ...deletePolicy
          }
        ];

        const subject = { username: 'test' };

        const expectedResult = policyOperations.map((p) => ({
          ...p,
          success: false
        }));

        const result = await Resource.bulkOperation(policyOperations, subject);

        Code.expect(result).to.equal(expectedResult);
        Sandbox.assert.called(enforceJsonStub);
        Sandbox.assert.notCalled(removeJsonPolicyStub);
      }
    );
  });
});
