package main

import (
	"fmt"
	"gatekeeper/internal"
	"gatekeeper/internal/server"
	"log/slog"
	"os"
)

func panicOnErr(err error) {
	if err != nil {
		slog.Error("%s", err)
		os.Exit(1)
	}
}

func main() {
	injector := internal.NewInjector()
	defer injector.Shutdown()

	s := server.NewServer(injector)
	addr := ":3000"
	err := s.Serve(addr)
	panicOnErr(fmt.Errorf("failed to serve http server on %s: %w", addr, err))
}
