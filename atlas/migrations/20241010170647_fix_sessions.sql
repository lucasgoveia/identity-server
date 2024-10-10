-- Modify "user_sessions" table
ALTER TABLE "public"."user_sessions" DROP COLUMN "ended_at", DROP COLUMN "session_secret", ADD COLUMN "expires_at" timestamp NOT NULL;
-- Create index "user_sessions_session_id_expires_at_idx" to table: "user_sessions"
CREATE INDEX "user_sessions_session_id_expires_at_idx" ON "public"."user_sessions" ("session_id", "expires_at");
-- Create index "user_sessions_user_id_session_id_expires_at_idx" to table: "user_sessions"
CREATE INDEX "user_sessions_user_id_session_id_expires_at_idx" ON "public"."user_sessions" ("user_id", "session_id", "expires_at");
