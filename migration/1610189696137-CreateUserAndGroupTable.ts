/* eslint-disable class-methods-use-this */
import { MigrationInterface, QueryRunner } from 'typeorm';

export class CreateUserAndGroupTable1610189696137
  implements MigrationInterface {
  name = 'CreateUserAndGroupTable1610189696137';

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `CREATE TYPE "groups_privacy_enum" AS ENUM('public', 'private')`
    );
    await queryRunner.query(
      `CREATE TABLE "groups" ("id" BIGSERIAL NOT NULL, "name" character varying NOT NULL, "title" character varying NOT NULL, "email" character varying NOT NULL, "description" character varying, "slack" character varying, "privacy" "groups_privacy_enum" NOT NULL, "product_group" character varying, "created_at" TIMESTAMP NOT NULL DEFAULT now(), "updated_at" TIMESTAMP NOT NULL DEFAULT now(), CONSTRAINT "UQ_ef4dc213c6cbcc25c78126c2df5" UNIQUE ("email"), CONSTRAINT "PK_659d1483316afb28afd3a90646e" PRIMARY KEY ("id"))`
    );
    await queryRunner.query(
      `CREATE TABLE "users" ("id" BIGSERIAL NOT NULL, "username" character varying NOT NULL, "email" character varying NOT NULL, "name" character varying, "slack" character varying, "designation" character varying, "company" character varying, "created_at" TIMESTAMP NOT NULL DEFAULT now(), "updated_at" TIMESTAMP NOT NULL DEFAULT now(), CONSTRAINT "UQ_fe0bb3f6520ee0469504521e710" UNIQUE ("username"), CONSTRAINT "UQ_97672ac88f789774dd47f7c8be3" UNIQUE ("email"), CONSTRAINT "PK_a3ffb1c0c8416b9fc6f907b7433" PRIMARY KEY ("id"))`
    );
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP TABLE "users"`);
    await queryRunner.query(`DROP TABLE "groups"`);
    await queryRunner.query(`DROP TYPE "groups_privacy_enum"`);
  }
}
