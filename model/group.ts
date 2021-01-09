import { EntitySchema } from 'typeorm';

export enum PrivacyEnum {
  PUBLIC = 'public',
  PRIVATE = 'private'
}

export interface GroupType {
  id: number;
  name: string;
  email: string;
  title: string;
  slack?: string;
  description?: string;
  product_group?: string;
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
      primary: true,
      generated: 'increment'
    },
    name: {
      name: 'name',
      type: String,
      nullable: false
    },
    title: {
      name: 'title',
      type: String,
      nullable: false
    },
    email: {
      name: 'email',
      type: String,
      unique: true,
      nullable: true
    },
    description: {
      name: 'description',
      type: String,
      nullable: true
    },
    slack: {
      name: 'slack',
      type: String,
      nullable: true
    },
    privacy: {
      name: 'privacy',
      type: 'enum',
      enum: PrivacyEnum
    },
    product_group: {
      name: 'product_group',
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
