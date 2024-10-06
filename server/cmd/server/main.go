package main

import (
	"gatekeeper/internal"
	"gatekeeper/internal/server"
	"log/slog"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

func exitOnErr(msg string, err error) {
	if err != nil {
		slog.With("error", err.Error()).Error(msg)
		os.Exit(1)
	}
}

func main() {
	i := internal.NewInjector()
	defer i.Shutdown()

	var cfg server.Config
	err := cleanenv.ReadEnv(&cfg)
	exitOnErr("failed to read server config from env: %s", err)

	s := server.NewServer(i, cfg)
	err = s.Serve()
	exitOnErr("failed to serve http server", err)
}
