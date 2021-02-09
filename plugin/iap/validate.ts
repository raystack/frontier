import Hapi from '@hapi/hapi';
import { getUsernameFromEmail } from './utils';

const validateByEmail = async (request: Hapi.Request, email: string) => {
  // TODO: fetch user from db using email first, which will be present in metadata => use getUserByMetadata({email: email})
  // TODO: if not found then search using username, which will be present in metadata => use getUserByMetadata({username: getUsernameFromEmail(email)})
  // TODO: if user is still not found from both queries then createUser({ displayName: username, metadata: { username: '13123ads', email: email } })
  // TODO: if user is found in one of the query then updateUserById too just to keep google IAP and our DB in sync
  // TODO: Then fill the entire user object in credentials => const credentials = {id: '23241egdsf', displayName: '', metadata: { username: '', email: '' }}
  const username = getUsernameFromEmail(email);
  const id = '5655ed21-eaab-41d0-ae54-1acce751b307'; // alphanumeric user id
  const credentials = { id, username, email };

  return { isValid: true, credentials };
};

export default validateByEmail;
