import Faker from 'faker';
import { define } from 'typeorm-seeding';
import { Activity } from '../model/activity';

define(Activity, (faker: typeof Faker) => {
  const activity = new Activity();
  activity.id = faker.random.uuid();
  activity.title = faker.random.words(3).toString();
  activity.model = 'User';
  activity.document = {};
  activity.documentId = faker.random.uuid();
  activity.diffs = [];
  return activity;
});
