package service

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/bugfixes/go-bugfixes/middleware"
	cms "github.com/chewedfeed/soon-cms/internal/soon-cms"
	flagsService "github.com/flags-gg/go-flags"
	ConfigBuilder "github.com/keloran/go-config"
	"github.com/keloran/go-healthcheck"
	"github.com/keloran/go-probe"
)

type Service struct {
	Config *ConfigBuilder.Config
	Flags  *flagsService.Client
}

func New(config *ConfigBuilder.Config, flags *flagsService.Client) *Service {
	return &Service{
		Config: config,
		Flags:  flags,
	}
}

func (s *Service) Start() error {
	errChan := make(chan error)
	go s.StartHTTP(errChan)
	return <-errChan
}

func (s *Service) StartHTTP(errChan chan error) {
	mux := http.NewServeMux()
	c := cms.NewCMS(s.Config, s.Flags, errChan)
	// Read endpoints
	mux.HandleFunc("GET /services", c.ServicesHandler)
	mux.HandleFunc("GET /service/{service}", c.ServiceHandler)
	mux.HandleFunc("GET /script", c.ScriptHandler)
	mux.HandleFunc("GET /health", healthcheck.HTTP)
	mux.HandleFunc("GET /probe", probe.HTTP)

	// Write endpoints - services
	mux.HandleFunc("POST /service", c.CreateServiceHandler)
	mux.HandleFunc("PUT /service/{service}", c.UpdateServiceHandler)
	mux.HandleFunc("DELETE /service/{service}", c.DeleteServiceHandler)

	// Write endpoints - links
	mux.HandleFunc("POST /service/{service}/links", c.CreateLinkHandler)
	mux.HandleFunc("DELETE /service/{service}/links/{id}", c.DeleteLinkHandler)

	// Write endpoints - roadmap
	mux.HandleFunc("POST /service/{service}/roadmap", c.CreateRoadmapHandler)
	mux.HandleFunc("PUT /service/{service}/roadmap/{id}", c.UpdateRoadmapHandler)
	mux.HandleFunc("DELETE /service/{service}/roadmap/{id}", c.DeleteRoadmapHandler)

	// Write endpoints - launch tasks
	mux.HandleFunc("GET /service/{service}/tasks", c.ListTasksHandler)
	mux.HandleFunc("POST /service/{service}/tasks", c.CreateTaskHandler)
	mux.HandleFunc("PUT /service/{service}/tasks/{id}", c.UpdateTaskHandler)
	mux.HandleFunc("DELETE /service/{service}/tasks/{id}", c.DeleteTaskHandler)

	mw := middleware.NewMiddleware()
	mw.AddMiddleware(middleware.SetupLogger(middleware.Error).Logger)
	mw.AddMiddleware(middleware.RequestID)
	mw.AddMiddleware(middleware.Recoverer)
	mw.AddMiddleware(mw.CORS)
	mw.AddAllowedMethods("GET", "POST", "PUT", "DELETE", "OPTIONS")
	mw.AddAllowedHeaders("Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-User-Token")
	mw.AddAllowedOrigins("*", "https://chewedfeed.com", "https://www.chewedfeed.com", "https://admin.chewedfeed.com")

	port := s.Config.Local.HTTPPort
	if s.Config.ProjectProperties["railway_port"].(string) != "" && s.Config.ProjectProperties["on_railway"].(bool) {
		i, err := strconv.Atoi(s.Config.ProjectProperties["railway_port"].(string))
		if err != nil {
			_ = logs.Errorf("Failed to parse port: %v", err)
			return
		}
		port = i
	}

	logs.Logf("Starting HTTP on %d", port)
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
