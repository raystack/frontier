import { MigrationInterface, QueryRunner } from 'typeorm';

export class RolesTags1624879411053 implements MigrationInterface {
  name = 'RolesTags1624879411053';

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `ALTER TABLE "roles" ADD "tags" character varying array NOT NULL DEFAULT array[]::text[]`
    );
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "roles" DROP COLUMN "tags"`);
  }
}
