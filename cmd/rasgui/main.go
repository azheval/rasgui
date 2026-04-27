package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"rasgui/internal/app"
	"rasgui/internal/config"
	"rasgui/internal/logging"
	"rasgui/internal/runner"
	"rasgui/internal/store"
	"rasgui/internal/web"
)

func main() {
	workdir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := config.Load(workdir)
	if err != nil {
		log.Fatal(err)
	}
	logger, closer, err := logging.BuildLogger(cfg.LogDir, cfg.LogLevel)
	if err != nil {
		log.Fatal(err)
	}
	defer closer.Close()

	st, err := store.Open(cfg.DataDir)
	if err != nil {
		logger.Error("store open failed", "error", err)
		log.Fatal(err)
	}
	defer st.Close()

	appSvc := app.New(st, runner.New(cfg), logger.With("layer", "application"))
	if err := appSvc.SeedAdmin(cfg.DefaultAdminUser, cfg.DefaultAdminPassword); err != nil {
		log.Fatal(err)
	}
	for _, item := range config.DiscoverLocalToolchains(cfg) {
		if err := appSvc.SeedToolchainProfile(item.Name, item.Version, item.RACPath, item.RASPath, item.Description, item.IsDefault); err != nil {
			log.Fatal(err)
		}
	}
	if err := appSvc.SeedConnectionProfile("default", cfg.DefaultRASHost, cfg.DefaultRASPort, "Default remote RAS profile"); err != nil {
		log.Fatal(err)
	}

	server, err := web.New(cfg, appSvc, logger.With("layer", "web"))
	if err != nil {
		log.Fatal(err)
	}

	addr := fmt.Sprintf(":%d", cfg.HTTPPort)
	logger.Info("rasgui started", "addr", addr, "http_port", cfg.HTTPPort, "log_dir", cfg.LogDir, "log_level", cfg.LogLevel)
	log.Fatal(http.ListenAndServe(addr, server.Handler()))
}
