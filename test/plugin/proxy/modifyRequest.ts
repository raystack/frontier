import Code from 'code';
import Lab from '@hapi/lab';
import Sinon from 'sinon';
import { factory } from 'typeorm-seeding';
import { lab } from '../../setup';
import * as ModifyRequest from '../../../plugin/proxy/modifyRequest';
import { User } from '../../../model/user';
import { Group } from '../../../model/group';

const {
  checkAndAppendUserData,
  checkAndAppendGroupData,
  getModifiedPayload,
  getModifiedHeaders,
  default: modifyRequest
} = ModifyRequest;

exports.lab = Lab.script();

const Sandbox = Sinon.createSandbox();

lab.afterEach(() => {
  Sandbox.restore();
});

lab.experiment('ModifyRequest', () => {
  lab.experiment('modifyRequest', () => {
    let getModifiedHeadersStub, getModifiedPayloadStub;

    const request = {
      payload: Buffer.from(JSON.stringify({ user_id: 1, group_id: 1 }), 'utf8')
    };

    lab.before(() => {
      getModifiedHeadersStub = Sandbox.stub(
        ModifyRequest,
        'getModifiedHeaders'
      ).resolves(request);
      getModifiedPayloadStub = Sandbox.stub(
        ModifyRequest,
        'getModifiedPayload'
      ).resolves(request);
    });

    lab.test('should return proxy options', async () => {
      await modifyRequest(request, <any>{ continue: true });
      Sandbox.assert.calledWithExactly(getModifiedPayloadStub, request);
      Sandbox.assert.calledWithExactly(getModifiedHeadersStub, request);
    });
  });

  lab.experiment('checkAndAppendUserData', () => {
    let user;
    lab.before(async () => {
      user = await factory(User)().create();
    });

    lab.test('should append if user is authenticated', async () => {
      const request = <any>{
        auth: {
          credentials: user
        }
      };
      const payload = {
        test: '123'
      };
      const expetedResult = {
        ...payload,
        created_by: user.username,
        user
      };
      const modifiedPayload = await checkAndAppendUserData(request, payload);
      Code.expect(modifiedPayload).to.equal(expetedResult);
    });

    lab.test('should not append if user is not authenticated', async () => {
      const payload = {
        test: '123'
      };
      const modifiedPayload = await checkAndAppendUserData(<any>{}, payload);
      Code.expect(modifiedPayload).to.equal(payload);
    });
  });

  lab.experiment('checkAndAppendGroupData', () => {
    let group;
    lab.before(async () => {
      group = await factory(Group)().create();
    });

    lab.test('should append group if group_id is present', async () => {
      const payload = { group_id: group.id, test: '233' };
      const expetedResult = {
        test: '233',
        group_id: group.id,
        group,
        team: group.groupname
      };
      const result = await checkAndAppendGroupData(payload);
      Code.expect(result).to.equal(expetedResult);
    });

    lab.test('should not append group if group_id is present', async () => {
      const body = {};
      const result = await checkAndAppendGroupData(body);
      Code.expect(result.group).to.be.undefined();
    });

    lab.test('should throw error if group not found', async () => {
      const payload = { group_id: group.id, test: '233' };
      let exception = null;
      try {
        await checkAndAppendGroupData(payload);
      } catch (e) {
        exception = e;
      }
      Code.expect(exception).to.not.be.undefined();
    });
  });

  lab.experiment('getModifiedHeaders', () => {
    lab.test(
      'should append username in header if google iap header is given',
      async () => {
        const request = <any>{
          auth: {
            credentials: {
              username: 'test'
            }
          },
          headers: {
            test: 123
          }
        };
        const result: any = await getModifiedHeaders(request);
        const expectedResult = { ...request.headers, username: 'test' };
        Code.expect(result).to.equal(expectedResult);
      }
    );

    lab.test(
      'should not append username in header if google iap header is not given',
      async () => {
        const request = <any>{
          headers: {
            test: 123
          }
        };
        const result: any = await getModifiedHeaders(request);
        Code.expect(result).to.equal(request.headers);
      }
    );
  });

  lab.experiment('getModifiedPayload', () => {
    let checkAndAppendUserDataStub, checkAndAppendGroupDataStub;
    const payload = { user_id: 1, group_id: 1 };
    const request = <any>{
      payload: Buffer.from(JSON.stringify(payload), 'utf8')
    };

    lab.before(() => {
      checkAndAppendUserDataStub = Sandbox.stub(
        ModifyRequest,
        'checkAndAppendUserData'
      ).returns(payload);
      checkAndAppendGroupDataStub = Sandbox.stub(
        ModifyRequest,
        'checkAndAppendGroupData'
      ).resolves(payload);
    });

    lab.test('should appendProxyPayload to options', async () => {
      const result = await getModifiedPayload(request);
      Code.expect(result).to.equal(payload);
      Sandbox.assert.calledWithExactly(
        checkAndAppendUserDataStub,
        request,
        payload
      );
      Sandbox.assert.calledWithExactly(checkAndAppendGroupDataStub, payload);
    });
  });
});
