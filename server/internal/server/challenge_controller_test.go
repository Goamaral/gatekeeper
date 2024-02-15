package server_test

import (
	"context"
	"gatekeeper/internal"
	"gatekeeper/internal/server"
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
	res := internal.SendTestRequest(
		t, s, http.MethodPost, "/v1/challenges/issue",
		server.ChallengeController_IssueRequest{WalletAddress: "WalletAddress"},
	)
	require.Equal(t, http.StatusOK, res.Code)
}

func TestChallengeController_Verify(t *testing.T) {
	walletAddressA, privateKeyA := internal.GenerateWalletAddress(t)
	challengeTokenA, err := server.GenerateChallengeToken()
	require.NoError(t, err)
	challengeA := server.ChallengeMessagePrefix + challengeTokenA
	signatureA, err := crypto_ext.PersonalSign([]byte(challengeA), privateKeyA)
	require.NoError(t, err)

	_, privateKeyB := internal.GenerateWalletAddress(t)
	challengeTokenB, err := server.GenerateChallengeToken()
	require.NoError(t, err)
	challengeB := server.ChallengeMessagePrefix + challengeTokenB
	signatureB, err := crypto_ext.PersonalSign([]byte(challengeB), privateKeyB)
	require.NoError(t, err)

	sendReq := func(t *testing.T, challenge, signature string, expiredAt time.Time) *httptest.ResponseRecorder {
		s := server.NewServer(internal.NewTestInjector(t), server.Config{Env: "test"})

		_, err = s.ChallengeCtrl.DB.ExecContext(context.Background(),
			"INSERT INTO challenges (wallet_address, token, expired_at) VALUES (?, ?, ?)",
			walletAddressA, challengeTokenA, expiredAt,
		)
		require.NoError(t, err)

		return internal.SendTestRequest(
			t, s, http.MethodPost, "/v1/challenges/verify",
			server.ChallengeController_VerifyRequest{Challenge: challenge, Signature: signature},
		)
	}

	t.Run("Success", func(t *testing.T) {
		res := sendReq(t, challengeA, hexutil.Encode(signatureA), time.Now().UTC().Add(time.Minute))
		require.Equal(t, http.StatusNoContent, res.Code)
	})

	t.Run("ChallengeDoesNotExist", func(t *testing.T) {
		res := sendReq(t, challengeB, hexutil.Encode(signatureB), time.Now().UTC().Add(time.Minute))
		require.Equal(t, http.StatusUnprocessableEntity, res.Code)
		body := internal.ReadBody[server.ErrorResponse](t, res.Body)
		assert.Equal(t, server.MsgChallengeDoesNotExistOrExpired, body.Error)
	})

	t.Run("ChallengeExpired", func(t *testing.T) {
		res := sendReq(t, challengeA, hexutil.Encode(signatureA), time.Now().UTC().Add(-time.Minute))
		require.Equal(t, http.StatusUnprocessableEntity, res.Code)
		body := internal.ReadBody[server.ErrorResponse](t, res.Body)
		assert.Equal(t, server.MsgChallengeDoesNotExistOrExpired, body.Error)
	})

	t.Run("InvalidSignature", func(t *testing.T) {
		res := sendReq(t, challengeA, hexutil.Encode(signatureB), time.Now().UTC().Add(time.Minute))
		require.Equal(t, http.StatusUnprocessableEntity, res.Code)
		body := internal.ReadBody[server.ErrorResponse](t, res.Body)
		assert.Equal(t, server.MsgInvalidSignature, body.Error)
	})
}
