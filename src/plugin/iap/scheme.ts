import Hapi from '@hapi/hapi';
import Boom from '@hapi/boom';
import { getEmailFromIAPHeader } from './utils';

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

export default scheme;
