package server_test

import (
	"gatekeeper/internal"
	"gatekeeper/internal/server"
	server_testing "gatekeeper/internal/server/testing"
	"gatekeeper/pkg/echo_ext"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	// "github.com/golang-jwt/jwt/v5"

	"github.com/labstack/echo/v4"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountController_Create(t *testing.T) {
	newProofToken := func(t *testing.T, i *do.Injector, walletAddress string) string {
		return server_testing.GenerateProofToken(
			t, i,
			walletAddress,
			time.Now().Add(time.Minute),
		)
	}
	walletAddress, _ := server_testing.GenerateWalletAddress(t)
	metadata := []byte("{\"email\": \"client@gatekeeper.com\"}")

	newTest := func(testFn func(t *testing.T, i *do.Injector, s server.Server)) func(t *testing.T) {
		i := internal.NewTestInjector(t)
		s := server.NewServer(i, server.Config{Env: "test"})
		return func(t *testing.T) { testFn(t, i, s) }
	}
	sendReq := func(t *testing.T, e *echo.Echo, proofToken string, walletAddress string, metadata []byte) *httptest.ResponseRecorder {
		return echo_ext.SendTestRequest(
			t, e, http.MethodPost, "/v1/accounts",
			map[string]string{"Api-Key": server_testing.ApiKey, "Proof-Token": proofToken},
			map[string]any{"walletAddress": walletAddress, "metadata": metadata},
		)
	}

	t.Run("Success", newTest(
		func(t *testing.T, i *do.Injector, s server.Server) {
			res := sendReq(t, s.Echo, newProofToken(t, i, walletAddress), walletAddress, metadata)
			require.Equal(t, http.StatusNoContent, res.Code)
		},
	))

	t.Run("MetadataIsInvalid", newTest(
		func(t *testing.T, i *do.Injector, s server.Server) {
			res := sendReq(t, s.Echo, newProofToken(t, i, walletAddress), walletAddress, []byte("jiberish"))
			require.Equal(t, http.StatusBadRequest, res.Code)
			body := echo_ext.ReadBody[server.ErrorResponse](t, res.Body)
			assert.Equal(t, server.MsgMetadataIsInvalid, body.Error)
		},
	))

	t.Run("ProofTokenWalletAddressDoesNotMatch", newTest(
		func(t *testing.T, i *do.Injector, s server.Server) {
			res := sendReq(t, s.Echo, newProofToken(t, i, walletAddress), server_testing.WalletAddress, metadata)
			require.Equal(t, http.StatusBadRequest, res.Code)
			body := echo_ext.ReadBody[server.ErrorResponse](t, res.Body)
			assert.Equal(t, server.MsgProofTokenIsInvalidOrExpired, body.Error)
		},
	))

	t.Run("AccountAlreadyExists", newTest(
		func(t *testing.T, i *do.Injector, s server.Server) {
			server_testing.CreateAccount(t, i, 1, metadata)
			res := sendReq(t, s.Echo, newProofToken(t, i, server_testing.WalletAddress), server_testing.WalletAddress, metadata)
			require.Equal(t, http.StatusBadRequest, res.Code)
			body := echo_ext.ReadBody[server.ErrorResponse](t, res.Body)
			assert.Equal(t, server.MsgAccountAlreadyExists, body.Error)
		},
	))
}
