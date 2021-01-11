import Code from 'code';
import Lab from '@hapi/lab';
import * as Config from '../../../config/config';
import setupSampleData from './sample';
import CasbinSingleton from '../../../lib/casbin';
import { lab } from '../../setup';

exports.lab = Lab.script();

lab.experiment('IAM scenarios:', () => {
  let enforcer = null;

  lab.before(async () => {
    const dbUri = Config.get('/postgres').uri;
    enforcer = await CasbinSingleton.create(dbUri);
  });

  lab.beforeEach(async () => {
    await setupSampleData();
  });

  lab.test(
    'resource.manager(alice) of a team(transport) with particular project(gojek::production::id) access should have firehose.write access to the team resource(p-gojek-id-firehose-transport-123)',
    async () => {
      const sub = { user: 'alice' };
      const obj = {
        resource: 'p-gojek-id-firehose-transport-123'
      };
      const act = { action: 'firehose.write' };
      const res = await enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);
    }
  );

  lab.test(
    'resource.manager(alice) of a team(transport) should have view access to resource(p-gojek-id-firehose-augur-345) of other teams(augur) of an entity(gojek)',
    async () => {
      const sub = { user: 'alice' };
      const obj = {
        resource: 'p-gojek-id-firehose-augur-345'
      };
      const act = { action: 'firehose.read' };

      const res = await enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);
    }
  );

  lab.test(
    'resource.manager(alice) of a team(transport) should not have edit access to resource(p-gojek-id-firehose-augur-345) of other teams(augur) of an entity(gojek)',
    async () => {
      const sub = { user: 'alice' };
      const obj = {
        resource: 'p-gojek-id-firehose-augur-345'
      };
      const act = { action: 'firehose.write' };

      const res = await enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(false);
    }
  );

  lab.test(
    'resource.manager(alice) of a team(transport) should have create access to firehose of his project(gojek::production::id) + his own team(transport)',
    async () => {
      const sub = { user: 'alice' };
      const obj = {
        entity: 'gojek',
        environment: 'production',
        landscape: 'id',
        team: 'transport'
      };
      const act = { action: 'firehose.write' };

      const res = await enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);
    }
  );

  lab.test(
    'resource.manager(alice) of a team(transport) should not have create access to firehose of his project(gojek::production::id) + other team(augur)',
    async () => {
      const sub = { user: 'alice' };
      const obj = {
        entity: 'gojek',
        environment: 'production',
        landscape: 'id',
        team: 'augur'
      };
      const act = { action: 'firehose.write' };

      const res = await enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(false);
    }
  );

  lab.test(
    'resource.manager(alice) of a team(transport) of one entity(gojek) should not have view access to resource(p-gofin-id-firehose-gofinance-789) of another entity(gofin)',
    async () => {
      const sub = { user: 'alice' };
      const obj = {
        resource: 'p-gofin-id-firehose-gofinance-789'
      };
      const act = { action: 'firehose.read' };

      const res = await enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(false);
    }
  );

  lab.test(
    'dwh.manager(alice) of an entity project(gojek) should be able to create a beast in that entity/sub-entity(gojek::production::id)',
    async () => {
      const sub = { user: 'alice' };
      const obj = {
        entity: 'gojek',
        environment: 'production',
        landscape: 'vn'
      };
      const act = { action: 'beast.write' };

      const res = await enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);
    }
  );

  lab.test(
    'dwh.manager(alice) of an entity project(gojek) should not be able to create a beast in another entity/sub-entity(gofin::production::id)',
    async () => {
      const sub = { user: 'alice' };
      const obj = {
        entity: 'gofin',
        environment: 'production',
        landscape: 'id'
      };
      const act = { action: 'beast.write' };

      const res = await enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(false);
    }
  );

  lab.test(
    'team.member(bob) of a team(transport) within an entity(gojek) should have view access to beast(p-gojek-id-beast-123) in that entity/sub-entity(gojek::production::id)',
    async () => {
      const sub = { user: 'bob' };
      const obj = {
        resource: 'p-gojek-id-beast-123'
      };
      const act = { action: 'beast.read' };

      const res = await enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);
    }
  );

  lab.test(
    'team.member(bob) of a team(transport) within an entity(gojek) should not have create access to beast in that entity/sub-entity(gojek::production::id)',
    async () => {
      const sub = { user: 'bob' };
      const obj = {
        entity: 'gojek',
        environment: 'production',
        landscape: 'id'
      };
      const act = { action: 'beast.write' };

      const res = await enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(false);
    }
  );

  lab.test(
    'team.member(bob) of a team without a dwh.manager role within an entity(gojek) should not have edit access to a beast(p-gojek-id-beast-123) in that entity/sub-entity(gojek::production::id)',
    async () => {
      const sub = { user: 'bob' };
      const obj = {
        resource: 'p-gojek-id-beast-123'
      };
      const act = { action: 'beast.write' };

      const res = await enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(false);
    }
  );

  lab.test(
    'team.admin(bob) of a team(transport) should have access to add members with role in the same team(transport)',
    async () => {
      const sub = { user: 'bob' };
      const obj = {
        team: 'transport'
      };
      const act = { action: 'role.creator' };

      const res = await enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);
    }
  );

  lab.test(
    'team.member(alice) of a team(transport) should have not access to add members with role in the same team(transport)',
    async () => {
      const sub = { user: 'alice' };
      const obj = {
        team: 'transport'
      };
      const act = { action: 'role.creator' };

      const res = await enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(false);
    }
  );

  lab.test(
    'entity.admin(cathy) of an entity(gojek) has team creation access in that entity(gojek) ',
    async () => {
      const sub = { user: 'cathy' };
      const obj = {
        entity: 'gojek'
      };
      const act = { action: 'team.creator' };

      const res = await enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);
    }
  );

  lab.test(
    'entity.admin(cathy) of an entity(gojek) has resource creation access in the team(transport) of that entity(gojek::production::id)',
    async () => {
      const sub = { user: 'cathy' };
      const obj = {
        entity: 'gojek',
        environment: 'production',
        landscape: 'id',
        team: 'transport'
      };
      const act = { action: 'firehose.write' };

      const res = await enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);
    }
  );

  lab.test(
    'entity.admin(cathy) of an entity(gojek) has resource edit access to any resource(p-gojek-id-firehose-transport-123) in that entity(gojek)',
    async () => {
      const sub = { user: 'cathy' };
      const obj = {
        resource: 'p-gojek-id-firehose-transport-123'
      };
      const act = { action: 'firehose.write' };

      const res = await enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);
    }
  );

  lab.test(
    'entity.admin(cathy) of an entity(gojek) has resource edit access to private resource(p-gojek-id-firehose-augur-private-345) in that entity(gojek)',
    async () => {
      const sub = { user: 'cathy' };
      const obj = {
        resource: 'p-gojek-id-firehose-augur-private-345'
      };
      const act = { action: 'firehose.write' };

      const res = await enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);
    }
  );

  lab.test(
    'entity.admin(cathy) of an entity has team member addition access in any team(transport) of that entity(gojek)',
    async () => {
      const sub = { user: 'cathy' };
      const obj = {
        team: 'transport'
      };
      const act = { action: 'role.creator' };

      const res = await enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);
    }
  );

  lab.test(
    'entity.admin(cathy) of an entity(gojek) should have beast creation access in the same entity(gojek::production::id)',
    async () => {
      const sub = { user: 'cathy' };
      const obj = {
        entity: 'gojek',
        environment: 'production',
        landscape: 'id'
      };
      const act = { action: 'beast.write' };

      const res = await enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);
    }
  );

  lab.test(
    'entity.admin(cathy) of an entity should not have view access to resource(p-gofin-id-firehose-gofinance-789) in a different entity(gofin)',
    async () => {
      const sub = { user: 'cathy' };
      const obj = {
        resource: 'p-gofin-id-firehose-gofinance-789'
      };
      const act = { action: 'firehose.read' };

      const res = await enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(false);
    }
  );

  lab.test(
    'entity.admin(cathy) of an entity(gojek) should not have team(gofinance) member addition access in a different entity(gofin)',
    async () => {
      const sub = { user: 'cathy' };
      const obj = {
        team: 'gofinance'
      };
      const act = { action: 'role.creator' };

      const res = await enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(false);
    }
  );

  lab.test(
    'a user(dave) of a team(augur) should have viewer access to private resources(p-gojek-id-firehose-augur-private-345) of the same team(augur) but not manager access',
    async () => {
      const sub = { user: 'dave' };
      const obj = {
        resource: 'p-gojek-id-firehose-augur-private-345'
      };
      const act = { action: 'firehose.read' };

      const res = await enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);

      const res2 = await enforcer.enforceJson(sub, obj, {
        action: 'firehose.write'
      });
      Code.expect(res2).to.equal(false);
    }
  );

  lab.test(
    'a user(frank) of a team(augur) should have manager access to private resources(p-gojek-id-firehose-augur-private-345) of the same team(augur) if given access to sub-entity(gojek::production::id)',
    async () => {
      const sub = { user: 'frank' };
      const obj = {
        resource: 'p-gojek-id-firehose-augur-private-345'
      };
      const act = { action: 'firehose.write' };

      const res = await enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);
    }
  );

  lab.test(
    'a user(alice) of another team(transport) of an entity(gojek) should not have viewer access to private resource(p-gojek-id-firehose-augur-private-345) of another team(augur) within that entity(gojek)',
    async () => {
      const sub = { user: 'alice' };
      const act = { action: 'firehose.read' };
      const obj = {
        resource: 'p-gojek-id-firehose-augur-private-345'
      };

      const res = await enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(false);
    }
  );

  lab.test('a user(gary) of DE should have all access', async () => {
    const sub = { user: 'gary' };
    const obj = {
      resource: 'p-gojek-id-firehose-augur-private-345'
    };
    const act = { action: 'firehose.write' };

    const res = await enforcer.enforceJson(sub, obj, act);
    Code.expect(res).to.equal(true);

    const act2 = { action: 'role.creator' };
    const res2 = await enforcer.enforceJson(
      sub,
      {
        team: 'gofinance'
      },
      act2
    );
    Code.expect(res2).to.equal(true);

    const act3 = { action: 'team.creator' };
    const res3 = await enforcer.enforceJson(
      sub,
      {
        entity: 'gojek'
      },
      act3
    );
    Code.expect(res3).to.equal(true);
  });
});
