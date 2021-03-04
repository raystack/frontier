import { MigrationInterface, QueryRunner } from 'typeorm';

export class CreateActivityTable1614858387492 implements MigrationInterface {
  name = 'CreateActivityTable1614858387492';

  public up = async (queryRunner: QueryRunner) => {
    await queryRunner.query(`
        create table activities (
          id uuid default uuid_generate_v4() not null constraint activities_pk primary key,
          title character varying not null,
          model character varying not null,
          document_id character varying not null,
          document jsonb,
          diffs jsonb,
          created_at timestamp default now() not null );
    `);
  };

  public down = async (queryRunner: QueryRunner) => {
    await queryRunner.query(`DROP TABLE "activities"`);
  };
}
