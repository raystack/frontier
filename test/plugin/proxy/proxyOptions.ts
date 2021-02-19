import Code from 'code';
import Lab from '@hapi/lab';
import Wreck from '@hapi/wreck';
import Sinon from 'sinon';
import { factory } from 'typeorm-seeding';
import * as R from 'ramda';
import { lab } from '../../setup';
import * as ProxyOptions from '../../../plugin/proxy/proxyOptions';
import { User } from '../../../model/user';
import { Group } from '../../../model/group';

const {
  checkAndAppendUserData,
  checkAndAppendGroupData,
  appendProxyPayload,
  appendHeaders,
  default: getProxyOptions
} = ProxyOptions;

exports.lab = Lab.script();

const Sandbox = Sinon.createSandbox();

lab.afterEach(() => {
  Sandbox.restore();
});

lab.experiment('ProxyOptions', () => {
  lab.experiment('getProxyOptions', () => {
    let appendHeadersStub, appendProxyPayloadStub;

    const options = {
      payload: { user_id: 1, group_id: 1 }
    };

    lab.before(() => {
      appendHeadersStub = Sandbox.stub(ProxyOptions, 'appendHeaders').resolves(
        options
      );
      appendProxyPayloadStub = Sandbox.stub(
        ProxyOptions,
        'appendProxyPayload'
      ).resolves(options);
    });

    lab.test('should return proxy options', async () => {
      await getProxyOptions(options);
      Sandbox.assert.calledWithExactly(appendProxyPayloadStub, options);
      Sandbox.assert.calledWithExactly(appendHeadersStub, options);
    });
  });

  lab.experiment('checkAndAppendUserData', () => {
    let user;
    lab.before(async () => {
      user = await factory(User)().create();
    });

    lab.test('should append user if user_id is present', async () => {
      const body = { user_id: user.id };
      const expetedResult = {
        user_id: user.id,
        user
      };
      const modifiedBody = await checkAndAppendUserData(body);
      Code.expect(modifiedBody).to.equal(expetedResult);
    });

    lab.test('should not append user if user_id is present', async () => {
      const body = {};
      const modifiedBody = await checkAndAppendUserData(body);
      Code.expect(modifiedBody.user).to.be.undefined();
    });
  });

  lab.experiment('checkAndAppendGroupData', () => {
    let group;
    lab.before(async () => {
      group = await factory(Group)().create();
    });

    lab.test('should append group if group_id is present', async () => {
      const body = { group_id: group.id };
      const expetedResult = {
        group_id: group.id,
        group
      };
      const modifiedBody = await checkAndAppendGroupData(body);
      Code.expect(modifiedBody).to.equal(expetedResult);
    });

    lab.test('should not append group if group_id is present', async () => {
      const body = {};
      const modifiedBody = await checkAndAppendGroupData(body);
      Code.expect(modifiedBody.group).to.be.undefined();
    });
  });

  lab.experiment('appendHeaders', () => {
    lab.test(
      'should append username in header if google iap header is given',
      async () => {
        const options = {
          headers: {
            'X-Goog-Authenticated-User-Email':
              'accounts.google.com:shreyas.adiyodi@go-jek.com'
          }
        };
        const result = await appendHeaders(options);
        const expected = R.assocPath(
          ['headers', 'username'],
          'shreyas.adiyodi',
          options
        );
        Code.expect(result).to.equal(expected);
      }
    );

    lab.test(
      'should not append username in header if google iap header is not given',
      async () => {
        const result = await appendHeaders({});
        Code.expect(result).to.equal({});
      }
    );
  });

  lab.experiment('appendProxyPayload', () => {
    let wreck,
      checkAndAppendUserDataStub,
      checkAndAppendGroupDataStub,
      checkAndAppendOdinDataStub;
    const options = {
      payload: { user_id: 1, group_id: 1 }
    };

    lab.before(() => {
      wreck = Sandbox.stub(Wreck, 'read').resolves(options.payload);
      checkAndAppendUserDataStub = Sandbox.stub(
        ProxyOptions,
        'checkAndAppendUserData'
      ).resolves(options.payload);
      checkAndAppendGroupDataStub = Sandbox.stub(
        ProxyOptions,
        'checkAndAppendGroupData'
      ).resolves(options.payload);
      checkAndAppendOdinDataStub = Sandbox.stub(
        ProxyOptions,
        'checkAndAppendDataForOdin'
      ).resolves(options.payload);
    });

    lab.test('should appendProxyPayload to options', async () => {
      const result = await appendProxyPayload(options);
      Code.expect(result).to.equal(options);
      Sandbox.assert.calledWith(wreck, options.payload);
      Sandbox.assert.calledWithExactly(
        checkAndAppendUserDataStub,
        options.payload
      );
      Sandbox.assert.calledWithExactly(
        checkAndAppendGroupDataStub,
        options.payload
      );
      Sandbox.assert.calledWithExactly(
        checkAndAppendOdinDataStub,
        options.payload
      );
    });
  });
});
