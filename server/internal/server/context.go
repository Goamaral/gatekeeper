package server

import "github.com/labstack/echo/v4"

type ContextKey string

const (
	ContextKey_CompanyId     ContextKey = "companyId"
	ContextKey_WalletAddress ContextKey = "walletAddress"
)

func setContextValue(c echo.Context, key ContextKey, value any) {
	c.Set(string(key), value)
}

func getContextValue[T any](c echo.Context, key ContextKey) T {
	return c.Get(string(key)).(T)
}
