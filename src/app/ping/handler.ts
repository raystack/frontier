import Hapi from '@hapi/hapi';
import Joi from 'joi';

// TODO: Need to investigate why typescript is not auto detecting auth value in the object
const auth: Hapi.RouteOptionsAccess = {
  mode: 'optional'
};

export const ping = {
  description: 'pong the request',
  tags: ['api', 'ping'],
  response: {
    status: {
      200: Joi.object()
        .keys({
          statusCode: Joi.number().integer().valid(200).required(),
          status: Joi.string().required().valid('ok'),
          message: Joi.string().required()
        })
        .label('PingPong')
    }
  },
  auth,
  handler: () => {
    return {
      statusCode: 200,
      status: 'ok',
      message: 'pong'
    };
  }
};
