package server_test

import (
	"gatekeeper/internal"
	"gatekeeper/internal/server"
	"gatekeeper/internal/test"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiKeyMiddleware(t *testing.T) {
	endpoints := []struct{ Method, Path string }{
		{Method: http.MethodPost, Path: "/v1/accounts"},
		{Method: http.MethodPost, Path: "/v1/challenges/issue"},
		{Method: http.MethodPost, Path: "/v1/challenges/verify"},
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.Method+" "+endpoint.Path, func(t *testing.T) {
			s := server.NewServer(internal.NewTestInjector(t), server.Config{Env: "test"})
			res := test.SendTestRequest(t, s, endpoint.Method, endpoint.Path, map[string]string{"Api-Key": "jiberish"}, nil)
			require.Equal(t, http.StatusBadRequest, res.Code)
			body := test.ReadBody[server.ErrorResponse](t, res.Body)
			assert.Equal(t, server.MsgApiKeyIsInvalid, body.Error)
		})
	}
}
