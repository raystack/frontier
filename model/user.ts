import { EntitySchema } from 'typeorm';

export interface UserType {
  id: number;
  username: string;
  email: string;
  name?: string;
  slack?: string;
  designation?: string;
  company?: string;
  createdAt: string;
  updatedAt: string;
}

export const UserEntity = new EntitySchema<UserType>({
  name: 'users',
  columns: {
    id: {
      name: 'id',
      type: 'bigint',
      primary: true
    },
    username: {
      name: 'username',
      type: String,
      unique: true,
      nullable: false
    },
    email: {
      name: 'email',
      type: String,
      unique: true,
      nullable: false
    },
    name: {
      name: 'name',
      type: String
    },
    slack: {
      name: 'slack',
      type: String
    },
    designation: {
      name: 'designation',
      type: String
    },
    company: {
      name: 'company',
      type: String
    },
    createdAt: {
      name: 'created_at',
      type: 'timestamp with time zone',
      createDate: true
    },
    updatedAt: {
      name: 'updated_at',
      type: 'timestamp with time zone',
      updateDate: true
    }
  }
});
