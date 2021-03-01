import {
  Entity,
  Column,
  CreateDateColumn,
  UpdateDateColumn,
  BaseEntity,
  PrimaryGeneratedColumn
} from 'typeorm';

@Entity('groups')
export class Group extends BaseEntity {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column({
    type: 'varchar',
    nullable: false,
    unique: true
  })
  groupname: string;

  @Column({
    type: 'varchar',
    nullable: false
  })
  displayname: string;

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
