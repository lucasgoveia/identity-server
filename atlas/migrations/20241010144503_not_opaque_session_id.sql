-- Modify "user_sessions" table
ALTER TABLE "public"."user_sessions" DROP COLUMN "id", ALTER COLUMN "session_id" TYPE character(26), ADD PRIMARY KEY ("session_id");
-- Rename an index from "column.user_id" to "user_sessions_user_id"
ALTER INDEX "public"."column.user_id" RENAME TO "user_sessions_user_id";
