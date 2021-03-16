const { SnakeNamingStrategy } = require('typeorm-naming-strategies');
const Config = require('./config/config');

const baseDir = Config.get('/typeormDir').dir;

module.exports = {
  type: 'postgres',
  url: Config.get('/postgres').uri,
  logging: false,
  synchronize: false,
  entities: [`${baseDir}/model/*{.ts,.js}`],
  migrations: [`${baseDir}/migration/*{.ts,.js}`],
  factories: [`${baseDir}/factory/*{.ts,.js}`],
  subscribers: [`${baseDir}/subscriber/*{.ts,.js}`],
  cli: {
    migrationsDir: 'migration'
  },
  namingStrategy: new SnakeNamingStrategy()
};
