const Config = require('./build/config/config');

module.exports = {
  type: 'postgres',
  url: Config.get('/postgres').uri,
  logging: true,
  synchronize: false,
  entities: ['build/model/*.js'],
  migrations: ['build/migration/*.js'],
  cli: {
    migrationsDir: 'migration'
  }
};
