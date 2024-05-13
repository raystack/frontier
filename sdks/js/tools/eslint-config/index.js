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
    "no-unused-vars": "warn",
    "@next/next/no-img-element": "off"
  },
  parserOptions: {
    babelOptions: {
      presets: [require.resolve('next/babel')]
    }
  }
};
