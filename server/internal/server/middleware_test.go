package server_test

import (
	"gatekeeper/internal"
	"gatekeeper/internal/server"
	server_testing "gatekeeper/internal/server/testing"
	"gatekeeper/pkg/echo_ext"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_ApiKeyMiddleware(t *testing.T) {
	endpoints := []struct{ Method, Path string }{
		{Method: http.MethodPost, Path: "/v1/accounts"},
		{Method: http.MethodGet, Path: "/v1/accounts/" + server_testing.WalletAddress + "/metadata"},
		{Method: http.MethodPost, Path: "/v1/challenges/issue"},
		{Method: http.MethodPost, Path: "/v1/challenges/verify"},
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.Method+" "+endpoint.Path, func(t *testing.T) {
			s := server.NewServer(internal.NewTestInjector(t), server.Config{Env: "test"})
			res := echo_ext.SendTestRequest(
				t, s.Echo, endpoint.Method, endpoint.Path,
				map[string]string{"Api-Key": "jiberish"}, nil,
			)
			require.Equal(t, http.StatusBadRequest, res.Code)
			body := echo_ext.ReadBody[server.ErrorResponse](t, res.Body)
			assert.Equal(t, server.MsgApiKeyIsInvalid, body.Error)
		})
	}
}

func TestIntegration_ProofTokenyMiddleware(t *testing.T) {
	endpoints := []struct{ Method, Path string }{
		{Method: http.MethodPost, Path: "/v1/accounts"},
		{Method: http.MethodGet, Path: "/v1/accounts/" + server_testing.WalletAddress + "/metadata"},
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.Method+" "+endpoint.Path, func(t *testing.T) {
			s := server.NewServer(internal.NewTestInjector(t), server.Config{Env: "test"})
			res := echo_ext.SendTestRequest(
				t, s.Echo, endpoint.Method, endpoint.Path,
				map[string]string{"Api-Key": server_testing.ApiKey, "Proof-Token": "jiberish"}, nil,
			)
			require.Equal(t, http.StatusBadRequest, res.Code)
			body := echo_ext.ReadBody[server.ErrorResponse](t, res.Body)
			assert.Equal(t, server.MsgProofTokenIsInvalidOrExpired, body.Error)
		})
	}
}

func TestUnit_ProofTokenMiddleware(t *testing.T) {
	i := internal.NewTestInjector(t)
	handler := server.NewProofTokenMiddleware(i)
	expiredProofToken := server_testing.GenerateProofToken(t, i, server_testing.WalletAddress, time.Now().Add(-time.Minute))
	emptyProofToken := server_testing.GenerateProofToken(t, i, "", time.Now().Add(time.Minute))

	runTest := func(expectsErr bool, proofToken string) func(t *testing.T) {
		return func(t *testing.T) {
			err := echo_ext.RunMiddleware(t, handler, func(req *http.Request) {
				req.Header.Set("Proof-Token", proofToken)
			})
			if expectsErr {
				assert.Equal(t, server.NewHTTPError(http.StatusBadRequest, server.MsgProofTokenIsInvalidOrExpired), err)
			} else {
				assert.NoError(t, err)
			}
		}
	}

	t.Run("Invalid", runTest(true, "jiberish"))
	t.Run("Expired", runTest(true, expiredProofToken))
	t.Run("Empty", runTest(true, emptyProofToken))
}
