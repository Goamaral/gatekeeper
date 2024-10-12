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
	"github.com/labstack/echo/v4"
	"github.com/samber/do"
	sqlite3 "modernc.org/sqlite/lib"
)

const (
	MsgMetadataIsInvalid    = "Metadata is invalid"
	MsgAccountAlreadyExists = "Account already exists"
	MsgAccountDoesNotExist  = "Account does not exist"
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

	accounts := echoGrp.Group("/accounts", NewApiKeyMiddleware(i), NewProofTokenMiddleware(i))
	accounts.POST("", ct.Create)
	accounts.GET("/:walletAddress/metadata", ct.GetMetadata)

	return ct
}

type AccountController_CreateRequest struct {
	WalletAddress string `json:"walletAddress" validate:"required"`
	Metadata      []byte `json:"metadata" validate:"-"`
}

func (ct AccountController) Create(c echo.Context) error {
	req, err := bindAndValidate[AccountController_CreateRequest](c)
	if err != nil {
		return err
	}

	// Unmarshal metadata
	metadataOpt := sql.Null[[]byte]{}
	if req.Metadata != nil {
		metadata := map[string]any{}
		err := json.Unmarshal(req.Metadata, &metadata)
		if err != nil {
			return NewHTTPError(http.StatusBadRequest, MsgMetadataIsInvalid)
		}
		metadataOpt = sql.Null[[]byte]{Valid: true, V: req.Metadata}
	}

	// Create account
	if getContextValue[string](c, ContextKey_WalletAddress) != req.WalletAddress {
		return NewHTTPError(http.StatusBadRequest, MsgProofTokenIsInvalidOrExpired)
	}
	companyId := getContextValue[uint](c, ContextKey_CompanyId)
	_, err = ct.DB.ExecContext(c.Request().Context(),
		"INSERT INTO accounts (company_id, wallet_address, metadata) VALUES (?, ?, ?)",
		companyId, req.WalletAddress, metadataOpt,
	)
	if err != nil {
		if sqlite_ext.HasErrCode(err, sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY) {
			return NewHTTPError(http.StatusBadRequest, MsgAccountAlreadyExists)
		}
		return errtrace.Errorf("failed to create account: %w", err)
	}

	return errtrace.Wrap(c.NoContent(http.StatusNoContent))
}

type AccountController_GetMetadataResponse struct {
	Metadata map[string]any `json:"metadata"`
}

func (ct AccountController) GetMetadata(c echo.Context) error {
	walletAddress := c.Param("walletAddress")

	if getContextValue[string](c, ContextKey_WalletAddress) != walletAddress {
		return NewHTTPError(http.StatusBadRequest, MsgProofTokenIsInvalidOrExpired)
	}
	companyId := getContextValue[uint](c, ContextKey_CompanyId)

	var metadataBytes []byte
	err := sqlscan.Get(c.Request().Context(), ct.DB, &metadataBytes,
		"SELECT metadata FROM accounts WHERE company_id = ? AND wallet_address = ? LIMIT 1", companyId, walletAddress,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return errtrace.Errorf("failed to get account metadata: %w", err)
	}

	metadata := map[string]any{}
	err = json.Unmarshal(metadataBytes, &metadata)
	if err != nil {
		return errtrace.Errorf("failed to unmarshal metadata: %w", err)
	}

	return c.JSON(http.StatusOK, AccountController_GetMetadataResponse{Metadata: metadata})
}
