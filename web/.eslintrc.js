module.exports = {
  root: true,
  extends: ['@raystack/eslint-config'],
  env: {
    es2020: true,
  },
  settings: {
    next: {
      rootDir: ['apps/*/']
    }
  }
};
