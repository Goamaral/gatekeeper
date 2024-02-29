package server_test

import (
	"gatekeeper/internal"
	"gatekeeper/internal/server"
	"gatekeeper/internal/test"
	"gatekeeper/pkg/crypto_ext"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChallengeController_Issue(t *testing.T) {
	s := server.NewServer(internal.NewTestInjector(t), server.Config{Env: "test"})
	res := test.SendTestRequest(
		t, s, http.MethodPost, "/v1/challenges/issue",
		server.ChallengeController_IssueRequest{WalletAddress: "WalletAddress"},
	)
	require.Equal(t, http.StatusOK, res.Code)
}

func TestChallengeController_Verify(t *testing.T) {
	walletAddressA, privateKeyA := test.GenerateWalletAddress(t)
	challengeTokenA, err := server.GenerateChallengeToken()
	require.NoError(t, err)
	challengeA := server.ChallengeMessagePrefix + challengeTokenA
	signatureA, err := crypto_ext.PersonalSign([]byte(challengeA), privateKeyA)
	require.NoError(t, err)

	_, privateKeyB := test.GenerateWalletAddress(t)
	challengeTokenB, err := server.GenerateChallengeToken()
	require.NoError(t, err)
	challengeB := server.ChallengeMessagePrefix + challengeTokenB
	signatureB, err := crypto_ext.PersonalSign([]byte(challengeB), privateKeyB)
	require.NoError(t, err)

	type Test struct {
		ExpiredAt time.Time
	}

	newTest := func(test Test, testFn func(t *testing.T, s server.Server)) func(t *testing.T) {
		s := server.NewServer(internal.NewTestInjector(t), server.Config{Env: "test"})

		_, err = s.ChallengeCtrl.DB.Exec(
			"INSERT INTO challenges (wallet_address, token, expired_at) VALUES (?, ?, ?)",
			walletAddressA, challengeTokenA, test.ExpiredAt,
		)
		require.NoError(t, err)

		return func(t *testing.T) { testFn(t, s) }
	}

	sendReq := func(t *testing.T, s server.Server, challenge, signature string) *httptest.ResponseRecorder {
		return test.SendTestRequest(
			t, s, http.MethodPost, "/v1/challenges/verify",
			server.ChallengeController_VerifyRequest{Challenge: challenge, Signature: signature},
		)
	}

	t.Run("Success", newTest(
		Test{ExpiredAt: time.Now().UTC().Add(time.Minute)},
		func(t *testing.T, s server.Server) {
			res := sendReq(t, s, challengeA, hexutil.Encode(signatureA))
			require.Equal(t, http.StatusOK, res.Code)
			body := test.ReadBody[server.ChallengeController_VerifyResponse](t, res.Body)

			claims, err := s.ChallengeCtrl.JwtProvider.GetClaims(body.ProofToken)
			require.NoError(t, err)
			sub, err := claims.GetSubject()
			require.NoError(t, err)
			assert.Equal(t, walletAddressA, sub)
			expiredAt, err := claims.GetExpirationTime()
			require.NoError(t, err)
			assert.Greater(t, expiredAt.Time, time.Now())
		},
	))

	t.Run("ChallengeDoesNotExist", newTest(
		Test{ExpiredAt: time.Now().UTC().Add(time.Minute)},
		func(t *testing.T, s server.Server) {
			res := sendReq(t, s, challengeB, hexutil.Encode(signatureB))
			require.Equal(t, http.StatusUnprocessableEntity, res.Code)
			body := test.ReadBody[server.ErrorResponse](t, res.Body)
			assert.Equal(t, server.MsgChallengeDoesNotExistOrExpired, body.Error)
		},
	))

	t.Run("ChallengeExpired", newTest(
		Test{ExpiredAt: time.Now().UTC().Add(-time.Minute)},
		func(t *testing.T, s server.Server) {
			res := sendReq(t, s, challengeA, hexutil.Encode(signatureA))
			require.Equal(t, http.StatusUnprocessableEntity, res.Code)
			body := test.ReadBody[server.ErrorResponse](t, res.Body)
			assert.Equal(t, server.MsgChallengeDoesNotExistOrExpired, body.Error)
		},
	))

	t.Run("InvalidSignature", newTest(
		Test{ExpiredAt: time.Now().UTC().Add(time.Minute)},
		func(t *testing.T, s server.Server) {
			res := sendReq(t, s, challengeA, hexutil.Encode(signatureB))
			require.Equal(t, http.StatusUnprocessableEntity, res.Code)
			body := test.ReadBody[server.ErrorResponse](t, res.Body)
			assert.Equal(t, server.MsgSignatureInvalid, body.Error)
		},
	))
}
