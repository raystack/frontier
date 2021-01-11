import Faker from 'faker';
import { define } from 'typeorm-seeding';
import { User } from '../model/user';

define(User, (faker: typeof Faker) => {
  const randomNum = faker.random.number(1);
  const name = faker.name.firstName(randomNum);

  const user = new User();
  user.username = name.toLowerCase();
  user.email = `${name.toLowerCase()}@go-jek.com`;
  user.name = name;
  return user;
});
