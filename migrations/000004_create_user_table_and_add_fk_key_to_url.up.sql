CREATE TABLE "users" (
                         "id" BIGSERIAL PRIMARY KEY,
                         "email" VARCHAR(255) NOT NULL UNIQUE,
                         "password_hash" VARCHAR(255) NOT NULL,
                         "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                         "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE "urls" ADD COLUMN IF NOT EXISTS "user_id" BIGINT;

ALTER TABLE "urls"
    ADD CONSTRAINT "fk_urls_user"
        FOREIGN KEY ("user_id")
            REFERENCES "users"("id")
            ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS "idx_users_email" ON "users"("email");
CREATE INDEX IF NOT EXISTS "idx_urls_user_id" ON "urls"("user_id");