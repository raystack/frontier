import Faker from 'faker';
import { define } from 'typeorm-seeding';
import { Group } from '../model/group';

define(Group, (faker: typeof Faker) => {
  const randomNum = faker.random.number(1);
  const name = faker.name.firstName(randomNum);

  const group = new Group();
  group.name = name.toLowerCase();
  group.title = name;
  group.email = `${name.toLowerCase()}@go-jek.com`;
  group.privacy = 'public';
  return group;
});
