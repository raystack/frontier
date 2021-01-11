import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  UpdateDateColumn
} from 'typeorm';

@Entity('users')
export class User {
  @PrimaryGeneratedColumn({
    type: 'bigint'
  })
  id: number;

  @Column({
    type: 'varchar',
    unique: true,
    nullable: false
  })
  username: string;

  @Column({
    type: 'varchar',
    unique: true,
    nullable: false
  })
  email: string;

  @Column({
    type: 'varchar',
    nullable: true
  })
  name: string;

  @Column({
    type: 'varchar',
    nullable: true
  })
  slack: string;

  @Column({
    type: 'varchar',
    nullable: true
  })
  designation: string;

  @Column({
    type: 'varchar',
    nullable: true
  })
  company: string;

  @CreateDateColumn()
  created_at: string;

  @UpdateDateColumn()
  updated_at: string;
}
