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
	if e.Err == nil {
		return http.StatusText(e.Code)
	}

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

var ErrBadRequest = NewHTTPError(http.StatusBadRequest, nil)
var ErrNotFound = NewHTTPError(http.StatusNotFound, nil)

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
	Echo          *echo.Echo
	ChallengeCtrl ChallengeController
	AccountCtrl   AccountController
}

func NewServer(i *do.Injector, config Config) Server {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.HTTPErrorHandler = func(err error, c echo.Context) {
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
			msg = map[string]any{"errors": e}
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
				e.Logger.Error(err)
			}
		}

		if httpErr.Code == http.StatusInternalServerError {
			e.Logger.Error(httpErr)
		}
	}

	v1 := e.Group("/v1")

	return Server{
		Config:        config,
		Echo:          e,
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
	return errtrace.Wrap(s.Echo.Server.Serve(lst))
}

func bindAndValidate[R any](c echo.Context) (R, error) {
	var req R
	err := c.Bind(&req)
	if err != nil {
		return req, ErrBadRequest
	}
	v := validate.Struct(req)
	if !v.Validate() {
		return req, NewValidationErrorResponse(v.Errors)
	}
	return req, nil
}
