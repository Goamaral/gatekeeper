package server

import (
	"github.com/labstack/echo/v4"
)

type ChallengeController struct {
}

func NewChallengeController(echoGrp *echo.Group) ChallengeController {
	ct := ChallengeController{}

	echoGrp.POST("/issue", ct.Issue)
	echoGrp.POST("/verify", ct.Verify)

	return ct
}

func (ct ChallengeController) Issue(c echo.Context) error {
	// TODO
	return nil
}

func (ct ChallengeController) Verify(c echo.Context) error {
	// TODO
	return nil
}
