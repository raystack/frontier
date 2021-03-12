import Lab from '@hapi/lab';
import Sinon from 'sinon';
import Code from 'code';
import * as Faker from 'faker';
import { factory } from 'typeorm-seeding';
import { lab } from '../../setup';
import { Activity } from '../../../model/activity';
import * as Resource from '../../../app/activity/resource';
import { User } from '../../../model/user';
import { delta } from '../../../utils/deep-diff';

exports.lab = Lab.script();
const Sandbox = Sinon.createSandbox();

lab.afterEach(() => {
  Sandbox.restore();
});

lab.experiment('Activity::resource', () => {
  lab.experiment('create activity', () => {
    lab.afterEach(() => {
      Sandbox.restore();
    });

    lab.test('should create activity', async () => {
      const activity = {
        createdAt: 'Fri Mar 12 2021 13:24:53 GMT+0530 (India Standard Time)',
        diffs: [],
        document: {},
        documentId: 'ba256f55-bfce-4d17-8174-a0abbf26ccd4',
        model: 'User',
        title: 'Bypass Chips heuristic'
      };
      const activityId = Faker.random.uuid();
      const activitySaveStub = Sandbox.stub(Activity, 'save').returns(<any>{
        id: activityId,
        ...activity
      });
      const response = await Resource.create(activity);
      Sandbox.assert.calledWithExactly(activitySaveStub, <any>activity);
      Code.expect(response).to.equal({ id: activityId, ...activity });
    });
  });

  lab.experiment('log activity', () => {
    lab.afterEach(() => {
      Sandbox.restore();
    });

    lab.test(
      'should log activity by typeorm subscriber after insert event',
      async () => {
        const user = await factory(User)().create();
        const activityId = Faker.random.uuid();
        const event = {
          metadata: {
            tableName: 'groups'
          },
          entity: {
            id: activityId,
            displayname: 'Mock',
            first: 'First',
            second: 'Second',
            metadata: {
              key: 'value',
              number: 1,
              bool: true
            },
            createdAt: 'Fri Mar 12 2021 13:24:53 GMT+0530 (India Standard Time)'
          },
          queryRunner: {
            data: {
              user
            }
          }
        };

        const activity = {
          document: {},
          documentId: '0',
          diffs: delta({}, event.entity, {
            exclude: ['createdAt', 'updatedAt']
          }),
          title: 'Created Mock team ',
          createdBy: user
        };

        Sandbox.stub(Activity, 'create').returns(<any>{
          id: activityId,
          ...activity
        });
        await Resource.log(event, Resource.actions.CREATE);
      }
    );

    lab.test(
      'should log activity by typeorm subscriber after update event',
      async () => {
        const user = await factory(User)().create();
        const activityId = Faker.random.uuid();
        const event = {
          metadata: {
            tableName: 'groups'
          },
          databaseEntity: {
            id: activityId,
            displayname: 'Mock',
            first: 'First',
            second: 'Second',
            metadata: {
              key: 'value',
              number: 1,
              bool: true
            },
            createdAt: 'Fri Mar 12 2021 13:24:53 GMT+0530 (India Standard Time)'
          },
          entity: {
            id: activityId,
            displayname: 'Mock',
            first: 'First Edited',
            second: 'Second',
            metadata: {
              key: 'value edited',
              number: 2,
              bool: true
            },
            createdAt: 'Fri Mar 12 2021 13:24:53 GMT+0530 (India Standard Time)'
          },
          queryRunner: {
            data: {
              user
            }
          }
        };

        const activity = {
          document: {
            id: activityId,
            displayname: 'Mock',
            first: 'First',
            second: 'Second',
            metadata: {
              key: 'value',
              number: 1,
              bool: true
            },
            createdAt: 'Fri Mar 12 2021 13:24:53 GMT+0530 (India Standard Time)'
          },
          documentId: activityId,
          diffs: delta(event.databaseEntity, event.entity, {
            exclude: ['createdAt', 'updatedAt']
          }),
          title: 'Edited Mock team ',
          createdBy: user
        };

        Sandbox.stub(Activity, 'create').returns(<any>{
          id: activityId,
          ...activity
        });
        await Resource.log(event, Resource.actions.EDIT);
      }
    );

    lab.test(
      'should log activity by typeorm subscriber after remove event',
      async () => {
        const user = await factory(User)().create();
        const activityId = Faker.random.uuid();
        const event = {
          metadata: {
            tableName: 'groups'
          },
          databaseEntity: {
            id: activityId,
            displayname: 'Mock',
            first: 'First',
            second: 'Second',
            metadata: {
              key: 'value',
              number: 1,
              bool: true
            },
            createdAt: 'Fri Mar 12 2021 13:24:53 GMT+0530 (India Standard Time)'
          },
          entity: {},
          queryRunner: {
            data: {
              user
            }
          }
        };

        const activity = {
          document: {
            id: activityId,
            displayname: 'Mock',
            first: 'First',
            second: 'Second',
            metadata: {
              key: 'value',
              number: 1,
              bool: true
            },
            createdAt: 'Fri Mar 12 2021 13:24:53 GMT+0530 (India Standard Time)'
          },
          documentId: activityId,
          diffs: delta(event.databaseEntity, event.entity, {
            exclude: ['createdAt', 'updatedAt']
          }),
          title: 'Removed Mock team ',
          createdBy: user
        };

        Sandbox.stub(Activity, 'create').returns(<any>{
          id: activityId,
          ...activity
        });
        await Resource.log(event, Resource.actions.DELETE);
      }
    );
  });

  // lab.experiment('list users', () => {
  //   let users: any, groups, userEntityPolicy: any;
  //
  //   lab.beforeEach(async () => {
  //     // setup data
  //     const dbUri = Config.get('/postgres').uri;
  //     const enforcer = await CasbinSingleton.create(dbUri);
  //
  //     users = await factory(User)().createMany(5);
  //     groups = await factory(Group)().createMany(2);
  //     const user = users[0];
  //
  //     // user group mapping
  //     // eslint-disable-next-line @typescript-eslint/ban-ts-comment
  //     // @ts-ignore
  //     await enforcer.addSubjectGroupingJsonPolicy(
  //       { user: users[0].id },
  //       { group: groups[0].id },
  //       { created_by: user }
  //     );
  //
  //     // eslint-disable-next-line @typescript-eslint/ban-ts-comment
  //     // @ts-ignore
  //     await enforcer.addSubjectGroupingJsonPolicy(
  //       { user: users[1].id },
  //       { group: groups[0].id },
  //       { created_by: user }
  //     );
  //
  //     // eslint-disable-next-line @typescript-eslint/ban-ts-comment
  //     // @ts-ignore
  //     await enforcer.addSubjectGroupingJsonPolicy(
  //       { user: users[3].id },
  //       { group: groups[1].id },
  //       { created_by: user }
  //     );
  //
  //     // eslint-disable-next-line @typescript-eslint/ban-ts-comment
  //     // @ts-ignore
  //     await enforcer?.addSubjectGroupingJsonPolicy(
  //       { user: users[2].id },
  //       { group: groups[1].id },
  //       { created_by: user }
  //     );
  //
  //     // create relavant policies
  //     await enforcer?.addJsonPolicy(
  //       { group: groups[0].id },
  //       { entity: 'gojek', privacy: 'public' },
  //       { action: 'firehose.read' },
  //       { created_by: user }
  //     );
  //     userEntityPolicy = {
  //       subject: { user: users[2].id },
  //       resource: { entity: 'gojek' },
  //       action: { action: 'firehose.read' }
  //     };
  //     await enforcer?.addJsonPolicy(
  //       userEntityPolicy.subject,
  //       userEntityPolicy.resource,
  //       userEntityPolicy.action,
  //       { created_by: user }
  //     );
  //     await enforcer?.addJsonPolicy(
  //       { user: users[2].id },
  //       { group: groups[1].id },
  //       { role: 'team.admin' },
  //       { created_by: user }
  //     );
  //   });
  //
  //   lab.test(
  //     'should return all users if no filters are specified',
  //     async () => {
  //       const getListWithFiltersStub = Sandbox.stub(
  //         Resource,
  //         'getListWithFilters'
  //       ).returns(<any>users);
  //
  //       const result = await Resource.list();
  //       Code.expect(result).to.equal(users);
  //       Sandbox.assert.notCalled(getListWithFiltersStub);
  //     }
  //   );
  //
  //   lab.test('should return users that match the filter', async () => {
  //     const removeTimestamps = R.omit(['createdAt', 'updatedAt']);
  //
  //     const filter = {
  //       entity: 'gojek',
  //       privacy: 'public',
  //       action: 'firehose.read'
  //     };
  //     const result = (await Resource.list(filter)).map(removeTimestamps);
  //
  //     const expectedResult = [
  //       { ...users[0], policies: [] },
  //       { ...users[1], policies: [] },
  //       { ...users[2], policies: [userEntityPolicy] }
  //     ].map(removeTimestamps);
  //
  //     // ? We need to sort before checking because [1, 2, 3] != [2, 1, 3]
  //     Code.expect(R.sortBy(R.propOr(null, 'id'), result)).to.equal(
  //       R.sortBy(R.propOr(null, 'id'), expectedResult)
  //     );
  //   });
  // });
});
