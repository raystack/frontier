/* eslint-disable class-methods-use-this */
import { MigrationInterface, QueryRunner } from 'typeorm';

export class AddUsernameGroupnameColumns1614617481382
  implements MigrationInterface {
  name = 'AddUsernameGroupnameColumns1614617481382';

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `ALTER TABLE "roles" RENAME COLUMN "display_name" TO "displayname"`
    );
    await queryRunner.query(`ALTER TABLE "groups" DROP COLUMN "display_name"`);
    await queryRunner.query(`ALTER TABLE "users" DROP COLUMN "display_name"`);
    await queryRunner.query(
      `ALTER TABLE "groups" ADD "groupname" character varying NOT NULL`
    );
    await queryRunner.query(
      `ALTER TABLE "groups" ADD CONSTRAINT "UQ_7da46d9af319f3d3fa3c09cff75" UNIQUE ("groupname")`
    );
    await queryRunner.query(
      `ALTER TABLE "groups" ADD "displayname" character varying NOT NULL`
    );
    await queryRunner.query(
      `ALTER TABLE "users" ADD "username" character varying NOT NULL`
    );
    await queryRunner.query(
      `ALTER TABLE "users" ADD CONSTRAINT "UQ_fe0bb3f6520ee0469504521e710" UNIQUE ("username")`
    );
    await queryRunner.query(
      `ALTER TABLE "users" ADD "displayname" character varying NOT NULL`
    );
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`ALTER TABLE "users" DROP COLUMN "displayname"`);
    await queryRunner.query(
      `ALTER TABLE "users" DROP CONSTRAINT "UQ_fe0bb3f6520ee0469504521e710"`
    );
    await queryRunner.query(`ALTER TABLE "users" DROP COLUMN "username"`);
    await queryRunner.query(`ALTER TABLE "groups" DROP COLUMN "displayname"`);
    await queryRunner.query(
      `ALTER TABLE "groups" DROP CONSTRAINT "UQ_7da46d9af319f3d3fa3c09cff75"`
    );
    await queryRunner.query(`ALTER TABLE "groups" DROP COLUMN "groupname"`);
    await queryRunner.query(
      `ALTER TABLE "users" ADD "display_name" character varying NOT NULL`
    );
    await queryRunner.query(
      `ALTER TABLE "groups" ADD "display_name" character varying NOT NULL`
    );
    await queryRunner.query(
      `ALTER TABLE "roles" RENAME COLUMN "displayname" TO "display_name"`
    );
  }
}
