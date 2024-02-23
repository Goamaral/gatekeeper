package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"

	"braces.dev/errtrace"
	"github.com/gookit/validate"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/samber/do"
)

func NewHTTPError(code int, err any) HTTPError {
	return HTTPError{Code: code, Err: err}
}

type HTTPError struct {
	Code int
	Err  any
}

func (e HTTPError) Error() string {
	switch err := e.Err.(type) {
	case error:
		return err.Error()
	case fmt.Stringer:
		return err.String()
	case string:
		return err
	default:
		panic(fmt.Sprintf("can't get error message from %T", err))
	}
}

var ErrRequestMalformed = NewHTTPError(http.StatusBadRequest, "Request malformed")

func NewValidationErrorResponse(errs validate.Errors) HTTPError {
	return NewHTTPError(http.StatusBadRequest, errs)
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type Config struct {
	Env  string `env:"ENV" env-default:"production"`
	Port uint   `env:"HTTP_PORT" env-default:"3000"`
}

type Server struct {
	Config        Config
	EchoInst      *echo.Echo
	ChallengeCtrl ChallengeController
	AccountCtrl   AccountController
}

func NewServer(i *do.Injector, config Config) Server {
	echoInst := echo.New()
	echoInst.Use(middleware.Logger())
	echoInst.Use(middleware.Recover())
	echoInst.HTTPErrorHandler = func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		httpErr, ok := err.(HTTPError)
		if !ok {
			httpErr = HTTPError{
				Code: http.StatusInternalServerError,
				Err:  http.StatusText(http.StatusInternalServerError),
			}
		}

		var msg any
		switch e := httpErr.Err.(type) {
		case json.Marshaler:
			msg = e
		case validate.Errors:
			msg = map[string]any{"errors": err}
		default:
			msg = ErrorResponse{Error: httpErr.Error()}
		}

		// Send response
		if !c.Response().Committed {
			if c.Request().Method == http.MethodHead {
				err = c.NoContent(httpErr.Code)
			} else {
				err = c.JSON(httpErr.Code, msg)
			}
			if err != nil {
				echoInst.Logger.Error(err)
			}
		}

		if httpErr.Code == http.StatusInternalServerError {
			echoInst.Logger.Error(httpErr)
		}
	}

	v1 := echoInst.Group("/v1")

	return Server{
		Config:        config,
		EchoInst:      echoInst,
		ChallengeCtrl: NewChallengeController(v1, i),
		AccountCtrl:   NewAccountController(v1, i),
	}
}

func (s Server) Serve() error {
	addr := fmt.Sprintf(":%d", s.Config.Port)
	lst, err := net.Listen("tcp", addr)
	if err != nil {
		return errtrace.Errorf("failed to listen to tcp port on %s: %w", addr, err)
	}
	log.Printf("Http server listening on %s", addr)
	return errtrace.Wrap(s.EchoInst.Server.Serve(lst))
}
