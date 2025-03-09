package main

import (
	"fmt"
	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/caarlos0/env/v11"
	"github.com/chewedfeed/soon-cms/internal/service"
	ConfigBuilder "github.com/keloran/go-config"
)

var (
	BuildVersion = "dev"
	BuildHash    = "unknown"
	ServiceName  = "base-service"
)

type ProjectConfig struct{}

func (pc ProjectConfig) Build(cfg *ConfigBuilder.Config) error {
	type FlagsService struct {
		ProjectID     string `env:"FLAGS_PROJECT_ID" envDefault:"flags-gg"`
		AgentID       string `env:"FLAGS_AGENT_ID" envDefault:"orchestrator"`
		EnvironmentID string `env:"FLAGS_ENVIRONMENT_ID" envDefault:"orchestrator"`
	}

	type PC struct {
		RailwayPort string `env:"PORT" envDefault:"3000"`
		OnRailway   bool   `env:"ON_RAILWAY" envDefault:"false"`
		Flags       FlagsService
	}
	p := PC{}

	if err := env.Parse(&p); err != nil {
		return logs.Errorf("Failed to parse services: %v", err)
	}
	if cfg.ProjectProperties == nil {
		cfg.ProjectProperties = make(map[string]interface{})
	}
	cfg.ProjectProperties["railway_port"] = p.RailwayPort
	cfg.ProjectProperties["on_railway"] = p.OnRailway

	cfg.ProjectProperties["flags_agent"] = p.Flags.AgentID
	cfg.ProjectProperties["flags_environment"] = p.Flags.EnvironmentID
	cfg.ProjectProperties["flags_project"] = p.Flags.ProjectID

	return nil
}

func main() {
	logs.Info(fmt.Sprintf("Starting %s", ServiceName))
	logs.Info(fmt.Sprintf("Version: %s, Hash: %s", BuildVersion, BuildHash))

	c := ConfigBuilder.NewConfigNoVault()

	if err := c.Build(ConfigBuilder.Local, ConfigBuilder.Database, ConfigBuilder.WithProjectConfigurator(ProjectConfig{})); err != nil {
		logs.Fatalf("config: %v", err)
	}

	if err := service.New(c).Start(); err != nil {
		logs.Fatalf("start service: %v", err)
	}
}
