package server

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"gatekeeper/pkg/db"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
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
	echoGrp.POST("/validate", ct.Validate)

	return ct
}

type ChallengeController_IssueRequest struct {
	WalletAddress string `json:"walletAddress" validate:"required"`
}

type ChallengeController_IssueResponse struct {
	Challenge string `json:"challenge"`
}

func GenerateChallengeToken() (string, error) {
	challengeTokenBytes := make([]byte, ChallengeTokenLength)
	_, err := rand.Read(challengeTokenBytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(challengeTokenBytes), nil
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
	challengeToken, err := GenerateChallengeToken()
	if err != nil {
		ct.Logger.WithError(err).Error("failed to generate challenge token")
		return c.JSON(http.StatusInternalServerError, InternalServerErrorResponse)
	}

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

type ChallengeController_ValidateRequest struct {
	Challenge string `json:"challenge" validate:"required"`
	Signature string `json:"signature" validate:"required"`
}

type ChallengeController_ValidateResponse struct {
	Valid bool   `json:"challenge"`
	Error string `json:"error,omitempty"`
}

const MsgChallengeDoesNotExistOrExpired = "Challenge does not exist or has expired"
const MsgInvalidSignature = "Invalid signature for given challenge"

func (ct ChallengeController) Validate(c echo.Context) error {
	req := ChallengeController_ValidateRequest{}
	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, RequestMalformedResponse)
	}
	v := validate.Struct(req)
	if !v.Validate() {
		return c.JSON(http.StatusBadRequest, ValidationErrorResponse{Errors: v.Errors})
	}

	// Extract challenge token and get associated wallet address
	challengeToken := strings.ReplaceAll(req.Challenge, ChallengeMessagePrefix, "")
	var walletAddress string
	err = ct.DbProvider.DB.GetContext(c.Request().Context(), &walletAddress,
		"SELECT wallet_address FROM challenges WHERE token = ? LIMIT 1", challengeToken,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusUnprocessableEntity, ChallengeController_ValidateResponse{
				Error: MsgChallengeDoesNotExistOrExpired,
			})
		}

		ct.Logger.WithError(err).Error("failed to get challenge")
		return c.JSON(http.StatusInternalServerError, InternalServerErrorResponse)
	}

	// Validate message
	challengeHash := crypto.Keccak256([]byte(req.Challenge))
	signature, err := hexutil.Decode(req.Signature)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, ChallengeController_ValidateResponse{
			Error: MsgInvalidSignature,
		})
	}
	publicKey, err := crypto.SigToPub(challengeHash, signature)
	if err != nil || crypto.PubkeyToAddress(*publicKey).Hex() != walletAddress {
		return c.JSON(http.StatusUnprocessableEntity, ChallengeController_ValidateResponse{
			Error: MsgInvalidSignature,
		})
	}

	// Delete challenge
	_, err = ct.DbProvider.DB.ExecContext(c.Request().Context(),
		"DELETE FROM challenges WHERE token = ?", challengeToken,
	)
	if err != nil {
		ct.Logger.WithError(err).Errorf("failed to delete challenge (token: %s)", challengeToken)
		return c.JSON(http.StatusInternalServerError, InternalServerErrorResponse)
	}

	return c.JSON(http.StatusOK, ChallengeController_ValidateResponse{Valid: true})
}
