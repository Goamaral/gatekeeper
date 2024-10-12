package server

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"gatekeeper/internal/entity"
	"gatekeeper/pkg/jwt_provider"
	"net/http"
	"strconv"
	"strings"
	"time"

	"braces.dev/errtrace"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/georgysavva/scany/sqlscan"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/samber/do"
)

const ChallengeTokenLength uint = 16
const ChallengeMessagePrefix = "Authentication request\n"
const ChallengeValidDuration = 5 * time.Minute

type ChallengeController struct {
	DB          *sql.DB
	JwtProvider jwt_provider.Provider
}

func NewChallengeController(echoGrp *echo.Group, i *do.Injector) ChallengeController {
	ct := ChallengeController{
		DB:          do.MustInvoke[*sql.DB](i),
		JwtProvider: do.MustInvoke[jwt_provider.Provider](i),
	}

	challenges := echoGrp.Group("/challenges", newApiKeyMiddleware(i))
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
	req, err := bindAndValidate[ChallengeController_IssueRequest](c)
	if err != nil {
		return err
	}

	// Generate challenge token
	challengeToken, err := GenerateChallengeToken()
	if err != nil {
		return errtrace.Errorf("failed to generate challenge token: %w", err)
	}

	// Save challenge
	_, err = ct.DB.ExecContext(c.Request().Context(),
		"INSERT INTO challenges (wallet_address, token, expired_at) VALUES (?, ?, ?)",
		req.WalletAddress, challengeToken, time.Now().UTC().Add(ChallengeValidDuration),
	)
	if err != nil {
		return errtrace.Errorf("failed to save challenge: %w", err)
	}

	return errtrace.Wrap(
		c.JSON(http.StatusOK, ChallengeController_IssueResponse{
			Challenge: ChallengeMessagePrefix + challengeToken,
		}),
	)
}

type ChallengeController_VerifyRequest struct {
	Challenge string `json:"challenge" validate:"required"`
	Signature string `json:"signature" validate:"required"`
}

type ChallengeController_VerifyResponse struct {
	ProofToken string `json:"proofToken"`
}

const MsgChallengeDoesNotExistOrExpired = "Challenge does not exist or has expired"
const MsgSignatureInvalid = "Signature is invalid for given challenge"

func (ct ChallengeController) Verify(c echo.Context) error {
	req, err := bindAndValidate[ChallengeController_VerifyRequest](c)
	if err != nil {
		return err
	}

	// Extract challenge token and get associated wallet address
	challengeToken := strings.ReplaceAll(req.Challenge, ChallengeMessagePrefix, "")
	challenge := entity.Challenge{Token: challengeToken}
	err = sqlscan.Get(c.Request().Context(), ct.DB, &challenge,
		"SELECT id, wallet_address, expired_at FROM challenges WHERE token = ? LIMIT 1", challengeToken,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return NewHTTPError(http.StatusUnprocessableEntity, MsgChallengeDoesNotExistOrExpired)
		}
		return errtrace.Errorf("failed to get challenge: %w", err)
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
		return NewHTTPError(http.StatusUnprocessableEntity, MsgSignatureInvalid)
	}
	// https://eips.ethereum.org/EIPS/eip-155
	if signature[64] == 27 || signature[64] == 28 {
		signature[64] -= 27
	}
	publicKey, err := crypto.SigToPub(challengeHash, signature)
	if err != nil {
		return NewHTTPError(http.StatusUnprocessableEntity, MsgSignatureInvalid)
	}
	if crypto.PubkeyToAddress(*publicKey).Hex() != challenge.WalletAddress {
		return NewHTTPError(http.StatusUnprocessableEntity, MsgSignatureInvalid)
	}

	// Delete challenge
	_, err = ct.DB.ExecContext(c.Request().Context(),
		"DELETE FROM challenges WHERE id = ?", challenge.Id,
	)
	if err != nil {
		return errtrace.Errorf("failed to delete challenge (token: %s): %w", challengeToken, err)
	}

	// Generate proof token
	proofToken, err := ct.JwtProvider.GenerateSignedToken(jwt.RegisteredClaims{
		Subject:   challenge.WalletAddress,
		ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(5 * time.Minute)},
	})
	if err != nil {
		return errtrace.Errorf("failed to generate proof token: %w", err)
	}

	return errtrace.Wrap(
		c.JSON(http.StatusOK, ChallengeController_VerifyResponse{ProofToken: proofToken}),
	)
}
