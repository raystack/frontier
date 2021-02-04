import Hapi from '@hapi/hapi';
import { getUsernameFromEmail } from './utils';

const validateByEmail = async (request: Hapi.Request, email: string) => {
  // TODO: fetch user from db using username and upsert and validate the user
  const username = getUsernameFromEmail(email);
  const id = ''; // alphanumeric user id
  const credentials = { id, username, email };

  return { isValid: true, credentials };
};

export default validateByEmail;
