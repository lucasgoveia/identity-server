schema "public" {}

table "users" {
  schema = schema.public
  column "id" {
    null = false
    type = char(26)
  }
  column "name" {
    null = false
    type = varchar(256)
  }
  column "avatar_link" {
    null = true
    type = varchar(1000)
  }
  column "created_at" {
    null = false
    type = timestamp
  }
  column "updated_at" {
    null = false
    type = timestamp
  }
  column "deleted_at" {
    null = true
    type = timestamp
  }
  column "lockout_enabled" {
    null    = false
    type    = boolean
    default = true
  }
  column "lockout_end_date" {
    null = true
    type = timestamp
  }
  column "access_failed_count" {
    null    = false
    type    = int
    default = 0
  }
  column "two_factor_enabled" {
    null    = false
    type    = boolean
    default = false
  }
  primary_key {
    columns = [column.id]
  }

  index "users_id_access_failed_count_idx" {
    columns = [column.id, column.access_failed_count]
  }
}


enum "identity_type" {
  schema = schema.public
  values = ["email", "username", "phone", "social", "b2b", "passkey"]
}


table "user_identities" {
  schema = schema.public
  column "id" {
    null = false
    type = char(26)
  }
  column "user_id" {
    null = false
    type = char(26)
  }
  column "type" {
    null = false
    type = enum.identity_type // e.g., email, phone, social, sso, passkey
  }
  column "value" {
    null = true
    type = varchar(256) // For email, phone number, provider ID, etc.
  }
  column "credential" {
    null = true
    type = varchar(512) // Could be password hash, public key for passkeys, or null for SSO/social
  }
  column "provider" {
    null = true
    type = varchar(100) // Used for social and SSO (e.g., google, facebook, azure_ad), nullable for email/phone
  }
  column "verified" {
    null = false
    type = boolean
    default = false
  }
  column "created_at" {
    null = false
    type = timestamp
  }
  column "updated_at" {
    null = false
    type = timestamp
  }
  column "deleted_at" {
    null = true
    type = timestamp
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "user_identities_user_fk" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
  }
  index "user_identities_identity_type_idx" {
    columns = [column.type]
  }
  index "user_identities_user_identity_type_idx" {
    columns = [column.user_id, column.type]
  }

  unique "unique_identity_per_type" {
    columns = [column.type, column.value]
  }
}

table "user_sessions" {
  schema = schema.public
  column "session_id" {
    null = false
    type = char(26)
  }
  column "user_id" {
    null = false
    type = char(26)
  }
  column "identity_id" {
    null = false
    type = char(26)
  }
  column "ip_address" {
    null = false
    type = varchar(50)
  }
  column "user_agent" {
    null = false
    type = varchar(200)
  }
  column "created_at" {
    null = false
    type = timestamp
  }
  column "expires_at" {
    null = false
    type = timestamp
  }

  primary_key {
    columns = [column.session_id]
  }

  index "user_sessions_user_id" {
    columns = [ column.user_id ]
  }
  index "user_sessions_session_id_expires_at_idx" {
    columns = [column.session_id, column.expires_at]
  }
  index "user_sessions_user_id_session_id_expires_at_idx" {
    columns = [column.user_id, column.session_id, column.expires_at]
  }
}
