package server

import (
	"database/sql"
	"errors"
	"gatekeeper/pkg/jwt_provider"
	"net/http"

	"braces.dev/errtrace"
	"github.com/georgysavva/scany/sqlscan"
	"github.com/labstack/echo/v4"
	"github.com/samber/do"
)

const (
	MsgApiKeyIsInvalid              = "Api key is invalid"
	MsgProofTokenIsInvalidOrExpired = "Proof token is invalid or has expired"
)

func NewApiKeyMiddleware(i *do.Injector) echo.MiddlewareFunc {
	db := do.MustInvoke[*sql.DB](i)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			apiKey := c.Request().Header.Get("Api-Key")

			// Check if api key exists and extract company id
			var companyId uint
			err := sqlscan.Get(c.Request().Context(), db, &companyId,
				"SELECT id FROM companies WHERE api_key = ?", apiKey,
			)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return NewHTTPError(http.StatusBadRequest, MsgApiKeyIsInvalid)
				}
				return errtrace.Errorf("failed to check if api key exists: %w", err)
			}

			setContextValue(c, ContextKey_CompanyId, companyId)

			return next(c)
		}
	}
}

func NewProofTokenMiddleware(i *do.Injector) echo.MiddlewareFunc {
	jwtProvider := do.MustInvoke[jwt_provider.Provider](i)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if proof token is invalid or has expired and extract wallet address
			claims, err := jwtProvider.GetClaims(c.Request().Header.Get("Proof-Token"))
			if err != nil {
				return NewHTTPError(http.StatusBadRequest, MsgProofTokenIsInvalidOrExpired)
			}
			walletAddress, err := claims.GetSubject()
			if err != nil {
				return errtrace.Errorf("failed to get subject from claims: %w", err)
			}
			if len(walletAddress) == 0 {
				return NewHTTPError(http.StatusBadRequest, MsgProofTokenIsInvalidOrExpired)
			}

			setContextValue(c, ContextKey_WalletAddress, walletAddress)

			return next(c)
		}
	}
}
