package echo_ext

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func SendTestRequest(t *testing.T, e *echo.Echo, method, path string, headers map[string]string, body any) *httptest.ResponseRecorder {
	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(method, path, bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	return res
}

func ReadBody[T any](t *testing.T, bodyBuffer *bytes.Buffer) T {
	bodyBytes, err := io.ReadAll(bodyBuffer)
	require.NoError(t, err)

	var body T
	require.NoError(t, json.Unmarshal(bodyBytes, &body))

	return body
}

func RunMiddleware(t *testing.T, handler echo.MiddlewareFunc, setupReq func(*http.Request)) error {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	setupReq(req)
	c := echo.New().NewContext(req, httptest.NewRecorder())
	return handler(func(c echo.Context) error { return nil })(c)
}
