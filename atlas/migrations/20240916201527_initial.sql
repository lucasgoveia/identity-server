-- Create enum type "identity_type"
CREATE TYPE "public"."identity_type" AS ENUM ('email', 'username', 'phone', 'social', 'b2b', 'passkey');
-- Create "user_sessions" table
CREATE TABLE "public"."user_sessions" ("id" integer NOT NULL GENERATED ALWAYS AS IDENTITY, "session_id" character varying(32) NOT NULL, "user_id" character(26) NOT NULL, "identity_id" character(26) NOT NULL, "ip_address" character varying(50) NOT NULL, "user_agent" character varying(200) NOT NULL, "device_fingerprint" character varying(128) NOT NULL, "created_at" timestamp NOT NULL, "ended_at" timestamp NULL, "session_secret" character varying(256) NOT NULL);
-- Create index "column.user_id" to table: "user_sessions"
CREATE INDEX "column.user_id" ON "public"."user_sessions" ("user_id");
-- Create index "user_sessions_session_id_ended_at_idx" to table: "user_sessions"
CREATE INDEX "user_sessions_session_id_ended_at_idx" ON "public"."user_sessions" ("session_id", "ended_at");
-- Create index "user_sessions_user_id_session_id_ended_at_idx" to table: "user_sessions"
CREATE INDEX "user_sessions_user_id_session_id_ended_at_idx" ON "public"."user_sessions" ("user_id", "session_id", "ended_at");
-- Create "users" table
CREATE TABLE "public"."users" ("id" character(26) NOT NULL, "name" character varying(256) NOT NULL, "avatar_link" character varying(1000) NULL, "created_at" timestamp NOT NULL, "updated_at" timestamp NOT NULL, "deleted_at" timestamp NULL, "lockout_enabled" boolean NOT NULL DEFAULT true, "lockout_end_date" timestamp NULL, "access_failed_count" integer NOT NULL DEFAULT 0, "two_factor_enabled" boolean NOT NULL DEFAULT false, PRIMARY KEY ("id"));
-- Create index "users_id_access_failed_count_idx" to table: "users"
CREATE INDEX "users_id_access_failed_count_idx" ON "public"."users" ("id", "access_failed_count");
-- Create "user_identities" table
CREATE TABLE "public"."user_identities" ("id" character(26) NOT NULL, "user_id" character(26) NOT NULL, "type" "public"."identity_type" NOT NULL, "value" character varying(256) NULL, "credential" character varying(512) NULL, "provider" character varying(100) NULL, "verified" boolean NOT NULL DEFAULT false, "created_at" timestamp NOT NULL, "updated_at" timestamp NOT NULL, "deleted_at" timestamp NULL, PRIMARY KEY ("id"), CONSTRAINT "unique_identity_per_type" UNIQUE ("type", "value"), CONSTRAINT "user_identities_user_fk" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- Create index "user_identities_identity_type_idx" to table: "user_identities"
CREATE INDEX "user_identities_identity_type_idx" ON "public"."user_identities" ("type");
-- Create index "user_identities_user_identity_type_idx" to table: "user_identities"
CREATE INDEX "user_identities_user_identity_type_idx" ON "public"."user_identities" ("user_id", "type");
