package server

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"gatekeeper/internal/entity"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/georgysavva/scany/sqlscan"
	"github.com/gookit/validate"
	"github.com/labstack/echo/v4"
	"github.com/samber/do"
)

const ChallengeTokenLength uint = 16
const ChallengeMessagePrefix = "Login request\n"
const ChallengeValidDuration = 5 * time.Minute

type ChallengeController struct {
	DB *sql.DB
}

func NewChallengeController(echoGrp *echo.Group, i *do.Injector) ChallengeController {
	ct := ChallengeController{
		DB: do.MustInvoke[*sql.DB](i),
	}

	challenges := echoGrp.Group("/challenges")
	challenges.POST("/issue", ct.Issue)
	challenges.POST("/verify", ct.Verify)

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
		return ErrRequestMalformed
	}
	v := validate.Struct(req)
	if !v.Validate() {
		return NewValidationErrorResponse(v.Errors)
	}

	// Generate challenge token
	challengeToken, err := GenerateChallengeToken()
	if err != nil {
		return fmt.Errorf("failed to generate challenge token: %w", err)
	}

	// Save challenge
	_, err = ct.DB.ExecContext(c.Request().Context(),
		"INSERT INTO challenges (wallet_address, token, expired_at) VALUES (?, ?, ?)",
		req.WalletAddress, challengeToken, time.Now().UTC().Add(ChallengeValidDuration),
	)
	if err != nil {
		return fmt.Errorf("failed to save challenge: %w", err)
	}

	return c.JSON(http.StatusOK, ChallengeController_IssueResponse{
		Challenge: ChallengeMessagePrefix + challengeToken,
	})
}

type ChallengeController_VerifyRequest struct {
	Challenge string `json:"challenge" validate:"required"`
	Signature string `json:"signature" validate:"required"`
}

const MsgChallengeDoesNotExistOrExpired = "Challenge does not exist or has expired"
const MsgInvalidSignature = "Invalid signature for given challenge"

func (ct ChallengeController) Verify(c echo.Context) error {
	req := ChallengeController_VerifyRequest{}
	err := c.Bind(&req)
	if err != nil {
		return ErrRequestMalformed
	}
	v := validate.Struct(req)
	if !v.Validate() {
		return NewValidationErrorResponse(v.Errors)
	}

	// Extract challenge token and get associated wallet address
	challengeToken := strings.ReplaceAll(req.Challenge, ChallengeMessagePrefix, "")
	challenge := entity.Challenge{Token: challengeToken}
	err = sqlscan.Get(c.Request().Context(), ct.DB, &challenge,
		"SELECT id, wallet_address, expired_at FROM challenges WHERE token = ? LIMIT 1", challengeToken,
	)
	if err != nil {
		if sqlscan.NotFound(err) {
			return NewHTTPError(http.StatusUnprocessableEntity, MsgChallengeDoesNotExistOrExpired)
		}
		return fmt.Errorf("failed to get challenge: %w", err)
	}

	// Check if expired
	if challenge.ExpiredAt.Before(time.Now()) {
		return NewHTTPError(http.StatusUnprocessableEntity, MsgChallengeDoesNotExistOrExpired)
	}

	// Verify message
	// https://eips.ethereum.org/EIPS/eip-191
	challengeHash := crypto.Keccak256([]byte("\x19Ethereum Signed Message:\n" + strconv.Itoa(len(req.Challenge)) + req.Challenge))
	signature, err := hexutil.Decode(req.Signature)
	if err != nil {
		return NewHTTPError(http.StatusUnprocessableEntity, MsgInvalidSignature)
	}
	// https://eips.ethereum.org/EIPS/eip-155
	if signature[64] == 27 || signature[64] == 28 {
		signature[64] -= 27
	}
	publicKey, err := crypto.SigToPub(challengeHash, signature)
	if err != nil {
		return NewHTTPError(http.StatusUnprocessableEntity, MsgInvalidSignature)
	}
	if crypto.PubkeyToAddress(*publicKey).Hex() != challenge.WalletAddress {
		return NewHTTPError(http.StatusUnprocessableEntity, MsgInvalidSignature)
	}

	// Delete challenge
	_, err = ct.DB.ExecContext(c.Request().Context(),
		"DELETE FROM challenges WHERE id = ?", challenge.Id,
	)
	if err != nil {
		return fmt.Errorf("failed to delete challenge (token: %s): %w", challengeToken, err)
	}

	return c.NoContent(http.StatusNoContent)
}
