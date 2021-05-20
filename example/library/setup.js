/* eslint-disable no-console */
/* eslint-disable @typescript-eslint/ban-ts-comment */
// @ts-nocheck

// ROLE setup
const R = require('ramda');
const Wreck = require('wreck');
const { Client } = require('pg');
const printMessage = require('print-message');
const checkAccess = require('./check');

const dbClient = new Client({
  connectionString: 'postgresql://shield_library@postgres:5432/shield_library'
});

const SHIELD_URL = 'http://shield:5000/api';

const SUPER_ADMIN_EMAIL = 'admin@library.com';

const defaultHeaders = {
  'X-Goog-Authenticated-User-Email': SUPER_ADMIN_EMAIL
};

const PRINT_STACK = [];

const getActionsPayload = R.map((action) => ({
  action,
  operation: 'create'
}));

const createRoles = async () => {
  const rolesPayload = [
    {
      displayname: 'Team Admin',
      actions: getActionsPayload(['*']),
      attributes: ['group']
    },
    {
      displayname: 'Book Manager',
      actions: getActionsPayload([
        'book.read',
        'book.create',
        'book.update',
        'book.delete'
      ]),
      attributes: ['group', 'category']
    }
  ];

  const promiseList = rolesPayload.map((role) => {
    return Wreck.request('POST', `${SHIELD_URL}/roles`, {
      headers: defaultHeaders,
      payload: role
    });
  });

  await Promise.all(promiseList);

  rolesPayload.forEach((r) =>
    PRINT_STACK.push(`${r.displayname} role was created`)
  );
  PRINT_STACK.push('');
};

// create super admin
const createSuperAdmin = async () => {
  // find the user with SUPER_ADMIN_EMAIL
  // set admin policy

  const { payload: users } = await Wreck.get(`${SHIELD_URL}/users`, {
    headers: defaultHeaders,
    json: true
  });
  const admin = users.find((user) => user.displayname === 'admin');

  const text =
    'INSERT INTO casbin_rule(ptype, v0, v1, v2) VALUES($1, $2, $3, $4) RETURNING *';
  const values = [
    'p',
    JSON.stringify({ user: admin.id }),
    JSON.stringify({ '*': '*' }),
    JSON.stringify({ action: '*' })
  ];
  await dbClient.query(text, values);

  PRINT_STACK.push(`${SUPER_ADMIN_EMAIL} was given super admin privileges`);
  PRINT_STACK.push('');
};

const createUsers = async () => {
  const users = [
    {
      displayname: 'Einstein',
      metadata: {
        email: 'einstein@library.com'
      }
    },
    {
      displayname: 'Newton',
      metadata: {
        email: 'newton@library.com'
      }
    },
    {
      displayname: 'Darwin',
      metadata: {
        email: 'darwin@library.com'
      }
    }
  ];

  const promiseList = users.map((user) => {
    return Wreck.request('POST', `${SHIELD_URL}/users`, {
      headers: defaultHeaders,
      payload: user
    });
  });

  await Promise.all(promiseList);

  users.forEach((u) => PRINT_STACK.push(`${u.displayname} was added`));
  PRINT_STACK.push('');
};

const createGroups = async () => {
  const groups = [
    {
      displayname: 'Scientists',
      metadata: {
        email: 'scientists@library.com'
      }
    },
    {
      displayname: 'Mathematicians',
      metadata: {
        email: 'mathematicians@library.com'
      }
    }
  ];

  const promiseList = groups.map((group) => {
    return Wreck.request('POST', `${SHIELD_URL}/groups`, {
      headers: defaultHeaders,
      payload: group
    });
  });

  await Promise.all(promiseList);

  groups.forEach((g) => PRINT_STACK.push(`${g.displayname} group was added`));
  PRINT_STACK.push('');
};

