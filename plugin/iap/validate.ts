import Hapi from '@hapi/hapi';
import { getUsernameFromEmail } from './utils';
import {
  getUserByMetadata,
  updateUserFromIAP
} from '../../app/profile/resource';
import { create as createUser } from '../../app/user/resource';

const validateByEmail = async (request: Hapi.Request, email: string) => {
  let credentials;
  const username = getUsernameFromEmail(email);

  const user =
    (await getUserByMetadata({ email })) ||
    (await getUserByMetadata({ username }));

  if (user) {
    const metadata = { ...user.metadata, email, username };
    // updateUserById just to keep google IAP and our DB in sync with email & username
    credentials = await updateUserFromIAP(user.id, {
      ...user,
      metadata
    });
  } else {
    credentials = await createUser({
      displayName: username,
      metadata: { username, email }
    });
  }

  return { isValid: true, credentials };
};

export default validateByEmail;
