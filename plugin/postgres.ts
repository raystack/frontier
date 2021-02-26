/* eslint-disable no-use-before-define */
import 'reflect-metadata';
import { Connection, createConnection } from 'typeorm';
import Hapi from '@hapi/hapi';
import Logger from '../lib/logger';

export interface ConnectionConfig {
  uri: string;
  options?: any;
}

// eslint-disable-next-line import/prefer-default-export
export const plugin = {
  name: 'postgres',
  async register(server: Hapi.Server) {
    const tryConnectToPostgres: () => Promise<Connection> = async () => {
      try {
        return createConnection();
      } catch (e) {
        return postgresConnectionErrorHandler(e);
      }
    };

    const postgresConnectionErrorHandler: (
      err: any
    ) => Promise<Connection> = async (err: any) => {
      return new Promise((resolve) => {
        Logger.error(
          `Failed to connect to postgres on start up - retrying in 5 sec : ${err}`
        );
        setTimeout(async () => {
          resolve(await tryConnectToPostgres());
        }, 5000);
      });
    };

    const connection: Connection = await tryConnectToPostgres();
    await connection.runMigrations();

    server.events.on('stop', async () => {
      Logger.info('Closing Postgres connections on server stop');
      await connection.close();
    });
  }
};
