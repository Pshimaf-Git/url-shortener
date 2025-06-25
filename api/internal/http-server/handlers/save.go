package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Pshimaf-Git/url-shortener/internal/database"
	"github.com/Pshimaf-Git/url-shortener/internal/http-server/reqcontext"
	"github.com/Pshimaf-Git/url-shortener/internal/lib/api/resp"
	"github.com/Pshimaf-Git/url-shortener/internal/lib/sl"
)

type Responce struct {
	resp.Response
	Alias string `json:"alias"`
}

const maxRetries int = 10

func (h *Handler) NewSave() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.Save"

		c := reqcontext.New(w, r)
		log := h.log.With(
			slog.String("fn", fn),
			slog.String(RequestID, c.RequestID()),
		)

		var req Request

		log.Info("decoding request body")

		if err := c.DecodeJSON(&req); err != nil {
			log.Error("decode req body", sl.Error(err))

			c.JSON(http.StatusInternalServerError, resp.Error(ErrInternalServer))
			return
		}

		log.Info("decoded requst body", slog.Any("request body", req))

		userProvaidedAlias := req.Alias
		url := req.URL
		stdLength := h.cfg.StdAliasLen

		if strings.EqualFold(url, "") {
			log.Info("request without url")

			c.JSON(http.StatusBadRequest, resp.Error(ErrEmprtyURl))
			return
		}

		if err := ValidateURLFormat(url); err != nil {
			h.log.Error("invalid URL format", slog.String("url", req.URL))

			c.JSON(http.StatusBadRequest, resp.Error(ErrInvalidURLFormat))

			return
		}

		var (
			finalAlias string
			err        error
		)

		if strings.EqualFold(userProvaidedAlias, "") {
			finalAlias, err = h.storage.SaveGeneratedURl(c.Context(), url, stdLength, maxRetries)
			if err != nil {
				if errors.Is(err, database.ErrMaxRetriesForGenerate) {
					log.Error("generate randim alias",
						slog.Int("max retries", maxRetries),
						slog.Int("standart alias length", stdLength),
						sl.Error(err),
					)

					c.JSON(http.StatusInternalServerError, resp.Error(ErrCanNotGenAlias))

					return
				}

				log.Error("saving URl with generated alias", slog.String("url", url), sl.Error(err))

				c.JSON(http.StatusInternalServerError, resp.Error(ErrInternalServer))

				return
			}

			log.Info("added alias", slog.String("url", url), slog.String("alias", finalAlias))

			c.JSON(http.StatusCreated, Responce{
				Response: resp.OK(),
				Alias:    finalAlias,
			})

		} else {
			log = log.With(slog.String("url", url), slog.String("alias", userProvaidedAlias))

			err := h.storage.SaveURL(c.Context(), url, userProvaidedAlias)
			if err != nil {
				if errors.Is(err, database.ErrURLExist) {
					log.Info("alias already exist")

					c.JSON(http.StatusBadRequest, resp.Error(ErrAliasExist))
					return
				}

				log.Error("url saver", sl.Error(err))

				c.JSON(http.StatusInternalServerError, resp.Error(ErrInternalServer))
				return
			}

			log.Info("url added")

			c.JSON(http.StatusCreated, Responce{
				Response: resp.OK(),
				Alias:    userProvaidedAlias,
			})
		}
	}
}
