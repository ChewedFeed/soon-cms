package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/bugfixes/go-bugfixes/middleware"
	cms "github.com/chewedfeed/soon-cms/internal/soon-cms"
	ConfigBuilder "github.com/keloran/go-config"
	"github.com/keloran/go-healthcheck"
	"github.com/keloran/go-probe"
	"net/http"
	"strconv"
	"time"
)

type Service struct {
	Config  *ConfigBuilder.Config
	Origins []string
}

func New(config *ConfigBuilder.Config) *Service {
	return &Service{
		Config: config,
	}
}

func (s *Service) Start() error {
	errChan := make(chan error)
	go s.StartHTTP(errChan)
	return <-errChan
}

func (s *Service) StartHTTP(errChan chan error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/services", cms.NewCMS(s.Config, errChan).ServicesHandler)
	mux.HandleFunc("/service/{service}", cms.NewCMS(s.Config, errChan).ServiceHandler)
	mux.HandleFunc("/script", cms.NewCMS(s.Config, errChan).ScriptHandler)
	mux.HandleFunc("/health", healthcheck.HTTP)
	mux.HandleFunc("/probe", probe.HTTP)

	mw := middleware.NewMiddleware(context.Background())
	mw.AddMiddleware(middleware.SetupLogger(middleware.Error).Logger)
	mw.AddMiddleware(middleware.RequestID)
	mw.AddMiddleware(middleware.Recoverer)
	mw.AddMiddleware(mw.CORS)
	mw.AddAllowedMethods("GET", "OPTIONS")
	mw.AddAllowedHeaders("Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-User-Token")
	mw.AddAllowedOrigins("http://localhost:8080", "https://chewedfeed.com", "https://www.chewedfeed.com")
	if s.Config.Local.Development {
		mw.AddAllowedOrigins("*")
	}

	port := s.Config.Local.HTTPPort
	if s.Config.ProjectProperties["railway_port"].(string) != "" && s.Config.ProjectProperties["on_railway"].(bool) {
		i, err := strconv.Atoi(s.Config.ProjectProperties["railway_port"].(string))
		if err != nil {
			_ = logs.Errorf("Failed to parse port: %v", err)
			return
		}
		port = i
	}

	logs.Logf("Starting HTTP on %d", s.Config.Local.HTTPPort)
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mw.Handler(mux),
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		TLSNextProto:      make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}
	errChan <- server.ListenAndServe()
}
