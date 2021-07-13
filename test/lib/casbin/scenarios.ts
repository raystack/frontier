import Code from 'code';
import Lab from '@hapi/lab';
import { factory } from 'typeorm-seeding';
import * as Config from '../../../src/config/config';
import setupSampleData from './sample';
import CasbinSingleton from '../../../src/lib/casbin';
import { lab } from '../../setup';
import connection from '../../connection';
import { User } from '../../../src/model/user';

exports.lab = Lab.script();

const testScenarios = () => {
  lab.beforeEach(async () => {
    await setupSampleData();
  });

  lab.test(
    'enforcer should work for empty policy list',
    async ({ context }) => {
      await connection.clear();
      let error;
      try {
        await context.enforcer.batchEnforceJson([]);
      } catch (e) {
        error = e;
      }
      Code.expect(error).to.undefined();
    }
  );

  lab.test(
    'henry of a team(marketplace) should have default access to the team(marketplace)',
    async ({ context }) => {
      const sub = { user: 'henry' };
      const obj = {
        team: 'marketplace'
      };
      const act = { action: 'dataaccess.manage' };
      const res = await context.enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);
    }
  );

  lab.test(
    'henry of a team(marketplace) should not have default access to the team(augur)',
    async ({ context }) => {
      const sub = { user: 'henry' };
      const obj = {
        team: 'augur'
      };
      const act = { action: 'dataaccess.manage' };
      const res = await context.enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(false);
    }
  );

  lab.test(
    'resource.manager(dave) of a team(augur) with particular project(gojek::prrivacy) access should have any access to it',
    async ({ context }) => {
      const sub = { user: 'dave' };
      const obj = {
        entity: 'gojek',
        privacy: 'public'
      };
      const act = { action: 'any' };
      const res = await context.enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);
    }
  );

  lab.test(
    'resource.manager(alice) of a team(transport) with particular project(gojek::production::id) access should have firehose.write access to the team resource(p-gojek-id-firehose-transport-123)',
    async ({ context }) => {
      const sub = { user: 'alice' };
      const obj = {
        resource: 'p-gojek-id-firehose-transport-123'
      };
      const user = await factory(User)().create();
      user.id = 'alice';
      user.username = 'alice';
      user.displayname = 'alice';
      const act = { action: 'firehose.write' };
      const res = await context.enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);
    }
  );

  lab.test(
    'resource.manager(alice) of a team(transport) should have view access to resource(p-gojek-id-firehose-augur-345) of other teams(augur) of an entity(gojek)',
    async ({ context }) => {
      const sub = { user: 'alice' };
      const obj = {
        resource: 'p-gojek-id-firehose-augur-345'
      };
      const act = { action: 'firehose.read' };

      const res = await context.enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);
    }
  );

  lab.test(
    'resource.manager(alice) of a team(transport) should not have edit access to resource(p-gojek-id-firehose-augur-345) of other teams(augur) of an entity(gojek)',
    async ({ context }) => {
      const sub = { user: 'alice' };
      const obj = {
        resource: 'p-gojek-id-firehose-augur-345'
      };
      const act = { action: 'firehose.write' };

      const res = await context.enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(false);
    }
  );

  lab.test(
    'resource.manager(alice) of a team(transport) should have create access to firehose of his project(gojek::production::id) + his own team(transport)',
    async ({ context }) => {
      const sub = { user: 'alice' };
      const obj = {
        entity: 'gojek',
        environment: 'production',
        landscape: 'id',
        team: 'transport'
      };
      const act = { action: 'firehose.write' };

      const res = await context.enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);
    }
  );

  lab.test(
    'resource.manager(alice) of a team(transport) should not have create access to firehose of his project(gojek::production::id) + other team(augur)',
    async ({ context }) => {
      const sub = { user: 'alice' };
      const obj = {
        entity: 'gojek',
        environment: 'production',
        landscape: 'id',
        team: 'augur'
      };
      const act = { action: 'firehose.write' };

      const res = await context.enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(false);
    }
  );

  lab.test(
    'resource.manager(alice) of a team(transport) of one entity(gojek) should not have view access to resource(p-gofin-id-firehose-gofinance-789) of another entity(gofin)',
    async ({ context }) => {
      const sub = { user: 'alice' };
      const obj = {
        resource: 'p-gofin-id-firehose-gofinance-789'
      };
      const act = { action: 'firehose.read' };

      const res = await context.enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(false);
    }
  );

  lab.test(
    'dwh.manager(alice) of an entity project(gojek) should be able to create a beast in that entity/sub-entity(gojek::production::id)',
    async ({ context }) => {
      const sub = { user: 'alice' };
      const obj = {
        entity: 'gojek',
        environment: 'production',
        landscape: 'vn'
      };
      const act = { action: 'beast.write' };

      const res = await context.enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);
    }
  );

  lab.test(
    'dwh.manager(alice) of an entity project(gojek) should not be able to create a beast in another entity/sub-entity(gofin::production::id)',
    async ({ context }) => {
      const sub = { user: 'alice' };
      const obj = {
        entity: 'gofin',
        environment: 'production',
        landscape: 'id'
      };
      const act = { action: 'beast.write' };

      const res = await context.enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(false);
    }
  );

  lab.test(
    'team.member(bob) of a team(transport) within an entity(gojek) should have view access to beast(p-gojek-id-beast-123) in that entity/sub-entity(gojek::production::id)',
    async ({ context }) => {
      const sub = { user: 'bob' };
      const obj = {
        resource: 'p-gojek-id-beast-123'
      };
      const act = { action: 'beast.read' };

      const res = await context.enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);
    }
  );

  lab.test(
    'team.member(bob) of a team(transport) within an entity(gojek) should not have create access to beast in that entity/sub-entity(gojek::production::id)',
    async ({ context }) => {
      const sub = { user: 'bob' };
      const obj = {
        entity: 'gojek',
        environment: 'production',
        landscape: 'id'
      };
      const act = { action: 'beast.write' };

      const res = await context.enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(false);
    }
  );

  lab.test(
    'team.member(bob) of a team without a dwh.manager role within an entity(gojek) should not have edit access to a beast(p-gojek-id-beast-123) in that entity/sub-entity(gojek::production::id)',
    async ({ context }) => {
      const sub = { user: 'bob' };
      const obj = {
        resource: 'p-gojek-id-beast-123'
      };
      const act = { action: 'beast.write' };

      const res = await context.enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(false);
    }
  );

  lab.test(
    'team.admin(bob) of a team(transport) should have access to add members with role in the same team(transport)',
    async ({ context }) => {
      const sub = { user: 'bob' };
      const obj = {
        team: 'transport'
      };
      const act = { action: 'role.creator' };

      const res = await context.enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);
    }
  );

  lab.test(
    'team.member(alice) of a team(transport) should have not access to add members with role in the same team(transport)',
    async ({ context }) => {
      const sub = { user: 'alice' };
      const obj = {
        team: 'transport'
      };
      const act = { action: 'role.creator' };

      const res = await context.enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(false);
    }
  );

  lab.test(
    'entity.admin(cathy) of an entity(gojek) has team creation access in that entity(gojek) ',
    async ({ context }) => {
      const sub = { user: 'cathy' };
      const obj = {
        entity: 'gojek'
      };
      const act = { action: 'team.creator' };

      const res = await context.enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);
    }
  );

  lab.test(
    'entity.admin(cathy) of an entity(gojek) has resource creation access in the team(transport) of that entity(gojek::production::id)',
    async ({ context }) => {
      const sub = { user: 'cathy' };
      const obj = {
        entity: 'gojek',
        environment: 'production',
        landscape: 'id',
        team: 'transport'
      };
      const act = { action: 'firehose.write' };

      const res = await context.enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);
    }
  );

  lab.test(
    'entity.admin(cathy) of an entity(gojek) has resource edit access to any resource(p-gojek-id-firehose-transport-123) in that entity(gojek)',
    async ({ context }) => {
      const sub = { user: 'cathy' };
      const obj = {
        resource: 'p-gojek-id-firehose-transport-123'
      };
      const act = { action: 'firehose.write' };

      const res = await context.enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);
    }
  );

  lab.test(
    'entity.admin(cathy) of an entity(gojek) has resource edit access to private resource(p-gojek-id-firehose-augur-private-345) in that entity(gojek)',
    async ({ context }) => {
      const sub = { user: 'cathy' };
      const obj = {
        resource: 'p-gojek-id-firehose-augur-private-345'
      };
      const act = { action: 'firehose.write' };

      const res = await context.enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);
    }
  );

  lab.test(
    'entity.admin(cathy) of an entity has team member addition access in any team(transport) of that entity(gojek)',
    async ({ context }) => {
      const sub = { user: 'cathy' };
      const obj = {
        team: 'transport'
      };
      const act = { action: 'role.creator' };

      const res = await context.enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);
    }
  );

  lab.test(
    'entity.admin(cathy) of an entity(gojek) should have beast creation access in the same entity(gojek::production::id)',
    async ({ context }) => {
      const sub = { user: 'cathy' };
      const obj = {
        entity: 'gojek',
        environment: 'production',
        landscape: 'id'
      };
      const act = { action: 'beast.write' };

      const res = await context.enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);
    }
  );

  lab.test(
    'entity.admin(cathy) of an entity should not have view access to resource(p-gofin-id-firehose-gofinance-789) in a different entity(gofin)',
    async ({ context }) => {
      const sub = { user: 'cathy' };
      const obj = {
        resource: 'p-gofin-id-firehose-gofinance-789'
      };
      const act = { action: 'firehose.read' };

      const res = await context.enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(false);
    }
  );

  lab.test(
    'entity.admin(cathy) of an entity(gojek) should not have team(gofinance) member addition access in a different entity(gofin)',
    async ({ context }) => {
      const sub = { user: 'cathy' };
      const obj = {
        team: 'gofinance'
      };
      const act = { action: 'role.creator' };

      const res = await context.enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(false);
    }
  );

  lab.test(
    'a user(dave) of a team(augur) should have viewer access to private resources(p-gojek-id-firehose-augur-private-345) of the same team(augur) but not manager access',
    async ({ context }) => {
      const sub = { user: 'dave' };
      const obj = {
        resource: 'p-gojek-id-firehose-augur-private-345'
      };
      const act = { action: 'firehose.read' };

      const res = await context.enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);

      const res2 = await context.enforcer.enforceJson(sub, obj, {
        action: 'firehose.write'
      });
      Code.expect(res2).to.equal(false);
    }
  );

  lab.test(
    'a user(frank) of a team(augur) should have manager access to private resources(p-gojek-id-firehose-augur-private-345) of the same team(augur) if given access to sub-entity(gojek::production::id)',
    async ({ context }) => {
      const sub = { user: 'frank' };
      const obj = {
        resource: 'p-gojek-id-firehose-augur-private-345'
      };
      const act = { action: 'firehose.write' };

      const res = await context.enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(true);
    }
  );

  lab.test(
    'a user(alice) of another team(transport) of an entity(gojek) should not have viewer access to private resource(p-gojek-id-firehose-augur-private-345) of another team(augur) within that entity(gojek)',
    async ({ context }) => {
      const sub = { user: 'alice' };
      const act = { action: 'firehose.read' };
      const obj = {
        resource: 'p-gojek-id-firehose-augur-private-345'
      };

      const res = await context.enforcer.enforceJson(sub, obj, act);
      Code.expect(res).to.equal(false);
    }
  );

  lab.test('a user(gary) of DE should have all access', async ({ context }) => {
    const sub = { user: 'gary' };
    const obj = {
      resource: 'p-gojek-id-firehose-augur-private-345'
    };
    const act = { action: 'firehose.write' };

    const res = await context.enforcer.enforceJson(sub, obj, act);
    Code.expect(res).to.equal(true);

    const act2 = { action: 'role.creator' };
    const res2 = await context.enforcer.enforceJson(
      sub,
      {
        team: 'gofinance'
      },
      act2
    );
    Code.expect(res2).to.equal(true);

    const act3 = { action: 'team.creator' };
    const res3 = await context.enforcer.enforceJson(
      sub,
      {
        entity: 'gojek'
      },
      act3
    );
    Code.expect(res3).to.equal(true);
  });
};

lab.experiment('IAM load filtered policies:', () => {
  lab.before(async ({ context }) => {
    const dbUri = Config.get('/postgres').uri;
    context.enforcer = await CasbinSingleton.create(dbUri);
  });

  lab.after(() => {
    CasbinSingleton.enforcer = null;
  });

  testScenarios();
});

lab.experiment('IAM load all policies:', () => {
  lab.before(async ({ context }) => {
    CasbinSingleton.filtered = false;
    const dbUri = Config.get('/postgres').uri;
    context.enforcer = await CasbinSingleton.create(dbUri);
  });

  lab.after(() => {
    CasbinSingleton.filtered = true;
    CasbinSingleton.enforcer = null;
  });

  testScenarios();
});
