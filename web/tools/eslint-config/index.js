/* eslint-disable strict */
const extendsList = [
  'next',
  ...(process.env.ESLINT_SKIP_TURBO === '1' ? [] : ['turbo']),
  'prettier',
  "eslint:recommended",
  'plugin:test-selectors/recommended'
];

module.exports = {
  plugins: ["test-selectors"],
  extends: extendsList,
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
