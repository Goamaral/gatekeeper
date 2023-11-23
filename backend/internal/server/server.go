package server

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gookit/validate"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/samber/do"
	"github.com/sirupsen/logrus"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

var ErrRequestMalformed = echo.NewHTTPError(http.StatusBadRequest, "Request malformed")

type ErrorsResponse struct {
	Errors []string `json:"errors"`
}

type ValidationErrorResponse struct {
	Errors validate.Errors `json:"errors"`
}

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
	echoInst.HTTPErrorHandler = func(err error, c echo.Context) {
		statusCode := http.StatusInternalServerError
		errorMsg := http.StatusText(statusCode)

		if httpErr, ok := err.(*echo.HTTPError); ok {
			statusCode = httpErr.Code
			errorMsg = httpErr.Message.(string)
		}

		if statusCode == http.StatusInternalServerError {
			c.Logger().Error(err)
			errorMsg = http.StatusText(statusCode)
		}

		contentType := c.Request().Header.Get(echo.HeaderContentType)
		if contentType == echo.MIMETextHTML {
			err = c.String(statusCode, errorMsg) // TODO: Have error page for each status code
		} else {
			err = c.JSON(statusCode, ErrorResponse{Error: errorMsg})
		}
		if err != nil {
			c.Logger().Error(fmt.Errorf("failed to send error response: %w", err))
		}
	}

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
