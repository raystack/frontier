/* eslint-disable class-methods-use-this */
import { MigrationInterface, QueryRunner } from 'typeorm';

export class CreateGroupTable1610032711915 implements MigrationInterface {
  name = 'CreateGroupTable1610032711915';

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `CREATE TYPE "groups_privacy_enum" AS ENUM('public', 'private')`
    );
    await queryRunner.query(
      `CREATE TABLE "groups" ("id" bigint NOT NULL, "email" character varying NOT NULL, "name" character varying NOT NULL, "slack" character varying NOT NULL, "description" character varying NOT NULL, "privacy" "groups_privacy_enum" NOT NULL, "created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), "updated_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), CONSTRAINT "UQ_ef4dc213c6cbcc25c78126c2df5" UNIQUE ("email"), CONSTRAINT "PK_659d1483316afb28afd3a90646e" PRIMARY KEY ("id"))`
    );
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP TABLE "groups"`);
    await queryRunner.query(`DROP TYPE "groups_privacy_enum"`);
  }
}
