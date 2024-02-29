package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"gatekeeper/pkg/jwt_provider"
	"gatekeeper/pkg/sqlite_ext"
	"net/http"

	"braces.dev/errtrace"
	"github.com/georgysavva/scany/sqlscan"
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
	ApiKey     string `json:"apiKey" validate:"required"`
	ProofToken string `json:"proofToken" validate:"required"`
	Metadata   []byte `json:"metadata" validate:"-"`
}

const (
	MsgMetadataIsInvalid            = "Metadata is invalid"
	MsgApiKeyIsInvalid              = "Api key is invalid"
	MsgProofTokenIsInvalidOrExpired = "Proof token is invalid or has expired"
	MsgAccountAlreadyExists         = "Account already exists"
)

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

	// Check if api key exists
	var companyUuid string
	err = sqlscan.Get(c.Request().Context(), ct.DB, &companyUuid,
		"SELECT uuid FROM companies WHERE api_key = ?", req.ApiKey,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return NewHTTPError(http.StatusBadRequest, MsgApiKeyIsInvalid)
		}
		return errtrace.Errorf("failed to check if api key exists: %w", err)
	}

	// Unmarshal metadata
	metadata := map[string]any{}
	err = json.Unmarshal(req.Metadata, &metadata)
	if err != nil {
		return NewHTTPError(http.StatusBadRequest, MsgMetadataIsInvalid)
	}

	// Check if proof token is invalid or has expired and extract wallet address
	claims, err := ct.JwtProvider.GetClaims(req.ProofToken)
	if err != nil {
		return NewHTTPError(http.StatusBadRequest, MsgProofTokenIsInvalidOrExpired)
	}
	jwtExpiredAt, err := claims.GetExpirationTime()
	if err != nil {
		return errtrace.Errorf("failed to get expiration time from claims: %w", err)
	}
	if jwtExpiredAt == nil {
		return NewHTTPError(http.StatusBadRequest, MsgProofTokenIsInvalidOrExpired)
	}
	walletAddress, err := claims.GetSubject()
	if err != nil {
		return errtrace.Errorf("failed to get subject from claims: %w", err)
	}
	if len(walletAddress) == 0 {
		return NewHTTPError(http.StatusBadRequest, MsgProofTokenIsInvalidOrExpired)
	}

	// Create account
	accountUuid, err := uuid.NewV7()
	if err != nil {
		return errtrace.Errorf("failed to generate account uuid: %w", err)
	}
	_, err = ct.DB.ExecContext(c.Request().Context(),
		"INSERT INTO accounts (uuid, company_uuid, wallet_address, metadata) VALUES (?, ?, ?, ?)",
		accountUuid, companyUuid, walletAddress, req.Metadata,
	)
	if err != nil {
		if sqlite_ext.HasErrCode(err, sqlite3.SQLITE_CONSTRAINT_UNIQUE) {
			return NewHTTPError(http.StatusUnprocessableEntity, MsgAccountAlreadyExists)
		}
		return errtrace.Errorf("failed to create account: %w", err)
	}

	return errtrace.Wrap(c.NoContent(http.StatusNoContent))
}
