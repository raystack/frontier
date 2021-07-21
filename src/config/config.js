/* eslint-disable @typescript-eslint/ban-ts-comment */
const Confidence = require('confidence');

const internals = {
  criteria: {
    env: process.env.NODE_ENV,
    ci: process.env.CI
  }
};

// @ts-ignore
internals.config = {
  $meta: 'App configuration',
  env: {
    $filter: 'env',
    production: 'production',
    integration: 'integration',
    test: 'test',
    $default: 'dev'
  },
  port: {
    web: {
      $filter: 'env',
      test: 9000,
      production: process.env.PORT,
      integration: process.env.PORT,
      $default: process.env.PORT || 8080
    }
  },
  postgres: {
    $filter: 'env',
    test: {
      $filter: 'ci',
      gitlab: {
        uri: 'postgresql://shield_test@localhost:4322/shield_test'
      },
      $default: {
        uri: 'postgresql://shield_test@localhost:4322/shield_test'
      }
    },
    'test-migrate': {
      $filter: 'ci',
      gitlab: {
        uri: 'postgresql://shield_test@localhost:4322/shield_test'
      },
      $default: {
        uri: 'postgresql://shield_test@localhost:4322/shield_test'
      }
    },
    $default: {
      uri: process.env.DB_URI,
      options: {}
    }
  },
  typeormDir: {
    $filter: 'env',
    test: {
      dir: './'
    },
    dev: {
      dir: './'
    },
    'test-migrate': {
      dir: './build'
    },
    $default: {
      dir: './build'
    }
  },
  environment: {
    $filter: 'env',
    test: {
      name: 'local',
      prefix: 't'
    },
    $default: {
      name: { $env: 'NODE_ENV', $default: 'local' }
    }
  },
  new_relic: {
    APP_NAME: { $env: 'APP_NAME' },
    KEY: { $env: 'NEW_RELIC_KEY' },
    enabled: {
      $filter: 'env',
      test: 'false',
      $default: { $env: 'NEW_RELIC_ENABLED', $default: 'false' }
    }
  },

  // Joi validation options
  validationOptions: {
    abortEarly: false, // abort after the last validation error
    stripUnknown: true // remove unknown keys from the validated data
  }
};

// @ts-ignore
internals.store = new Confidence.Store(internals.config);

// @ts-ignore
const get = function (key) {
  // @ts-ignore
  return internals.store.get(key, internals.criteria);
};

module.exports = { get };
