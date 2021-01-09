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
      primary: true,
      generated: 'increment'
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
      type: String,
      nullable: true
    },
    slack: {
      name: 'slack',
      type: String,
      nullable: true
    },
    designation: {
      name: 'designation',
      type: String,
      nullable: true
    },
    company: {
      name: 'company',
      type: String,
      nullable: true
    },
    createdAt: {
      name: 'created_at',
      type: 'timestamp',
      createDate: true
    },
    updatedAt: {
      name: 'updated_at',
      type: 'timestamp',
      updateDate: true
    }
  }
});
