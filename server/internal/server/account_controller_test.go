package server_test

import (
	"gatekeeper/internal"
	"gatekeeper/internal/server"
	"gatekeeper/internal/test"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountController_Create(t *testing.T) {
	walletAddress, _ := test.GenerateWalletAddress(t)
	metadata := []byte("{\"email\": \"client@gatekeeper.com\"}")

	newProofToken := func(t *testing.T, s server.Server, walletAddress string) string {
		return test.GenerateProofToken(
			t, s.AccountCtrl.JwtProvider,
			walletAddress,
			time.Now().Add(time.Minute),
		)
	}

	newTest := func(testFn func(t *testing.T, s server.Server)) func(t *testing.T) {
		s := server.NewServer(internal.NewTestInjector(t), server.Config{Env: "test"})
		return func(t *testing.T) { testFn(t, s) }
	}
	sendReq := func(t *testing.T, s server.Server, proofToken string) *httptest.ResponseRecorder {
		return test.SendTestRequest(
			t, s, http.MethodPost, "/v1/accounts", map[string]string{"Api-Key": test.ApiKey},
			server.AccountController_CreateRequest{ProofToken: proofToken, Metadata: metadata},
		)
	}

	t.Run("Success", newTest(
		func(t *testing.T, s server.Server) {
			res := sendReq(t, s, newProofToken(t, s, walletAddress))
			require.Equal(t, http.StatusNoContent, res.Code)
		},
	))

	t.Run("ProofTokenInvalid", newTest(
		func(t *testing.T, s server.Server) {
			res := sendReq(t, s, "jiberish")
			require.Equal(t, http.StatusBadRequest, res.Code)
			body := test.ReadBody[server.ErrorResponse](t, res.Body)
			assert.Equal(t, server.MsgProofTokenIsInvalidOrExpired, body.Error)
		},
	))

	t.Run("ProofTokenExpired", newTest(
		func(t *testing.T, s server.Server) {
			proofToken := test.GenerateProofToken(
				t, s.AccountCtrl.JwtProvider,
				walletAddress,
				time.Now().Add(-time.Minute),
			)

			res := sendReq(t, s, proofToken)
			require.Equal(t, http.StatusBadRequest, res.Code)
			body := test.ReadBody[server.ErrorResponse](t, res.Body)
			assert.Equal(t, server.MsgProofTokenIsInvalidOrExpired, body.Error)
		},
	))

	t.Run("ProofTokenWithoutExp", newTest(
		func(t *testing.T, s server.Server) {
			proofToken, err := s.AccountCtrl.JwtProvider.GenerateSignedToken(jwt.RegisteredClaims{Subject: walletAddress})
			require.NoError(t, err)

			res := sendReq(t, s, proofToken)
			require.Equal(t, http.StatusBadRequest, res.Code)
			body := test.ReadBody[server.ErrorResponse](t, res.Body)
			assert.Equal(t, server.MsgProofTokenIsInvalidOrExpired, body.Error)
		},
	))

	t.Run("ProofTokenWithoutSub", newTest(
		func(t *testing.T, s server.Server) {
			proofToken, err := s.AccountCtrl.JwtProvider.GenerateSignedToken(
				jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute))},
			)
			require.NoError(t, err)

			res := sendReq(t, s, proofToken)
			require.Equal(t, http.StatusBadRequest, res.Code)
			body := test.ReadBody[server.ErrorResponse](t, res.Body)
			assert.Equal(t, server.MsgProofTokenIsInvalidOrExpired, body.Error)
		},
	))

	t.Run("DuplicateWalletAddress", newTest(
		func(t *testing.T, s server.Server) {
			res := sendReq(t, s, newProofToken(t, s, test.WalletAddress))
			require.Equal(t, http.StatusUnprocessableEntity, res.Code)
			body := test.ReadBody[server.ErrorResponse](t, res.Body)
			assert.Equal(t, server.MsgAccountAlreadyExists, body.Error)
		},
	))
}