const createResources = async () => {
  const { payload: groups } = await Wreck.get(`${SHIELD_URL}/groups`, {
    headers: defaultHeaders,
    json: true
  });
  const scientistGroup = R.find(R.propEq('displayname', 'Scientists'))(groups);
  const resourceUrn = 'relativity-the-special-general-theory';
  const resourcesPayload = [
    {
      operation: 'create',
      resource: { urn: resourceUrn },
      attributes: { category: 'physics', group: scientistGroup.id }
    }
  ];

  await Wreck.request('POST', `${SHIELD_URL}/resources`, {
    headers: defaultHeaders,
    payload: resourcesPayload
  });

  PRINT_STACK.push(
    `${resourceUrn} resource was mapped with ${JSON.stringify({
      category: 'physics',
      group: scientistGroup.id
    })}`
  );
  PRINT_STACK.push('');
};

// upsert policies
const addUsersToGroups = async () => {
  const { payload: roles } = await Wreck.get(`${SHIELD_URL}/roles`, {
    headers: defaultHeaders,
    json: true
  });
  const { payload: groups } = await Wreck.get(`${SHIELD_URL}/groups`, {
    headers: defaultHeaders,
    json: true
  });

  const { payload: users } = await Wreck.get(`${SHIELD_URL}/users`, {
    headers: defaultHeaders,
    json: true
  });

  const bookManagerRole = roles.find((r) => r.displayname === 'Book Manager');
  const teamAdminRole = roles.find((r) => r.displayname === 'Team Admin');
  const scientistGroup = groups.find((g) => g.displayname === 'Scientists');
  const einstein = users.find((u) => u.displayname === 'Einstein');
  const newton = users.find((u) => u.displayname === 'Newton');
  const darwin = users.find((u) => u.displayname === 'Darwin');

  const usersGroups = [
    {
      userId: einstein.id,
      groupId: scientistGroup.id,
      policies: [
        {
          operation: 'create',
          subject: {
            user: einstein.id
          },
          resource: {
            category: 'physics',
            group: scientistGroup.id
          },
          action: {
            role: bookManagerRole.id
          }
        }
      ]
    },
    {
      userId: darwin.id,
      groupId: scientistGroup.id,
      policies: [
        {
          operation: 'create',
          subject: {
            user: einstein.id
          },
          resource: {
            category: 'biology',
            group: scientistGroup.id
          },
          action: {
            role: bookManagerRole.id
          }
        }
      ]
    },
    {
      userId: newton.id,
      groupId: scientistGroup.id,
      policies: [
        {
          operation: 'create',
          subject: {
            user: einstein.id
          },
          resource: {
            group: scientistGroup.id
          },
          action: {
            role: teamAdminRole.id
          }
        }
      ]
    }
  ];

  const promiseList = usersGroups.map((ug) => {
    return Wreck.request(
      'POST',
      `${SHIELD_URL}/groups/${ug.groupId}/users/${ug.userId}`,
      {
        headers: defaultHeaders,
        payload: { policies: ug.policies }
      }
    );
  });

  await Promise.all(promiseList);

  usersGroups.forEach((ug) => {
    const user = users.find((u) => u.id === ug.userId);
    const group = groups.find((g) => g.id === ug.groupId);
    const role = roles.find((r) => r.id === ug.policies[0].action.role);

    PRINT_STACK.push(
      `${user.displayname} was given ${role.displayname} role in ${group.displayname} group`
    );
  });
  PRINT_STACK.push('');
};

// setup
const setup = async () => {
  try {
    await dbClient.connect();
    await dbClient.query('TRUNCATE TABLE users, groups, roles, casbin_rule');

    await createSuperAdmin();
    await createUsers();
    await createGroups();
    await createRoles();
    await createResources();
    await addUsersToGroups();

    printMessage(['SETTING UP LIBRARY DATA', '\n', ...PRINT_STACK]);

    checkAccess();
  } catch (e) {
    console.log('e', e);
  }
};

setup();
