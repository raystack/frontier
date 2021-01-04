import Hapi from '@hapi/hapi';
import Boom from '@hapi/boom';

const getEmailFromIAPHeader = (header: string) => {
  return header?.replace('accounts.google.com:', '');
};

const getUsernameFromEmail = (email: string) => {
  return email?.split('@')?.shift();
};

const scheme = (server: Hapi.Server, options: any) => {
  return {
    async authenticate(request: Hapi.Request, h: Hapi.ResponseToolkit) {
      const googleIAPHeader =
        request.headers['X-Goog-Authenticated-User-Email'] ||
        request.headers['x-goog-authenticated-user-email'];

      if (!googleIAPHeader) {
        throw Boom.unauthorized(null, 'IAP');
      }

      const email = getEmailFromIAPHeader(googleIAPHeader);

      if (!email) {
        throw Boom.unauthorized('Email is required', 'IAP');
      }

      const { isValid, credentials } = await options.validate(
        request,
        email,
        h
      );

      if (!isValid) {
        throw Boom.unauthorized('Could not authenticate the user', 'IAP');
      }

      return h.authenticated({ credentials });
    }
  };
};

const validate = async (request: Hapi.Request, email: string) => {
  // TODO: fetch user from db using username and upsert and validate the user
  const username = getUsernameFromEmail(email);

  const credentials = { username, email };

  return { isValid: true, credentials };
};

export const plugin = {
  name: 'iap',
  dependencies: ['postgres'],
  async register(server: Hapi.Server) {
    server.auth.scheme('IAP', scheme);
    server.auth.strategy('simple', 'IAP', { validate });
    server.auth.default('simple');
  }
};
