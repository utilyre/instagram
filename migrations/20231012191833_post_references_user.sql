-- +goose Up
-- +goose StatementBegin
ALTER TABLE "posts"
    ADD COLUMN "user_id" int4 NOT NULL;

ALTER TABLE "posts"
    ADD CONSTRAINT "fkey_user_id" FOREIGN KEY ("user_id")
    REFERENCES "users" ("id");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE "posts" DROP CONSTRAINT "fkey_user_id";
ALTER TABLE "posts" DROP COLUMN "user_id";
-- +goose StatementEnd
