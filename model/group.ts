import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  UpdateDateColumn
} from 'typeorm';

export enum PrivacyEnum {
  PUBLIC = 'public',
  PRIVATE = 'private'
}

@Entity('groups')
export class Group {
  @PrimaryGeneratedColumn({
    type: 'bigint'
  })
  id: number;

  @Column({
    type: 'varchar',
    unique: true,
    nullable: false
  })
  name: string;

  @Column({
    type: 'varchar',
    nullable: false
  })
  title: string;

  @Column({
    type: 'enum',
    enum: PrivacyEnum,
    nullable: false,
    default: PrivacyEnum.PUBLIC
  })
  privacy: string;

  @Column({
    type: 'varchar',
    nullable: true
  })
  email: string;

  @Column({
    type: 'varchar',
    nullable: true
  })
  slack: string;

  @Column({
    type: 'varchar',
    nullable: true
  })
  description: string;

  @Column({
    type: 'varchar',
    nullable: true
  })
  product_group: string;

  @CreateDateColumn()
  created_at: string;

  @UpdateDateColumn()
  updated_at: string;
}
