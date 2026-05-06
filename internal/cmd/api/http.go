package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-api/internal/activitylog"
	"github.com/statisticsnorway/dapla-api/internal/auth/authn"
	"github.com/statisticsnorway/dapla-api/internal/auth/authz"
	"github.com/statisticsnorway/dapla-api/internal/auth/middleware"
	"github.com/statisticsnorway/dapla-api/internal/database"
	"github.com/statisticsnorway/dapla-api/internal/database/notify"
	"github.com/statisticsnorway/dapla-api/internal/feature"
	"github.com/statisticsnorway/dapla-api/internal/graph/loader"
	"github.com/statisticsnorway/dapla-api/internal/group"
	"github.com/statisticsnorway/dapla-api/internal/message"
	"github.com/statisticsnorway/dapla-api/internal/reconciler"
	"github.com/statisticsnorway/dapla-api/internal/search"
	"github.com/statisticsnorway/dapla-api/internal/section"
	"github.com/statisticsnorway/dapla-api/internal/serviceaccount"
	"github.com/statisticsnorway/dapla-api/internal/session"
	"github.com/statisticsnorway/dapla-api/internal/sharedbucketsstopgap"
	"github.com/statisticsnorway/dapla-api/internal/team"
	"github.com/statisticsnorway/dapla-api/internal/user"
	"github.com/statisticsnorway/dapla-api/internal/usersync"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
)

// runHttpServer will start the HTTP server
func runHttpServer(
	ctx context.Context,
	fakes Fakes,
	listenAddress string,
	pool *pgxpool.Pool,
	authHandler authn.Handler,
	jwtMiddleware func(http.Handler) http.Handler,
	graphHandler *handler.Server,
	notifier *notify.Notifier,
	log logrus.FieldLogger,
) error {
	router := chi.NewRouter()
	router.Handle("/", http.RedirectHandler("/api-docs", http.StatusPermanentRedirect))
	router.Method("GET", "/api-docs", otelhttp.NewHandler(playground.Handler("GraphQL playground", "/graphql"), "playground"))

	contextDependencies, err := ConfigureGraph(
		ctx,
		fakes,
		pool,
		notifier,
		log,
	)
	if err != nil {
		return err
	}

	router.Route("/graphql", func(r chi.Router) {
		middlewares := []func(http.Handler) http.Handler{
			contextDependencies,
			cors.New(
				cors.Options{
					AllowedOrigins:   []string{"https://*", "http://*"},
					AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
					AllowCredentials: true,
				},
			).Handler,
		}

		if fakes.WithInsecureAuth {
			middlewares = append(middlewares, middleware.InsecureUserHeader())
		}
		if jwtMiddleware != nil {
			middlewares = append(middlewares, jwtMiddleware)
		}
		if authHandler != nil {
			middlewares = append(middlewares, middleware.Oauth2Authentication(authHandler))
		}

		middlewares = append(
			middlewares,
			middleware.ApiKeyAuthentication(),
			middleware.RequireAuthenticatedUser(),
			otelhttp.NewMiddleware(
				"graphql",
				otelhttp.WithPublicEndpointFn(func(*http.Request) bool { return true }),
				otelhttp.WithSpanOptions(trace.WithAttributes(semconv.ServiceName("http"))),
			),
		)
		r.Use(middlewares...)
		r.Method("POST", "/", graphHandler)
	})

	if authHandler != nil {
		router.Route("/oauth2", func(r chi.Router) {
			r.Use(contextDependencies)
			r.Get("/login", authHandler.Login)
			r.Get("/logout", authHandler.Logout)
			r.Get("/callback", authHandler.Callback)
		})
	}

	srv := &http.Server{
		Addr:              listenAddress,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		log.Infof("HTTP server shutting down...")
		if err := srv.Shutdown(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.WithError(err).Infof("HTTP server shutdown failed")
			return err
		}
		return nil
	})

	wg.Go(func() error {
		log.Infof("HTTP server accepting requests on %q", listenAddress)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.WithError(err).Infof("unexpected error from HTTP server")
			return err
		}
		log.Infof("HTTP server finished, terminating...")
		return nil
	})
	return wg.Wait()
}

func ConfigureGraph(
	ctx context.Context,
	fakes Fakes,
	pool *pgxpool.Pool,
	notifier *notify.Notifier,
	log logrus.FieldLogger,
) (func(http.Handler) http.Handler, error) {
	searcher, err := search.New(ctx, pool, log.WithField("subsystem", "search_bleve"))
	if err != nil {
		return nil, fmt.Errorf("init bleve: %w", err)
	}

	team.AddSearch(searcher, pool, notifier, log.WithField("subsystem", "team_search"))
	group.AddSearch(searcher, pool, notifier, log.WithField("subsystem", "group_search"))
	sharedbucketsstopgap.AddSearch(searcher, pool, notifier, log.WithField("subsystem", "sharedbucket_search"))

	// Re-index all to initialize the search index
	if err := searcher.ReIndex(ctx); err != nil {
		return nil, fmt.Errorf("reindex all: %w", err)
	}

	setupContext := func(ctx context.Context) context.Context {
		ctx = database.NewLoaderContext(ctx, pool)
		ctx = team.NewLoaderContext(ctx, pool)
		ctx = user.NewLoaderContext(ctx, pool)
		ctx = usersync.NewLoaderContext(ctx, pool)
		ctx = authz.NewLoaderContext(ctx, pool)
		ctx = activitylog.NewLoaderContext(ctx, pool)
		ctx = reconciler.NewLoaderContext(ctx, pool)
		ctx = serviceaccount.NewLoaderContext(ctx, pool)
		ctx = session.NewLoaderContext(ctx, pool)
		ctx = section.NewLoaderContext(ctx, pool)
		ctx = search.NewLoaderContext(ctx, pool, searcher)
		ctx = feature.NewLoaderContext(ctx)
		ctx = group.NewLoaderContext(ctx, pool, log)
		ctx = sharedbucketsstopgap.NewLoaderContext(ctx, pool)
		ctx = message.NewLoaderContext(ctx, pool)
		return ctx
	}

	return loader.Middleware(setupContext), nil
}
