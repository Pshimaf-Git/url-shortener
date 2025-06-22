package handlers

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/Pshimaf-Git/url-shortener/internal/database"
	"github.com/Pshimaf-Git/url-shortener/internal/lib/api/resp"
	"github.com/Pshimaf-Git/url-shortener/internal/lib/errors"
	"github.com/Pshimaf-Git/url-shortener/internal/lib/random"
	"github.com/Pshimaf-Git/url-shortener/internal/lib/sl"
)

type Responce struct {
	resp.Response
	Alias string `json:"alias"`
}

func (h *Handler) NewSave() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.Save"

		c := NewContext(w, r)
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

		alias := req.Alias
		url := req.URL

		if strings.EqualFold(url, "") {
			log.Info("request without url")

			c.JSON(http.StatusBadRequest, resp.Error(ErrEmprtyURl))
			return
		}

		if strings.EqualFold(alias, "") {
			alias = random.StringRandV2(h.cfg.Server.StdAliasLen)
		}

		if err := ValidateURLFormat(url); err != nil {
			h.log.Error("invalid URL format", slog.String("url", req.URL))

			c.JSON(http.StatusBadRequest, resp.Error(ErrInvalidURLFormat))

			return
		}

		log = log.With(slog.String("url", url), slog.String("alias", alias))

		log.Info("decoded requst body")

		err := h.storage.SaveURL(c.Context(), url, alias)
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
			Alias:    alias,
		})
	}
}
