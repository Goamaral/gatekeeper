package main

import (
	"gatekeeper/internal/server"
)

func main() {
	s := server.NewServer()
	err := s.Serve(":3000")
	if err != nil {
		panic(err)
	}
}
