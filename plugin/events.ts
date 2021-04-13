import Hapi from '@hapi/hapi';
import newrelic from 'newrelic';
import Boom from '@hapi/boom';

exports.plugin = {
  name: 'events',
  version: '1.0.0',
  async register(server: Hapi.Server) {
    server.ext({
      type: 'onPreResponse',
      async method(request: Hapi.Request, h) {
        const { response } = request;
        if (Boom.isBoom(response)) {
          newrelic.noticeError(response);
        }
        return h.continue;
      }
    });
  }
};
