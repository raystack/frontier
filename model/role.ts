import { Entity, Column, CreateDateColumn, UpdateDateColumn } from 'typeorm';

@Entity('roles')
export class Role {
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
    array: true
  })
  attributes: Array<string>[];

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
