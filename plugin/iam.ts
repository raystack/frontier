import Hapi from '@hapi/hapi';

// console.log(
//   'route',
//   route.settings.app,
//   request.auth.credentials,
//   request.query,
//   request.params,
//   request.payload
// );

export const plugin = {
  name: 'iam',
  dependencies: ['postgres', 'iap'],
  async register(server: Hapi.Server) {
    // initialize casbin

    // listen on request event
    server.ext({
      type: 'onPreHandler',
      method(request, h) {
        // const route = server.match(request.method, request.path);
        // TODO: Hook casbin here
        // if (route) {

        // }
        return h.continue;
      }
    });
  }
};
