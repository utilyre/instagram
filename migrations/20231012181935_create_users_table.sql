-- +goose Up
-- +goose StatementBegin
CREATE TABLE "users" (
    "id" serial NOT NULL PRIMARY KEY,
    "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,

    "name" varchar(64) NOT NULL UNIQUE,
    "email" varchar(320) NOT NULL UNIQUE,
    "password" bytea NOT NULL
);

CREATE TRIGGER "refresh_users_updated_at"
    BEFORE UPDATE
    ON "users"
    FOR EACH ROW
    EXECUTE PROCEDURE refresh_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER "refresh_users_updated_at" ON "users";
DROP TABLE "users";
-- +goose StatementEnd
