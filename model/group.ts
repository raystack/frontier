import { EntitySchema } from 'typeorm';

export enum PrivacyEnum {
  PUBLIC = 'public',
  PRIVATE = 'private'
}

export interface GroupType {
  id: number;
  name: string;
  email: string;
  title?: string;
  slack?: string;
  description?: string;
  privacy: PrivacyEnum;
  createdAt: string;
  updatedAt: string;
}

export const GroupEntity = new EntitySchema<GroupType>({
  name: 'groups',
  columns: {
    id: {
      name: 'id',
      type: 'bigint',
      primary: true
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
    description: {
      name: 'description',
      type: String
    },
    privacy: {
      name: 'privacy',
      type: 'enum',
      enum: PrivacyEnum
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
