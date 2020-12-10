/* eslint-disable no-use-before-define */
import 'reflect-metadata';
import { Connection, createConnection } from 'typeorm';
import Hoek from '@hapi/hoek';
import Hapi from '@hapi/hapi';
import Logger from '../lib/logger';

interface ConnectionConfig {
  uri: string;
  options?: any;
}

type NullOrConnection = Connection | null;

// eslint-disable-next-line import/prefer-default-export
export const plugin = {
  name: 'postgres',
  version: '1.0.0',
  async register(server: Hapi.Server, options: ConnectionConfig) {
    const tryConnectToPostgres: () => Promise<NullOrConnection> = async () => {
      try {
        return createConnection({
          type: 'postgres',
          url: options.uri,
          entities: [],
          logging: true
        });
      } catch (e) {
        return postgresConnectionErrorHandler(e);
      }
    };

    const postgresConnectionErrorHandler: (
      err: any
    ) => Promise<NullOrConnection> = async (err: any) => {
      return new Promise((resolve) => {
        if (err) {
          Logger.error(
            'Failed to connect to postgres on start up - retrying in 5 sec'
          );
          setTimeout(async () => {
            resolve(await tryConnectToPostgres());
          }, 5000);
        }
        Hoek.assert(!err, err);
      });
    };

    const connection: NullOrConnection = await tryConnectToPostgres();

    server.events.on('stop', async () => {
      Logger.info('Closing Postgres connections on server stop');
      if (connection && connection.close) {
        await connection.close();
      }
    });
  }
};
