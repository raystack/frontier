import Hapi from '@hapi/hapi';
import iapScheme from './scheme';
import validateByEmail from './validate';

export const plugin = {
  name: 'iap',
  dependencies: ['postgres'],
  async register(server: Hapi.Server) {
    server.auth.scheme('IAP', iapScheme);
    server.auth.strategy('email', 'IAP', { validate: validateByEmail });
    server.auth.default('email');
  }
};
