import redis, { ClientOpts } from 'redis';
import Hapi from '@hapi/hapi';

class RedisConnection {
  client: redis.RedisClient;
  constructor(options: ClientOpts) {
    this.client = redis.createClient(options);
  }
}

interface RedisConfig {
  url?: string;
  duration: number;
}

exports.plugin = {
  name: 'redis',
  version: '1.0.0',
  async register(server: Hapi.Server, options: RedisConfig) {
    const url = options?.url;
    if (url) {
      const redisClient = new RedisConnection({ url }).client;
      redisClient.on('ready', function () {
        server.log('info', 'Connected to redis!');
      });
      redisClient.on('error', (err) => {
        server.log('error', 'Redis Error ' + err);
      });
      server.expose('client', redisClient);
    }
  }
};
