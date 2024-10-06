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
	Metadata   any    `json:"metadata" validate:"-"`
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
	var companyId string
	err = sqlscan.Get(c.Request().Context(), ct.DB, &companyId,
		"SELECT id FROM companies WHERE api_key = ?", req.ApiKey,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return NewHTTPError(http.StatusBadRequest, MsgApiKeyIsInvalid)
		}
		return errtrace.Errorf("failed to check if api key exists: %w", err)
	}

	// Unmarshal metadata
	metadataOpt := sql.Null[[]byte]{}
	if req.Metadata != nil {
		metadataBytes, err := json.Marshal(req.Metadata)
		if err != nil {
			// This should never happend, because it should fail on request binding
			return NewHTTPError(http.StatusBadRequest, MsgMetadataIsInvalid)
		}
		metadataOpt = sql.Null[[]byte]{Valid: true, V: metadataBytes}
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
	_, err = ct.DB.ExecContext(c.Request().Context(),
		"INSERT INTO accounts (company_id, wallet_address, metadata) VALUES (?, ?, ?)",
		companyId, walletAddress, metadataOpt,
	)
	if err != nil {
		if sqlite_ext.HasErrCode(err, sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY) {
			return NewHTTPError(http.StatusUnprocessableEntity, MsgAccountAlreadyExists)
		}
		return errtrace.Errorf("failed to create account: %w", err)
	}

	return errtrace.Wrap(c.NoContent(http.StatusNoContent))
}
