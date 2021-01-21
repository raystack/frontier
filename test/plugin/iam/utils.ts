import Code from 'code';
import Lab from '@hapi/lab';
import Sinon from 'sinon';
import { lab } from '../../setup';
import * as IAMPluginUtils from '../../../plugin/iam/utils';

exports.lab = Lab.script();

const {
  getIAMAction,
  contructIAMActionFromConfig,
  constructIAMResourceFromConfig
} = IAMPluginUtils;

const Sandbox = Sinon.createSandbox();

lab.experiment('IAMPluginUtils::constructIAMResourceFromConfig', () => {
  lab.test(
    'should construct the resource object for the given transformConfig and request',
    async () => {
      const request = {
        query: {
          entity: 'platform',
          landscape: 'id'
        },
        params: {
          name: 'p-firehose-123'
        },
        payload: {
          environment: 'integration'
        }
      };

      const resourceTransformConfig = [
        {
          requestKey: 'entity',
          iamKey: 'newEntity'
        },
        {
          requestKey: 'name',
          iamKey: 'resource'
        },
        {
          requestKey: 'environment',
          iamKey: 'environment'
        }
      ];

      const expectedResult = {
        newEntity: 'platform',
        resource: 'p-firehose-123',
        environment: 'integration'
      };

      const resource = constructIAMResourceFromConfig(
        resourceTransformConfig,
        request
      );

      Code.expect(resource).to.equal(expectedResult);
    }
  );

  lab.test(
    'should construct the resource object for the given transformConfig and request even if iamKey is abset',
    async () => {
      const request = {
        query: {
          entity: 'platform',
          landscape: 'id'
        },
        params: {
          name: 'p-firehose-123'
        },
        payload: {
          environment: 'integration'
        }
      };

      const resourceTransformConfig = [
        {
          requestKey: 'entity'
        },
        {
          requestKey: 'name'
        },
        {
          requestKey: 'environment'
        }
      ];

      const expectedResult = {
        entity: 'platform',
        name: 'p-firehose-123',
        environment: 'integration'
      };

      const resource = constructIAMResourceFromConfig(
        resourceTransformConfig,
        request
      );

      Code.expect(resource).to.equal(expectedResult);
    }
  );

  lab.test(
    'should skip keys that are not present in request data',
    async () => {
      const request = {
        query: {
          entity: 'platform'
        },
        params: {
          name: 'p-firehose-123'
        }
      };

      const resourceTransformConfig = [
        {
          requestKey: 'entity',
          iamKey: 'newEntity'
        },
        {
          requestKey: 'name',
          iamKey: 'resource'
        },
        {
          requestKey: 'environment',
          iamKey: 'environment'
        }
      ];

      const expectedResult = {
        newEntity: 'platform',
        resource: 'p-firehose-123'
      };

      const resource = constructIAMResourceFromConfig(
        resourceTransformConfig,
        request
      );

      Code.expect(resource).to.equal(expectedResult);
    }
  );
});

lab.experiment('IAMPluginUtils::getIAMAction', () => {
  let contructIAMActionFromConfigStub;

  lab.before(() => {
    contructIAMActionFromConfigStub = Sandbox.stub(
      IAMPluginUtils,
      'contructIAMActionFromConfig'
    ).returns('firehose.get');
  });

  lab.test('should getIAMAction if action type is string ', () => {
    const action = 'firehose.get';

    const resultAction = getIAMAction(action, 'GET');

    Code.expect(resultAction).to.equal(action);
    Sandbox.assert.notCalled(contructIAMActionFromConfigStub);
  });

  lab.test('should getIAMAction if action type is ActionConfig', () => {
    const action = {
      baseName: 'firehose'
    };
    const method = 'GET';

    const resultAction = getIAMAction(action, 'GET');

    Code.expect(resultAction).to.equal('firehose.get');
    Sandbox.assert.calledWithExactly(
      contructIAMActionFromConfigStub,
      action,
      method
    );
  });
});

lab.experiment('IAMPluginUtils::constructIAMActionFromConfig', () => {
  lab.test(
    'should constructIAMActionFromConfig if baseName is given and operation is not given',
    () => {
      const action = {
        baseName: 'Firehose'
      };
      const method = 'POST';

      const resultAction = contructIAMActionFromConfig(action, method);
      const expectedAction = 'firehose.create';
      Code.expect(resultAction).to.equal(expectedAction);
    }
  );

  lab.test(
    'should constructIAMActionFromConfig if both baseName and operation is given',
    () => {
      const action = {
        baseName: 'firehose',
        operation: 'create'
      };
      const method = 'POST';

      const resultAction = contructIAMActionFromConfig(action, method);
      const expectedAction = 'firehose.create';
      Code.expect(resultAction).to.equal(expectedAction);
    }
  );

  lab.test('should return manage operation for PUT request method', () => {
    const action = {
      baseName: 'firehose'
    };
    const method = 'PUT';

    const resultAction = contructIAMActionFromConfig(action, method);
    const expectedAction = 'firehose.manage';
    Code.expect(resultAction).to.equal(expectedAction);
  });
});
