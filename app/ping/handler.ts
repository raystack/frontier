export const ping = {
  description: 'pong the request',
  tags: ['api'],
  handler: () => {
    return {
      statusCode: 200,
      status: 'ok',
      message: 'pong'
    };
  }
};
