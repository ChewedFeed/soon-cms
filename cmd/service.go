package main

import (
  "fmt"

  bugLog "github.com/bugfixes/go-bugfixes/logs"
  "github.com/chewedfeed/soon-cms/internal/config"
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

  cfg, err := config.Build()
  if err != nil {
    _ = bugLog.Errorf("config: %v", err)
    return
  }

  errChan := make(chan error)
  s := &service.Service{
    Config:       cfg,
    ErrorChannel: errChan,
  }

  if err := s.Start(); err != nil {
    _ = bugLog.Errorf("start service: %v", err)
    return
  }
}
