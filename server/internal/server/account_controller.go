package server

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"gatekeeper/pkg/jwt_provider"
	"gatekeeper/pkg/sqlite_ext"
	"net/http"
	"strings"

	"braces.dev/errtrace"
	"github.com/google/uuid"
	"github.com/gookit/validate"
	"github.com/labstack/echo/v4"
	"github.com/samber/do"
	sqlite3 "modernc.org/sqlite/lib"
)

type AccountController struct {
	DB          *sql.DB
	JwtProvider jwt_provider.Provider
}

func NewAccountController(echoGrp *echo.Group, i *do.Injector) AccountController {
	ct := AccountController{
		DB:          do.MustInvoke[*sql.DB](i),
		JwtProvider: do.MustInvoke[jwt_provider.Provider](i),
	}

	accounts := echoGrp.Group("/accounts")
	accounts.POST("", ct.Create)

	return ct
}

type AccountController_CreateRequest struct {
	ProofToken string `json:"proofToken" validate:"required"`
	Email      string `json:"email" validate:"required|email"`
}

const MsgProofTokenIsInvalidOrExpired = "Proof token is invalid or has expired"
const MsgAccountAlreadyExists = "Account already exists"
const ApiKeySuffixLength = 16

func GenerateApiKey(accountUuid uuid.UUID) (string, error) {
	b := make([]byte, ApiKeySuffixLength)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	compactAccoutnUuid := strings.ReplaceAll(accountUuid.String(), "-", "")
	return compactAccoutnUuid + base64.URLEncoding.EncodeToString(b), nil
}

func (ct AccountController) Create(c echo.Context) error {
	req := AccountController_CreateRequest{}
	err := c.Bind(&req)
	if err != nil {
		return ErrRequestMalformed
	}
	v := validate.Struct(req)
	if !v.Validate() {
		return NewValidationErrorResponse(v.Errors)
	}

	// Check if proof token is invalid or has expired and extract wallet address
	claims, err := ct.JwtProvider.GetClaims(req.ProofToken)
	if err != nil {
		return NewHTTPError(http.StatusUnprocessableEntity, MsgProofTokenIsInvalidOrExpired)
	}
	jwtExpiredAt, err := claims.GetExpirationTime()
	if err != nil {
		return errtrace.Errorf("failed to get expiration time from claims: %w", err)
	}
	if jwtExpiredAt == nil {
		return NewHTTPError(http.StatusUnprocessableEntity, MsgProofTokenIsInvalidOrExpired)
	}
	walletAddress, err := claims.GetSubject()
	if err != nil {
		return errtrace.Errorf("failed to get subject from claims: %w", err)
	}
	if len(walletAddress) == 0 {
		return NewHTTPError(http.StatusUnprocessableEntity, MsgProofTokenIsInvalidOrExpired)
	}

	// Create api key and account
	accountUuid, err := uuid.NewV7()
	if err != nil {
		return errtrace.Errorf("failed to generate account uuid: %w", err)
	}
	apiKey, err := GenerateApiKey(accountUuid)
	if err != nil {
		return errtrace.Errorf("failed to generate api key: %w", err)
	}
	_, err = ct.DB.ExecContext(c.Request().Context(),
		"INSERT INTO accounts (uuid, api_key, email, wallet_address) VALUES (?, ?, ?, ?)",
		accountUuid, apiKey, req.Email, walletAddress,
	)
	if err != nil {
		if sqlite_ext.HasErrCode(err, sqlite3.SQLITE_CONSTRAINT_UNIQUE) {
			return NewHTTPError(http.StatusUnprocessableEntity, MsgAccountAlreadyExists)
		}
		return errtrace.Errorf("failed to create account: %w", err)
	}

	return errtrace.Wrap(c.NoContent(http.StatusNoContent))
}
