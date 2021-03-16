import {
  Entity,
  Column,
  CreateDateColumn,
  UpdateDateColumn,
  BaseEntity,
  PrimaryGeneratedColumn
} from 'typeorm';
import Constants from '../utils/constant';

@Entity(Constants.MODEL.Role)
export class Role extends BaseEntity {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column({
    type: 'varchar',
    nullable: false
  })
  displayname: string;

  @Column({
    type: 'jsonb'
  })
  attributes: string[];

  @Column({
    type: 'jsonb',
    nullable: true
  })
  metadata: Record<string, any>;

  @CreateDateColumn()
  createdAt: string;

  @UpdateDateColumn()
  updatedAt: string;
}
