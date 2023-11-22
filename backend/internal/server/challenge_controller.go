package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"gatekeeper/pkg/db"
	"net/http"

	"github.com/gookit/validate"
	"github.com/labstack/echo/v4"
	"github.com/samber/do"
	"github.com/sirupsen/logrus"
)

const ChallengeTokenLength uint = 10

type ChallengeController struct {
	DbProvider db.Provider
	Logger     *logrus.Logger
}

func NewChallengeController(echoGrp *echo.Group, i *do.Injector) ChallengeController {
	ct := ChallengeController{
		DbProvider: do.MustInvoke[db.Provider](i),
		Logger:     do.MustInvoke[*logrus.Logger](i),
	}

	echoGrp.POST("/issue", ct.Issue)
	echoGrp.POST("/verify", ct.Verify)

	return ct
}

type ChallengeController_IssueRequest struct {
	WalletAddress string `json:"walletAddress" validate:"required"`
}

type ChallengeController_IssueResponse struct {
	Challenge string `json:"challenge"`
}

func (ct ChallengeController) Issue(c echo.Context) error {
	req := ChallengeController_IssueRequest{}
	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, RequestMalformedResponse)
	}

	// Validate request
	v := validate.Struct(req)
	if !v.Validate() {
		return c.JSON(http.StatusBadRequest, ValidationErrorResponse{Errors: v.Errors})
	}

	// Generate challenge token
	challengeTokenBytes := make([]byte, ChallengeTokenLength)
	_, err = rand.Read(challengeTokenBytes)
	if err != nil {
		ct.Logger.WithError(err).Error("failed to generate challenge token")
		return c.JSON(http.StatusInternalServerError, InternalServerErrorResponse)
	}
	challengeToken := hex.EncodeToString(challengeTokenBytes)

	// Save challenge
	_, err = ct.DbProvider.DB.ExecContext(c.Request().Context(),
		"INSERT INTO challenges (wallet_address, token) VALUES (?, ?)",
		req.WalletAddress, challengeToken,
	)
	if err != nil {
		ct.Logger.WithError(err).Error("failed to save challenge")
		return c.JSON(http.StatusInternalServerError, InternalServerErrorResponse)
	}

	return c.JSON(http.StatusOK, ChallengeController_IssueResponse{
		Challenge: fmt.Sprintf("Login request\n%s", challengeToken),
	})
}

func (ct ChallengeController) Verify(c echo.Context) error {
	// TODO
	return nil
}
