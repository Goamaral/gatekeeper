package server

import (
	"gatekeeper/internal/view"

	"github.com/labstack/echo/v4"
)

type PublicController struct {
}

func NewPublicController(echoGrp *echo.Group) PublicController {
	ct := PublicController{}

	echoGrp.GET("/", ct.Index)

	return ct
}

func (ct PublicController) Index(c echo.Context) error {
	return view.IndexPage("HELLO").
		Render(c.Request().Context(), c.Response().Writer)
}
