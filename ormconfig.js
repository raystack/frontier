const Config = require('./build/config/config');

const baseDir = Config.get('/typeormDir').dir;

module.exports = {
  type: 'postgres',
  url: Config.get('/postgres').uri,
  logging: false,
  synchronize: false,
  entities: [`${baseDir}/model/*{.ts,.js}`],
  migrations: [`${baseDir}/migration/*{.ts,.js}`],
  factories: [`${baseDir}/factory/*{.ts,.js}`],
  cli: {
    migrationsDir: 'migration'
  }
};
