import {
  Entity,
  Column,
  CreateDateColumn,
  UpdateDateColumn,
  BaseEntity,
  PrimaryGeneratedColumn
} from 'typeorm';

// eslint-disable-next-line import/no-cycle
// import { Activity } from './activity';
import Constants from '../utils/constant';

@Entity(Constants.MODEL.User)
export class User extends BaseEntity {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column({
    type: 'varchar',
    nullable: false,
    unique: true
  })
  username: string;

  @Column({
    type: 'varchar',
    nullable: false
  })
  displayname: string;

  @Column({
    type: 'jsonb'
  })
  metadata: Record<string, any>;

  // @OneToMany(() => Activity, (activity) => activity.createdBy)
  // activities: Activity[];

  @CreateDateColumn()
  createdAt: string;

  @UpdateDateColumn()
  updatedAt: string;
}
