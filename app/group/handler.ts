import Hapi from '@hapi/hapi';
import * as Schema from './schema';
import * as Resource from './resource';

const ACTION_BASE_NAME = 'iam';

export const list = {
  description: 'get list of groups',
  tags: ['api'],
  handler: async (request: Hapi.Request) => {
    return Resource.list(request.query);
  }
};

export const get = {
  description: 'get group by id',
  tags: ['api'],
  handler: async (request: Hapi.Request) => {
    return Resource.get(request.params.id, request.query);
  }
};

export const create = {
  description: 'create group',
  tags: ['api'],
  app: {
    iam: {
      authorize: [
        {
          action: { baseName: ACTION_BASE_NAME },
          resource: [{ requestKey: 'entity' }]
        }
      ],
      manage: {
        upsert: [
          {
            resource: [{ requestKey: 'name', iamKey: 'group' }],
            resourceAttributes: [{ requestKey: 'entity' }]
          }
        ]
      }
    }
  },
  validate: {
    payload: Schema.createPayload
  },
  handler: async (request: Hapi.Request) => {
    const { username } = request?.auth?.credentials || {};
    return Resource.create(request.payload, { username });
  }
};

export const update = {
  description: 'update group by id',
  tags: ['api'],
  app: {
    iam: {
      authorize: [
        {
          action: { baseName: ACTION_BASE_NAME },
          resource: [{ requestKey: 'name', iamKey: 'group' }]
        },
        {
          action: { baseName: ACTION_BASE_NAME },
          resource: [{ requestKey: 'entity' }]
        }
      ],
      manage: {
        upsert: [
          {
            resource: [{ requestKey: 'name', iamKey: 'group' }],
            resourceAttributes: [{ requestKey: 'entity' }]
          }
        ]
      }
    }
  },
  validate: {
    payload: Schema.updatePayload
  },
  handler: async (request: Hapi.Request) => {
    const { username } = request?.auth?.credentials || {};
    return Resource.update(request.params.id, request.payload, { username });
  }
};
