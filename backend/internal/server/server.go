package server

import (
	"fmt"
	"net"

	"github.com/gookit/validate"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/samber/do"
	"github.com/sirupsen/logrus"
)

type ErrorResponse struct {
	Error []string `json:"error"`
}

type ValidationErrorResponse struct {
	Errors validate.Errors `json:"errors"`
}

var RequestMalformedResponse = ErrorResponse{Error: []string{"Request Malformed"}}
var InternalServerErrorResponse = ErrorResponse{Error: []string{"Internal Server Error"}}

type Server struct {
	EchoInst      *echo.Echo
	PublicCtrl    PublicController
	ChallengeCtrl ChallengeController
	Logger        *logrus.Logger
}

func NewServer(i *do.Injector) Server {
	echoInst := echo.New()
	echoInst.Use(middleware.Logger())
	echoInst.Use(middleware.Recover())

	return Server{
		EchoInst:      echoInst,
		PublicCtrl:    NewPublicController(echoInst.Group("")),
		ChallengeCtrl: NewChallengeController(echoInst.Group("/v1/challenges"), i),
		Logger:        do.MustInvoke[*logrus.Logger](i),
	}
}

func (s Server) Serve(addr string) error {
	lst, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen to tcp port: %w", err)
	}
	s.Logger.Infof("Http server listening on %s", addr)
	return s.EchoInst.Server.Serve(lst)
}
