package save

import (
	"Project1_url-shotener/internal/lib/api/response"
	"Project1_url-shotener/internal/lib/logger/sl"
	"Project1_url-shotener/internal/lib/random"
	"Project1_url-shotener/internal/storage"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"golang.org/x/exp/slog"
	"io"
	"net/http"
)

type Request struct {
	Url   string `json:"url" validate:"required, url"`
	Alias string `json:"alias" validate:"required"`
}

const aliasLength = 6

type Response struct {
	response.Response

	Alias string `json:"alias,omitempty"`
}

type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			// Такую ошибку встретим, если получили запрос с пустым телом
			// Обработаем её отдельно
			log.Error("request body is empty")

			render.JSON(w, r, response.Response{
				Status: response.StatusError,
				Error:  "empty request",
			})

			return
		}
		if err != nil {
			log.Error("failed to decode request", sl.Err(err))
			render.JSON(w, r, response.Error("failed decode request"))
		}

		log.Info("request body decode", slog.Any("req", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, response.ValidationError(validateErr))
			return

		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}

		id, err := urlSaver.SaveURL(req.Url, alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url already exists", slog.String("url", req.Url))
			render.JSON(w, r, response.Error("url already exists"))
			return
		}
		if err != nil {
			log.Error("failed to add url", sl.Err(err))
			render.JSON(w, r, response.Error("failed add url"))

			return
		}

		log.Info("url added", slog.Int64("id", id))

		render.JSON(w, r, Response{
			Response: response.OK(),
			Alias:    alias,
		})

	}

}
