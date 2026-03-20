package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/caarlos0/env/v11"
	"github.com/chewedfeed/soon-cms/internal/service"
	flagsService "github.com/flags-gg/go-flags"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	ConfigBuilder "github.com/keloran/go-config"
	_ "github.com/lib/pq"
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
	logs.Logf("Starting %s version %s (build %s)", ServiceName, BuildVersion, BuildHash)

	c := ConfigBuilder.NewConfigNoVault()

	if err := c.Build(ConfigBuilder.Local, ConfigBuilder.Postgres, ConfigBuilder.WithProjectConfigurator(ProjectConfig{})); err != nil {
		logs.Fatalf("config: %v", err)
	}

	if err := migrateDB(c); err != nil {
		if shouldRequireDB(c) {
			logs.Fatalf("Failed to migrate db: %v", err)
		}
		logs.Errorf("Database unavailable during local startup, continuing without migrations: %v", err)
	}

	flags := flagsService.NewClient(flagsService.WithAuth(flagsService.Auth{
		ProjectID:     c.ProjectProperties["flags_project"].(string),
		AgentID:       c.ProjectProperties["flags_agent"].(string),
		EnvironmentID: c.ProjectProperties["flags_environment"].(string),
	}), flagsService.WithMemory())

	if err := service.New(c, flags).Start(); err != nil {
		logs.Fatalf("start service: %v", err)
	}
}

func shouldRequireDB(config *ConfigBuilder.Config) bool {
	onRailway, ok := config.ProjectProperties["on_railway"].(bool)
	return ok && onRailway
}

func migrateDB(config *ConfigBuilder.Config) error {
	db, err := sql.Open("postgres",
		fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			config.Database.User,
			config.Database.Password,
			config.Database.Host,
			config.Database.Port,
			config.Database.DBName))
	if err != nil {
		return logs.Errorf("Failed to connect to database: %v", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return logs.Errorf("Failed to create postgres driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		return logs.Errorf("Failed to create migration instance: %v", err)
	}

	if err := runMigrations(m, "migrations"); err != nil {
		return logs.Errorf("Failed to run migration: %v", err)
	}

	return nil
}

func runMigrations(m *migrate.Migrate, migrationsDir string) error {
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}

		var dirtyErr migrate.ErrDirty
		if !errors.As(err, &dirtyErr) {
			return err
		}

		if err := repairDirtyLatestMigration(m, migrationsDir, dirtyErr); err != nil {
			return err
		}

		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
	}

	return nil
}

func repairDirtyLatestMigration(m *migrate.Migrate, migrationsDir string, dirtyErr migrate.ErrDirty) error {
	latestVersion, err := latestMigrationVersion(migrationsDir)
	if err != nil {
		return err
	}

	if dirtyErr.Version != latestVersion {
		return dirtyErr
	}

	logs.Logf("Repairing dirty latest migration version %d", dirtyErr.Version)
	if err := m.Force(dirtyErr.Version); err != nil {
		return err
	}

	return nil
}

func latestMigrationVersion(migrationsDir string) (int, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return 0, err
	}

	latest := -1
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}

		prefix, _, found := strings.Cut(entry.Name(), "_")
		if !found {
			continue
		}

		version, err := strconv.Atoi(prefix)
		if err != nil {
			continue
		}

		if version > latest {
			latest = version
		}
	}

	if latest == -1 {
		return 0, errors.New("no migration files found")
	}

	return latest, nil
}
