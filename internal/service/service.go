package service

import (
	"fmt"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
	bugMiddleware "github.com/bugfixes/go-bugfixes/middleware"
	cms "github.com/chewedfeed/soon-cms/internal/soon-cms"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog"
	ConfigBuilder "github.com/keloran/go-config"
	healthcheck "github.com/keloran/go-healthcheck"
	probe "github.com/keloran/go-probe"
	"net/http"
	"time"
)

type Service struct {
	Config       *ConfigBuilder.Config
	ErrorChannel chan error
}

func (s *Service) Start() error {
	go s.StartHTTP()
	return <-s.ErrorChannel
}

func (s *Service) StartHTTP() {
	bugLog.Local().Info("Starting CMS")

	logger := httplog.NewLogger("soon-cms", httplog.Options{
		JSON: true,
	})

	allowedOrigins := []string{
		"http://localhost:8080",
		"https://chewedfeed.com",
		"https://www.chewedfeed.com",
	}

	if s.Config.Local.Development {
		allowedOrigins = append(allowedOrigins, "http://*")
	}
	services, err := cms.NewCMS(s.Config, s.ErrorChannel).AllowedOrigins()
	if err != nil {
		s.ErrorChannel <- bugLog.Error(err)
	}
	allowedOrigins = append(allowedOrigins, services...)

	c := cors.Options{
		AllowOriginFunc:  s.ValidateOrigin,
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-User-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}

	r := chi.NewRouter()

	r.Use(middleware.Heartbeat("/ping"))

	r.Route("/", func(r chi.Router) {
		r.Use(cors.Handler(c))
		r.Use(middleware.RequestID)
		r.Use(bugMiddleware.BugFixes)
		r.Use(httplog.RequestLogger(logger))

		r.Get("/services", cms.NewCMS(s.Config, s.ErrorChannel).ServicesHandler)
		r.Get("/service/{service}", cms.NewCMS(s.Config, s.ErrorChannel).ServiceHandler)
		r.Get("/script", cms.NewCMS(s.Config, s.ErrorChannel).ScriptHandler)
	})

	r.Get("/health", healthcheck.HTTP)
	r.Get("/probe", probe.HTTP)

	bugLog.Local().Infof("listening on %d\n", s.Config.Local.HTTPPort)
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", s.Config.Local.HTTPPort),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		s.ErrorChannel <- bugLog.Errorf("port failed: %+v", err)
	}
}

func (s *Service) ValidateOrigin(r *http.Request, checkOrigin string) bool {
	if s.Config.Local.Development {
		return true
	}

	services, err := cms.NewCMS(s.Config, s.ErrorChannel).AllowedOrigins()
	if err != nil {
		s.ErrorChannel <- bugLog.Error(err)
		return false
	}

	for _, service := range services {
		if service == checkOrigin {
			return true
		}
	}

	bugLog.Local().Infof("Origin checking failed: %s", checkOrigin)
	return false
}
