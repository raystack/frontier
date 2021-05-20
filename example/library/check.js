/* eslint-disable no-console */
/* eslint-disable @typescript-eslint/ban-ts-comment */
// @ts-nocheck

const Wreck = require('wreck');
const printMessage = require('print-message');

const SUPER_ADMIN_EMAIL = 'admin@library.com';

const defaultHeaders = {
  'X-Goog-Authenticated-User-Email': SUPER_ADMIN_EMAIL
};

const SHIELD_URL = 'http://shield:5000/api';

const checkAccess = async () => {
  const { payload: users } = await Wreck.get(`${SHIELD_URL}/users`, {
    headers: defaultHeaders,
    json: true
  });

  const einstein = users.find((u) => u.displayname === 'Einstein');
  const darwin = users.find((u) => u.displayname === 'Darwin');

  const resourceUrn = 'relativity-the-special-general-theory';

  // check whether einstein can access
  await Wreck.put(`${SHIELD_URL}/books/${resourceUrn}`, {
    headers: {
      'X-Goog-Authenticated-User-Email': einstein.metadata.email
    },
    payload: {
      category: 'physics',
      description: 'A book by Einstein'
    }
  });

  try {
    // check whether darwin can access
    await Wreck.put(`${SHIELD_URL}/books/${resourceUrn}`, {
      headers: {
        'X-Goog-Authenticated-User-Email': darwin.metadata.email
      },
      payload: {
        category: 'biology',
        description: 'A book by Darwin'
      }
    });
    // eslint-disable-next-line no-empty
  } catch (e) {}

  printMessage([
    `CHECKING ACCESS`,
    '\n',
    `Einstein has book.update action access to ${resourceUrn}`,
    `Darwin does not have book.update action access to ${resourceUrn}`
  ]);
};

module.exports = checkAccess;
