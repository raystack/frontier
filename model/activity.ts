import {
  Entity,
  Column,
  CreateDateColumn,
  BaseEntity,
  PrimaryGeneratedColumn
} from 'typeorm';

import Constants from '../utils/constant';
import { User } from './user';

@Entity(Constants.MODEL.Activity)
export class Activity extends BaseEntity {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column({
    type: 'varchar',
    nullable: false
  })
  title: string;

  @Column({
    type: 'varchar',
    nullable: false
  })
  model: string;

  @Column({
    type: 'varchar',
    nullable: false
  })
  documentId: string;

  @Column({
    type: 'jsonb',
    nullable: false
  })
  document: Record<string, string>;

  @Column({
    type: 'jsonb',
    nullable: true
  })
  diffs: Record<string, any>[];

  @CreateDateColumn()
  createdAt: string;

  @Column({
    type: 'jsonb',
    nullable: false
  })
  createdBy: User;
}
