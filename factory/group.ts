import Faker from 'faker';
import { define } from 'typeorm-seeding';
import { Group } from '../model/group';

define(Group, (faker: typeof Faker) => {
  const randomNum = faker.random.number(1);
  const name = faker.name.firstName(randomNum);

  const group = new Group();
  group.id = name.toLowerCase();
  group.displayName = name;
  group.metadata = {
    email: `${name.toLowerCase()}@go-jek.com`
  };
  return group;
});
