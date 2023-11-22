package server

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"gatekeeper/pkg/db"
	"net/http"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gookit/validate"
	"github.com/labstack/echo/v4"
	"github.com/samber/do"
	"github.com/sirupsen/logrus"
)

const ChallengeTokenLength uint = 10
const ChallengeMessagePrefix = "Login request\n"

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
		Challenge: ChallengeMessagePrefix + challengeToken,
	})
}

type ChallengeController_VerifyRequest struct {
	WalletAddress   string `json:"walletAddress" validate:"required"`
	SignedChallenge string `json:"signedChallenge" validate:"required"`
}

type ChallengeController_VerifyResponse struct {
	Valid bool   `json:"challenge"`
	Error string `json:"errors,omitempty"`
}

const MsgChallengeDoesNotExistOrExpired = "Challenge does not exist or has expired"
const MsgInvalidWalletAddressSignedChallengeCombination = "Invalid wallet address and signed challenge combination"

func (ct ChallengeController) Verify(c echo.Context) error {
	req := ChallengeController_VerifyRequest{}
	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, RequestMalformedResponse)
	}
	v := validate.Struct(req)
	if !v.Validate() {
		return c.JSON(http.StatusBadRequest, ValidationErrorResponse{Errors: v.Errors})
	}

	// Extract challenge token and get associated wallet address
	var challengeToken string
	err = ct.DbProvider.DB.GetContext(c.Request().Context(), &challengeToken,
		"SELECT token FROM challenges WHERE wallet_address = ? LIMIT 1", req.WalletAddress,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusUnprocessableEntity, ChallengeController_VerifyResponse{
				Error: MsgChallengeDoesNotExistOrExpired,
			})
		}

		ct.Logger.WithError(err).Error("failed to get challenge")
		return c.JSON(http.StatusInternalServerError, InternalServerErrorResponse)
	}

	// Verify message
	walletAddressBytes, err := crypto.Ecrecover([]byte(ChallengeMessagePrefix+challengeToken), []byte(req.SignedChallenge))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, ChallengeController_VerifyResponse{
			Error: MsgInvalidWalletAddressSignedChallengeCombination,
		})
	}
	if string(walletAddressBytes) != req.WalletAddress {
		return c.JSON(http.StatusUnprocessableEntity, ChallengeController_VerifyResponse{
			Error: MsgInvalidWalletAddressSignedChallengeCombination,
		})
	}

	return c.JSON(http.StatusOK, ChallengeController_VerifyResponse{Valid: true})
}
