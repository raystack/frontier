import Faker from 'faker';
import { define } from 'typeorm-seeding';
import { User } from '../../src/model/user';

define(User, (faker: typeof Faker) => {
  const randomNum = faker.random.number(100);
  const name = faker.name.firstName(randomNum);

  const user = new User();
  user.id = faker.random.uuid();
  user.username = faker.random.uuid();
  user.displayname = name;
  user.metadata = {
    email: `${user.username}@go-jek.com`
  };
  return user;
});
