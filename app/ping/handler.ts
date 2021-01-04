import Hapi from '@hapi/hapi';

// TODO: Need to investigate why typescript is not auto detecting auth value in the object
const auth: Hapi.RouteOptionsAccess = {
  mode: 'optional'
};

export const ping = {
  description: 'pong the request',
  tags: ['api'],
  auth,
  handler: () => {
    return {
      statusCode: 200,
      status: 'ok',
      message: 'pong'
    };
  }
};
