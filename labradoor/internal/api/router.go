package api

import (
	"crypto/subtle"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"

	"github.com/statisticsnorway/dapla-ctrl/labradoor/internal/config"
	"github.com/statisticsnorway/dapla-ctrl/labradoor/internal/parquedit"
)

func SetupRoutes(cfg config.RouterConfig, parquedit *parquedit.Client) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(httplog.RequestLogger(slog.Default(), &httplog.Options{}))

	// New routes should be added inside this group, for auth
	r.Group(func(r chi.Router) {
		r.Use(BearerAuth(cfg.AuthToken))

		r.Route("/parquedit/{team}", func(r chi.Router) {
			r.Use(func(h http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					ctx := r.Context()
					httplog.SetAttrs(ctx, slog.String("team", chi.URLParam(r, "team")))
					// So we can use slog as normal in other handler functions
					ctx = config.CtxWithLogger(ctx, slog.Default().With("team", chi.URLParam(r, "team")))

					if correlation_id:= r.Header.Get("X-Reconciler-CorrID"); correlation_id != "" {
						httplog.SetAttrs(ctx, slog.String("correlation_id", correlation_id))
						ctx = config.CtxWithLogger(ctx, slog.Default().With("correlation_id", correlation_id))

					}
					h.ServeHTTP(w, r.WithContext(ctx))
				})
			})
			r.Get("/", parquedit.HasEnabled)
			r.Put("/", parquedit.EnableForTeam)
			r.Delete("/", parquedit.DisableForTeam)
		})
	})

	return r
}

func BearerAuth(token string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			bearerAndToken := strings.Fields(auth)
			if len(bearerAndToken) != 2 || !strings.EqualFold(bearerAndToken[0], "bearer") {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			userToken := bearerAndToken[1]
			if subtle.ConstantTimeCompare([]byte(userToken), []byte(token)) != 1 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
