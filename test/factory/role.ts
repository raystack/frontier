import Faker from 'faker';
import { define } from 'typeorm-seeding';
import { Role } from '../../src/model/role';

define(Role, (faker: typeof Faker) => {
  const role = new Role();
  role.id = faker.random.uuid();
  role.displayname = 'Role display name';
  role.attributes = ['entity', 'landscape'];
  role.metadata = {
    username: 'dev.team',
    email: `dev.team@go-jek.com`
  };
  role.tags = [];
  return role;
});
