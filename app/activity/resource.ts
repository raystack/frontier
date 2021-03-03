import { Activity } from '../../model/activity';

export const get = async (team = '') => {
  let criteria: any = {
    order: {
      createdAt: 'DESC'
    }
  };

  if (team.length !== 0) {
    // fetch activities based on team
    criteria = Object.assign(criteria, {
      where: {
        team
      }
    });
  }

  return Activity.find(criteria);
};

export const create = async (payload: any) => {
  return await Activity.save({ ...payload });
};
