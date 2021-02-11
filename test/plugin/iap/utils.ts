import Code from 'code';
import Lab from '@hapi/lab';
import { lab } from '../../setup';
import * as IAPPluginUtils from '../../../plugin/iap/utils';

exports.lab = Lab.script();

lab.experiment('IAPPluginUtils::getUsernameFromEmail', () => {
  const email = 'demo@demo.com';
  const username = 'demo';

  lab.test('should get username from email if we provide email', () => {
    const resultAction = IAPPluginUtils.getUsernameFromEmail(email);
    Code.expect(resultAction).to.equal(username);
  });

  lab.test(
    'should get full name if we provide any string but not type of email',
    () => {
      const notAEmail = 'demoemail';
      const resultAction = IAPPluginUtils.getUsernameFromEmail(notAEmail);
      Code.expect(resultAction).to.equal(notAEmail);
    }
  );
});

lab.experiment('IAPPluginUtils::getEmailFromIAPHeader', () => {
  lab.test('should get email from IAP header', () => {
    const emailFromIAPHeader = 'demo@demo.com';
    const resultAction = IAPPluginUtils.getEmailFromIAPHeader(
      emailFromIAPHeader
    );
    Code.expect(resultAction).to.equal(emailFromIAPHeader);
  });

  lab.test(
    'should get email from IAP header even it contains `accounts.google.com`',
    () => {
      const emailFromIAPHeader = 'accounts.google.com:demo@demo.com';
      const expected = 'demo@demo.com';

      const resultAction = IAPPluginUtils.getEmailFromIAPHeader(
        emailFromIAPHeader
      );
      Code.expect(resultAction).to.equal(expected);
    }
  );
});
