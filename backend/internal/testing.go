package internal

import (
	"bytes"
	"encoding/json"
	"gatekeeper/internal/server"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func SendTestRequest(t *testing.T, s server.Server, method, path string, body any) *httptest.ResponseRecorder {
	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(method, path, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()
	s.EchoInst.ServeHTTP(res, req)
	return res
}

func RelativePath(relativePath string) string {
	_, file, _, _ := runtime.Caller(1)
	folderPath := filepath.Dir(file)
	return folderPath + "/" + relativePath
}
