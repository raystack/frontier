import Lab from '@hapi/lab';
import { useSeeding } from 'typeorm-seeding';
import connectionWrapper from './connection';

export const lab = Lab.script();

lab.before(async () => {
  await connectionWrapper.create();
  await useSeeding();
});

lab.afterEach(async () => {
  await connectionWrapper.clear();
});

lab.after(async () => {
  await connectionWrapper.close();
});
