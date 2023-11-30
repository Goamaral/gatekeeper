package main

import (
	"fmt"
	"gatekeeper/internal"
	"gatekeeper/internal/server"
	"log/slog"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

func panicOnErr(msg string, err error) {
	if err != nil {
		slog.Error("%s", fmt.Errorf("%s: %w", msg, err))
		os.Exit(1)
	}
}

func main() {
	injector := internal.NewInjector()
	defer injector.Shutdown()

	var serverConfig server.Config
	err := cleanenv.ReadEnv(&serverConfig)
	panicOnErr("failed to read server config from env: %s", err)

	s := server.NewServer(injector, serverConfig)
	err = s.Serve()
	panicOnErr("failed to serve http server", err)
}
