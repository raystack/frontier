import { getManager } from 'typeorm';
import { User } from '../../model/user';

// TODO: User email will move to metadata so we would have to change this function to be more generic like
// TODO: const metadataQuery = {email: 'a@g.com'} getUserByMetadata(metadataQuery)
export async function getUserByEmail(email: string): Promise<any> {
  const UserRepository = getManager().getRepository(User);
  const selectUserFields = [
    'id',
    'username',
    'email',
    'name',
    'slack',
    'designation',
    'company'
  ];
  return UserRepository.createQueryBuilder('user')
    .select(selectUserFields.map((field) => `user.${field}`))
    .where('user.email = :email', { email })
    .getOne();
}

export async function getUserById(id: number): Promise<any> {
  // TODO: we can use User directly in the same way we did for Group API instead of using Repository
  const UserRepository = getManager().getRepository(User);
  return UserRepository.createQueryBuilder('user')
    .where('user.id = :id', { id })
    .getOne();
}

export async function updateUserByID(id: number, data: any): Promise<any> {
  // TODO: We can use User directly in the same way we did for Group API instead of using Repository
  // TODO: We need to update the metadata column too
  const UserRepository = getManager().getRepository(User);
  await UserRepository.createQueryBuilder()
    .update(User)
    .set({
      ...data
    })
    .where('id = :id', { id })
    .execute();
  return getUserById(id);
}

// TODO: User email will move to metadata so we would have to change this function to be more generic like
// TODO: const metadataQuery = {email: 'a@g.com'} updateUserByMetadata(metadataQuery, data)
export async function updateUserByEmail(
  email: string,
  data: any
): Promise<any> {
  const UserRepository = getManager().getRepository(User);
  await UserRepository.createQueryBuilder()
    .update(User)
    .set({
      ...data
    })
    .where('email = :email', { email })
    .execute();
  return getUserByEmail(email);
}
