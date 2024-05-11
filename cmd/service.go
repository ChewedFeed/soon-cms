package main

import (
	"fmt"
	ConfigBuilder "github.com/keloran/go-config"
	ConfigVault "github.com/keloran/go-config/vault"
	vault_helper "github.com/keloran/vault-helper"

	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"github.com/chewedfeed/soon-cms/internal/service"
)

var (
	BuildVersion = "dev"
	BuildHash    = "unknown"
	ServiceName  = "base-service"
)

func main() {
	bugLog.Local().Info(fmt.Sprintf("Starting %s", ServiceName))
	bugLog.Local().Info(fmt.Sprintf("Version: %s, Hash: %s", BuildVersion, BuildHash))

	vh := vault_helper.NewVault("", "")
	c := ConfigBuilder.NewConfig(vh)
	c.VaultPaths = ConfigVault.Paths{
		Database: ConfigVault.Path{
			Credentials: "database/creds/chewedfeed_coming_soon-database-role",
			Details:     "kv/data/chewedfeed/coming_soon",
		},
	}
	if err := c.Build(ConfigBuilder.Local, ConfigBuilder.Vault, ConfigBuilder.Database); err != nil {
		bugLog.Local().Fatalf("config: %v", err)
	}

	errChan := make(chan error)
	s := &service.Service{
		Config:       c,
		ErrorChannel: errChan,
	}

	if err := s.Start(); err != nil {
		bugLog.Local().Fatalf("start service: %v", err)
	}
}
