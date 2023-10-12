package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/lib/pq"
	"github.com/utilyre/instagram/internal/auth"
	"github.com/utilyre/instagram/internal/storage"
	"github.com/utilyre/xmate"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists = xmate.NewHTTPError(http.StatusConflict, "user already exists")
	ErrUserNotFound      = xmate.NewHTTPError(http.StatusNotFound, "user not found")
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

type UsersResource struct {
	ErrorHandler xmate.ErrorHandler
	Validate     *validator.Validate
	UserStorage  storage.UserStorage
}

func (ur UsersResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", ur.ErrorHandler.HandleFunc(ur.create))
	r.Post("/login", ur.ErrorHandler.HandleFunc(ur.login))
	r.Get("/", ur.ErrorHandler.HandleFunc(ur.readAll))

	return r
}

func (ur UsersResource) create(w http.ResponseWriter, r *http.Request) error {
	type Params struct {
		Name     string `json:"name" validate:"required,min=3,max=64"`
		Email    string `json:"email" validate:"required,email,max=320"`
		Password string `json:"password" validate:"required,min=8,max=1024"`
	}

	type Response struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	params := new(Params)
	if err := json.NewDecoder(r.Body).Decode(params); err != nil {
		return xmate.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := ur.Validate.Struct(params); err != nil {
		return xmate.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	password, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &storage.User{
		Name:     params.Name,
		Email:    params.Email,
		Password: password,
	}

	if err := ur.UserStorage.Create(user); err != nil {
		if pqErr := new(pq.Error); errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrUserAlreadyExists
		}

		return err
	}

	return xmate.WriteJSON(w, http.StatusCreated, &Response{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	})
}

func (ur UsersResource) login(w http.ResponseWriter, r *http.Request) error {
	type Params struct {
		Name     string `json:"name" validate:"required,min=3,max=64"`
		Password string `json:"password" validate:"required,min=8,max=1024"`
	}

	type Response struct {
		Token string `json:"token"`
	}

	params := new(Params)
	if err := json.NewDecoder(r.Body).Decode(params); err != nil {
		return xmate.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := ur.Validate.Struct(params); err != nil {
		return xmate.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	user := &storage.User{
		Name: params.Name,
	}

	if err := ur.UserStorage.ReadByName(user); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}

		return err
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(params.Password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrUserNotFound
		}

		return err
	}

	token, err := auth.Generate(jwtSecret, &auth.Claims{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	})
	if err != nil {
		return err
	}

	return xmate.WriteJSON(w, http.StatusCreated, &Response{
		Token: token,
	})
}

func (ur UsersResource) readAll(w http.ResponseWriter, r *http.Request) error {
	type Response struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	users, err := ur.UserStorage.ReadAll()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return xmate.WriteJSON(w, http.StatusOK, []*Response{})
		}

		return err
	}

	resp := make([]*Response, 0, len(users))
	for _, user := range users {
		resp = append(resp, &Response{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		})
	}

	return xmate.WriteJSON(w, http.StatusOK, resp)
}
