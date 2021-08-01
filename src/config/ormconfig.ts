import path from 'path';
import { SnakeNamingStrategy } from 'typeorm-naming-strategies';
import { PostgresConnectionOptions } from 'typeorm/driver/postgres/PostgresConnectionOptions';
import * as Config from './config';

const baseDir = path.join(__dirname, '..');

const ormConfig: PostgresConnectionOptions = {
  type: 'postgres',
  url: Config.get('/postgres').uri,
  logging: false,
  synchronize: false,
  entities: [`${baseDir}/model/*{.ts,.js}`],
  migrations: [`${baseDir}/migration/*{.ts,.js}`],
  subscribers: [`${baseDir}/subscriber/*{.ts,.js}`],
  cli: {
    migrationsDir: 'src/migration'
  },
  namingStrategy: new SnakeNamingStrategy()
};

export default ormConfig;
