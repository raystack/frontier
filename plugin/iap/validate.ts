import Hapi from '@hapi/hapi';
import { getUsernameFromEmail } from './utils';

const validateByEmail = async (request: Hapi.Request, email: string) => {
  // TODO: fetch user from db using username and upsert and validate the user
  const username = getUsernameFromEmail(email);
  const id = '0c351e11-776f-41e1-a023-45e27e1728ee'; // alphanumeric user id
  const credentials = { id, username, email };

  return { isValid: true, credentials };
};

export default validateByEmail;
