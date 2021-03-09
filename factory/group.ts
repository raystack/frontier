import Faker from 'faker';
import { define } from 'typeorm-seeding';
import { Group } from '../model/group';

define(Group, (faker: typeof Faker) => {
  const randomNum = faker.random.number(100);
  const name = faker.name.firstName(randomNum);

  const group = new Group();
  group.id = faker.random.uuid();
  group.groupname = name.toLowerCase();
  group.displayname = name;
  group.metadata = {
    email: `${group.groupname}@go-jek.com`
  };
  return group;
});
