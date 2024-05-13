/* eslint-disable strict */
module.exports = {
  plugins: ["test-selectors"],
  extends: [
    'next',
    'turbo',
    'prettier',
    "eslint:recommended",
    "plugin:react/recommended",
    "plugin:react-hooks/recommended",
    'plugin:test-selectors/recommended'
  ],
  rules: {
    '@next/next/no-html-link-for-pages': 'off'
  },
  parserOptions: {
    babelOptions: {
      presets: [require.resolve('next/babel')]
    }
  }
};
