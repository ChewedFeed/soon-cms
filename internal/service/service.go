package service

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	bugLog "github.com/bugfixes/go-bugfixes/logs"
	bugMiddleware "github.com/bugfixes/go-bugfixes/middleware"
	"github.com/chewedfeed/soon-cms/internal/config"
	retro "github.com/chewedfeed/soon-cms/internal/soon-cms"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog"
	healthcheck "github.com/keloran/go-healthcheck"
	probe "github.com/keloran/go-probe"
)

type Service struct {
	Config *config.Config
}

func (s *Service) Start() error {
	bugLog.Local().Info("Starting CMS")

	logger := httplog.NewLogger("cms", httplog.Options{
		JSON: true,
	})

	allowedOrigins := []string{
		"http://localhost:8080",
		"https://chewedfeed.com",
		"https://www.chewedfeed.com",
	}
	if s.Config.Development {
		allowedOrigins = append(allowedOrigins, "http://*")
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-User-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})

	r := chi.NewRouter()

	r.Use(middleware.Heartbeat("/ping"))

	r.Route("/", func(r chi.Router) {
		r.Use(middleware.RequestID)
		r.Use(c.Handler)
		r.Use(bugMiddleware.BugFixes)
		r.Use(httplog.RequestLogger(logger))

		r.Get("/services", retro.NewCMS(s.Config).ServicesHandler)
		r.Get("/service/{service}", retro.NewCMS(s.Config).ServiceHandler)
	})

	r.Get("/health", healthcheck.HTTP)
	r.Get("/probe", probe.HTTP)

	bugLog.Local().Infof("listening on %d\n", s.Config.Local.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", s.Config.Local.Port), r); err != nil {
		return bugLog.Errorf("port failed: %+v", err)
	}

	return nil
}
