package server

import (
	"database/sql"
	"errors"
	"net/http"

	"braces.dev/errtrace"
	"github.com/georgysavva/scany/sqlscan"
	"github.com/labstack/echo/v4"
	"github.com/samber/do"
)

const MsgApiKeyIsInvalid = "Api key is invalid"

func newApiKeyMiddleware(i *do.Injector) echo.MiddlewareFunc {
	db := do.MustInvoke[*sql.DB](i)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			apiKey := c.Request().Header.Get("Api-Key")

			var companyId string
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
