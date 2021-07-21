const { SnakeNamingStrategy } = require('typeorm-naming-strategies');
const Config = require('./src/config/config');

const baseDir = Config.get('/typeormDir').dir;

module.exports = {
  type: 'postgres',
  url: Config.get('/postgres').uri,
  logging: false,
  synchronize: false,
  entities: [`${baseDir}/src/model/*{.ts,.js}`],
  migrations: [`${baseDir}/src/migration/*{.ts,.js}`],
  factories: [`${baseDir}/src/factory/*{.ts,.js}`],
  subscribers: [`${baseDir}/src/subscriber/*{.ts,.js}`],
  cli: {
    migrationsDir: 'src/migration'
  },
  namingStrategy: new SnakeNamingStrategy()
};
