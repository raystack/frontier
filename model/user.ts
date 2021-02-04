import { Entity, Column, CreateDateColumn, UpdateDateColumn } from 'typeorm';

@Entity('users')
export class User {
  @Column({
    type: 'varchar',
    unique: true,
    nullable: false
  })
  id: string;

  @Column({
    type: 'varchar',
    unique: true,
    nullable: false
  })
  displayName: string;

  @Column({
    type: 'jsonb',
    nullable: true
  })
  metadata: Record<string, any>;

  @CreateDateColumn()
  created_at: string;

  @UpdateDateColumn()
  updated_at: string;
}
