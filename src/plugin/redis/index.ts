import redis, { ClientOpts } from 'redis';
import Hapi from '@hapi/hapi';

export class RedisConnection {
  client: redis.RedisClient;
  constructor(options: ClientOpts) {
    this.client = redis.createClient(options);
  }

  getKeys = (pattern: string): Promise<string[]> =>
    new Promise((resolve, reject) => {
      this.client.keys(pattern, function (err, keys) {
        if (err) {
          reject(err);
        } else {
          resolve(keys);
        }
      });
    });
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
      const { client, getKeys } = new RedisConnection({ url });
      client.on('ready', function () {
        server.log('info', 'Connected to redis!');
      });
      client.on('error', (err) => {
        server.log('error', 'Redis Error ' + err);
      });
      server.expose('getKeys', getKeys);
      server.expose('client', client);
    }
  }
};
