package storage

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type Post struct {
	ID        int       `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	Image       string `db:"image"`
	Title       string `db:"title"`
	Description string `db:"description"`

	*User `db:"user"`
}

type PostStorage struct{ db *sqlx.DB }

func NewPostStorage(db *sqlx.DB) PostStorage {
	return PostStorage{db: db}
}

func (ps PostStorage) Create(post *Post) error {
	query := `
	INSERT INTO "posts"
	("user_id", "image", "title", "description")
	VALUES ($1, $2, $3, $4)
	RETURNING "id", "created_at", "updated_at";
	`

	return ps.db.Get(post, query, post.User.ID, post.Image, post.Title, post.Description)
}

func (ps PostStorage) ReadAll() ([]*Post, error) {
	query := `
	SELECT "posts"."id", "posts"."image", "posts"."title", "posts"."description", "users"."name" AS "user.name"
	FROM "posts"
	LEFT OUTER JOIN "users" ON "posts"."user_id" = "users"."id";
	`

	posts := []*Post{}
	if err := ps.db.Select(&posts, query); err != nil {
		return nil, err
	}

	return posts, nil
}
