package storage

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type User struct {
	ID        int       `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	Name     string `db:"name"`
	Email    string `db:"email"`
	Password []byte `db:"password"`
}

type UserStorage struct{ db *sqlx.DB }

func NewUserStorage(db *sqlx.DB) UserStorage {
	return UserStorage{db: db}
}

func (us UserStorage) Create(user *User) error {
	query := `
	INSERT INTO "users"
	("name", "email", "password")
	VALUES ($1, $2, $3)
	RETURNING "id", "created_at", "updated_at";
	`

	return us.db.Get(user, query, user.Name, user.Email, user.Password)
}

func (us UserStorage) ReadAll() ([]*User, error) {
	query := `
	SELECT "id", "name", "email"
	FROM "users";
	`

	users := []*User{}
	if err := us.db.Select(&users, query); err != nil {
		return nil, err
	}

	return users, nil
}

func (us UserStorage) ReadByName(user *User) error {
	query := `
	SELECT "id", "email", "password"
	FROM "users"
	WHERE "name" = $1;
	`

	return us.db.Get(user, query, user.Name)
}
