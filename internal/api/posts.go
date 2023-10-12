package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/utilyre/instagram/internal/storage"
	"github.com/utilyre/xmate"
)

type PostsResource struct {
	ErrorHandler xmate.ErrorHandler
	Validate     *validator.Validate
	PostStorage  storage.PostStorage
}

func (pr PostsResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", pr.ErrorHandler.HandleFunc(pr.create))
	r.Get("/", pr.ErrorHandler.HandleFunc(pr.readAll))

	return r
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

	post := &storage.Post{
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
		})
	}

	return xmate.WriteJSON(w, http.StatusOK, resp)
}
