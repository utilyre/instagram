package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/utilyre/instagram/internal/auth"
	"github.com/utilyre/instagram/internal/middleware"
	"github.com/utilyre/instagram/internal/storage"
	"github.com/utilyre/xmate"
)

const maxImageSize = 1 * 1024 * 1024 // 1MB

var (
	ErrNotMultipart = xmate.NewHTTPError(http.StatusBadRequest, "Content-Type not multipart/form-data")
)

type PostsResource struct {
	ErrorHandler xmate.ErrorHandler
	Validate     *validator.Validate
	PostStorage  storage.PostStorage
}

func (pr PostsResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", pr.ErrorHandler.HandleFunc(pr.readAll))

	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate(pr.ErrorHandler, jwtSecret))
		r.Post("/upload", pr.ErrorHandler.HandleFunc(pr.upload))
		r.Post("/", pr.ErrorHandler.HandleFunc(pr.create))
	})

	return r
}

func (pr PostsResource) upload(w http.ResponseWriter, r *http.Request) error {
	type Response struct {
		URL string `json:"url"`
	}

	if err := r.ParseMultipartForm(maxImageSize); err != nil {
		if errors.Is(err, http.ErrNotMultipart) {
			return ErrNotMultipart
		}

		return err
	}

	image, _, err := r.FormFile("image")
	if err != nil {
		return err
	}

	if err := os.MkdirAll("./images", 0700); err != nil {
		return err
	}

	name := uuid.New().String()
	f, err := os.OpenFile("./images/"+name, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	if _, err := io.Copy(f, image); err != nil {
		return err
	}

	url := *r.URL
	url.Path = "/images/" + name
	return xmate.WriteJSON(w, http.StatusOK, &Response{
		URL: url.String(),
	})
}

func (pr PostsResource) create(w http.ResponseWriter, r *http.Request) error {
	type Params struct {
		Image       string `json:"image" validate:"required,url,max=100"`
		Title       string `json:"title" validate:"required,max=50"`
		Description string `json:"description"`
	}

	type Response struct {
		ID          int    `json:"id"`
		Image       string `json:"image"`
		Title       string `json:"title"`
		Description string `json:"description,omitempty"`
	}

	params := new(Params)
	if err := json.NewDecoder(r.Body).Decode(params); err != nil {
		return xmate.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := pr.Validate.Struct(params); err != nil {
		return xmate.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	claims := r.Context().Value(middleware.ClaimsKey{}).(*auth.Claims)
	post := &storage.Post{
		User:        &storage.User{ID: claims.ID},
		Image:       params.Image,
		Title:       params.Title,
		Description: params.Description,
	}

	if err := pr.PostStorage.Create(post); err != nil {
		return err
	}

	return xmate.WriteJSON(w, http.StatusCreated, &Response{
		ID:          post.ID,
		Image:       post.Image,
		Title:       post.Title,
		Description: post.Description,
	})
}

func (pr PostsResource) readAll(w http.ResponseWriter, r *http.Request) error {
	type Response struct {
		ID          int    `json:"id"`
		Image       string `json:"image"`
		Title       string `json:"title"`
		Description string `json:"description,omitempty"`
		Author      string `json:"author,omitempty"`
	}

	posts, err := pr.PostStorage.ReadAll()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return xmate.WriteJSON(w, http.StatusOK, []*Response{})
		}

		return err
	}

	resp := make([]*Response, 0, len(posts))
	for _, post := range posts {
		resp = append(resp, &Response{
			ID:          post.ID,
			Image:       post.Image,
			Title:       post.Title,
			Description: post.Description,
			Author:      post.User.Name,
		})
	}

	return xmate.WriteJSON(w, http.StatusOK, resp)
}
