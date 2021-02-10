import { getManager } from 'typeorm';
import { User } from '../../model/user';

export async function getUserById(id: number): Promise<any> {
  return await User.findOne(id);
}

export async function updateUserFromIAP(id: number, data: any): Promise<any> {
  const user = await User.findOne(id);

  const payload = Object.assign(user, {
    ...data,
    metadata: {
      ...user?.metadata,
      ...data.metadata
    }
  });
  return await User.save(payload);
}

export async function updateUserByID(id: number, data: any): Promise<any> {
  const user = await User.findOne(id);
  const payload = Object.assign(user, {
    ...data,
    metadata: {
      ...user?.metadata,
      ...data.metadata,
      email: user?.metadata.email,
      username: user?.metadata.username
    }
  });
  return await User.save(payload);
}
export async function getUserByMetadata(
  metadataQuery: Record<string, any>
): Promise<any> {
  const UserRepository = getManager().getRepository(User);
  return UserRepository.createQueryBuilder('user')
    .where('metadata ::jsonb @> :metadataQuery', { metadataQuery })
    .getOne();
}

export async function updateUserByMetadata(
  metadata: Record<string, any>,
  data: any
): Promise<any> {
  const user = await getUserByMetadata(metadata);
  return await User.save({ ...user, ...data });
}
