package server

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	EchoInst      *echo.Echo
	PublicCtrl    PublicController
	ChallengeCtrl ChallengeController
}

func NewServer() Server {
	echoInst := echo.New()
	echoInst.Use(middleware.Logger())
	echoInst.Use(middleware.Recover())

	return Server{
		EchoInst:      echoInst,
		PublicCtrl:    NewPublicController(echoInst.Group("")),
		ChallengeCtrl: NewChallengeController(echoInst.Group("/v1/challenges")),
	}
}

func (s Server) Serve(addr string) error {
	data, _ := json.MarshalIndent(s.EchoInst.Routes(), "", "  ")
	fmt.Println(string(data))
	lst, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen to tcp port: %w", err)
	}
	return s.EchoInst.Server.Serve(lst)
}
