/* eslint-disable strict */
module.exports = {
  plugins: ["test-selectors"],
  extends: [
    'next',
    'turbo',
    'prettier',
    "eslint:recommended",
    'plugin:test-selectors/recommended'
  ],
  rules: {
    '@next/next/no-html-link-for-pages': 'off',
    "no-unused-vars": "warn"
  },
  parserOptions: {
    babelOptions: {
      presets: [require.resolve('next/babel')]
    }
  }
};
