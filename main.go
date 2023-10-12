package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/utilyre/instagram/internal/api"
	"github.com/utilyre/instagram/internal/storage"
	"github.com/utilyre/xmate"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	r := chi.NewRouter()
	eh := newErrorHandler(logger)
	validate := validator.New()

	dsn := os.Getenv("DATABASE_URL")
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		logger.Error("failed to connect to postgres database", "dsn", dsn, "error", err)
		os.Exit(1)
	}

	userStorage := storage.NewUserStorage(db)
	postStorage := storage.NewPostStorage(db)

	r.Mount("/api/users", api.UsersResource{
		ErrorHandler: eh,
		Validate:     validate,
		UserStorage:  userStorage,
	}.Routes())

	r.Mount("/api/posts", api.PostsResource{
		ErrorHandler: eh,
		Validate:     validate,
		PostStorage:  postStorage,
	}.Routes())

	srv := &http.Server{Addr: ":3000", Handler: r}
	logger.Info("starting to listen and serve", "address", srv.Addr)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		logger.Error("failed to listen and serve", "error", err)
		os.Exit(1)
	}
}

func newErrorHandler(logger *slog.Logger) xmate.ErrorHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.Context().Value(xmate.ErrorKey{}).(error)

		if httpErr := new(xmate.HTTPError); errors.As(err, &httpErr) {
			_ = xmate.WriteText(w, httpErr.Code, httpErr.Message)
			return
		}

		logger.Warn(
			"failed to execute http handler",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("error", err.Error()),
		)

		_ = xmate.WriteText(w, http.StatusInternalServerError, "Internal Server Error")
	}
}
