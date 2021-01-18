import { getManager } from 'typeorm';
import { User } from '../../model/user';

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
  const UserRepository = getManager().getRepository(User);
  return UserRepository.createQueryBuilder('user')
    .where('user.id = :id', { id })
    .getOne();
}

export async function updateUserByID(id: number, data: any): Promise<any> {
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
