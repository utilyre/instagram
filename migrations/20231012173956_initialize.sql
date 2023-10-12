-- +goose Up
-- +goose StatementBegin
CREATE TABLE "posts" (
    "id" serial NOT NULL PRIMARY KEY,
    "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,

    "image" varchar(100) NOT NULL,
    "title" varchar(50) NOT NULL,
    "description" text
);

CREATE FUNCTION refresh_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER "refresh_posts_updated_at"
    BEFORE UPDATE
    ON "posts"
    FOR EACH ROW
    EXECUTE PROCEDURE refresh_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER "refresh_posts_updated_at" ON "posts";
DROP FUNCTION "refresh_updated_at";
DROP TABLE "posts";
-- +goose StatementEnd
