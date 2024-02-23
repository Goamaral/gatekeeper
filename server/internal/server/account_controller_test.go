package server_test

import (
	"gatekeeper/internal"
	"gatekeeper/internal/server"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountController_Create(t *testing.T) {
	walletAddress, _ := internal.GenerateWalletAddress(t)
	email := "test@gatekeeper.com"

	newTest := func(testFn func(t *testing.T, s server.Server)) func(t *testing.T) {
		s := server.NewServer(internal.NewTestInjector(t), server.Config{Env: "test"})
		return func(t *testing.T) { testFn(t, s) }
	}

	sendReq := func(t *testing.T, s server.Server, proofToken, email string) *httptest.ResponseRecorder {
		return internal.SendTestRequest(
			t, s, http.MethodPost, "/v1/accounts",
			server.AccountController_CreateRequest{ProofToken: proofToken, Email: email},
		)
	}

	t.Run("Success", newTest(
		func(t *testing.T, s server.Server) {
			proofToken := internal.GenerateProofToken(
				t, s.AccountCtrl.JwtProvider,
				walletAddress,
				time.Now().Add(time.Minute),
			)
			res := sendReq(t, s, proofToken, email)
			require.Equal(t, http.StatusNoContent, res.Code)
		},
	))

	t.Run("ProofTokenInvalid", newTest(
		func(t *testing.T, s server.Server) {
			res := sendReq(t, s, "jiberish", email)
			require.Equal(t, http.StatusUnprocessableEntity, res.Code)
			body := internal.ReadBody[server.ErrorResponse](t, res.Body)
			assert.Equal(t, server.MsgProofTokenIsInvalidOrExpired, body.Error)
		},
	))

	t.Run("ProofTokenExpired", newTest(
		func(t *testing.T, s server.Server) {
			proofToken := internal.GenerateProofToken(
				t, s.AccountCtrl.JwtProvider,
				walletAddress,
				time.Now().Add(-time.Minute),
			)
			res := sendReq(t, s, proofToken, email)
			require.Equal(t, http.StatusUnprocessableEntity, res.Code)
			body := internal.ReadBody[server.ErrorResponse](t, res.Body)
			assert.Equal(t, server.MsgProofTokenIsInvalidOrExpired, body.Error)
		},
	))

	t.Run("ProofTokenWithoutExp", newTest(
		func(t *testing.T, s server.Server) {
			proofToken, err := s.AccountCtrl.JwtProvider.GenerateSignedToken(jwt.RegisteredClaims{Subject: walletAddress})
			require.NoError(t, err)

			res := sendReq(t, s, proofToken, email)
			require.Equal(t, http.StatusUnprocessableEntity, res.Code)
			body := internal.ReadBody[server.ErrorResponse](t, res.Body)
			assert.Equal(t, server.MsgProofTokenIsInvalidOrExpired, body.Error)
		},
	))

	t.Run("ProofTokenWithoutSub", newTest(
		func(t *testing.T, s server.Server) {
			proofToken, err := s.AccountCtrl.JwtProvider.GenerateSignedToken(
				jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute))},
			)
			require.NoError(t, err)

			res := sendReq(t, s, proofToken, email)
			require.Equal(t, http.StatusUnprocessableEntity, res.Code)
			body := internal.ReadBody[server.ErrorResponse](t, res.Body)
			assert.Equal(t, server.MsgProofTokenIsInvalidOrExpired, body.Error)
		},
	))

	t.Run("DuplicateWalletAddress", newTest(
		func(t *testing.T, s server.Server) {
			accountUuid, err := uuid.NewV7()
			require.NoError(t, err)
			apiKey, err := server.GenerateApiKey(accountUuid)
			require.NoError(t, err)
			_, err = s.AccountCtrl.DB.Exec(
				"INSERT INTO accounts (uuid, api_key, email, wallet_address) VALUES (?, ?, ?, ?)",
				accountUuid, apiKey, email, walletAddress,
			)
			require.NoError(t, err)

			proofToken := internal.GenerateProofToken(
				t, s.AccountCtrl.JwtProvider,
				walletAddress,
				time.Now().Add(time.Minute),
			)

			res := sendReq(t, s, proofToken, email)
			require.Equal(t, http.StatusUnprocessableEntity, res.Code)
			body := internal.ReadBody[server.ErrorResponse](t, res.Body)
			assert.Equal(t, server.MsgAccountAlreadyExists, body.Error)
		},
	))
}
