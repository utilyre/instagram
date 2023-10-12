package storage

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type Post struct {
	ID        int       `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	// TODO
	// UserID int `db:"user_id"`
	// User   User

	Image       string `db:"image"`
	Title       string `db:"title"`
	Description string `db:"description"`
}

type PostStorage struct{ db *sqlx.DB }

func NewPostStorage(db *sqlx.DB) PostStorage {
	return PostStorage{db: db}
}

func (s PostStorage) Create(post *Post) error {
	query := `
	INSERT
	INTO "posts"
	("image", "title", "description")
	VALUES ($1, $2, $3)
	RETURNING "id", "created_at", "updated_at";
	`

	return s.db.Get(post, query, post.Image, post.Title, post.Description)
}

func (s PostStorage) ReadAll() ([]*Post, error) {
	query := `
	SELECT "id", "image", "title", "description"
	FROM "posts";
	`

	posts := []*Post{}
	if err := s.db.Select(&posts, query); err != nil {
		return nil, err
	}

	return posts, nil
}
