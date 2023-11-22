package server_test

import (
	"gatekeeper/internal"
	"gatekeeper/internal/server"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChallengeController_Issue(t *testing.T) {
	s := server.NewServer(internal.NewTestInjector(t))
	res := internal.SendTestRequest(t, s,
		"POST", "/v1/challenges/issue", server.ChallengeController_IssueRequest{WalletAddress: "WalletAddress"},
	)
	require.Equal(t, http.StatusOK, res.Code)
}
