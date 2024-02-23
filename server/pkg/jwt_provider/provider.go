package jwt_provider

import (
	"crypto/ecdsa"
	"fmt"
	"io"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/samber/do"
	"github.com/stretchr/testify/require"
)

type Provider struct {
	PrivKey *ecdsa.PrivateKey
	PubKey  *ecdsa.PublicKey
}

func NewProvider(privKeyReader io.Reader, pubKeyReader io.Reader) (Provider, error) {
	privKeyBytes, err := io.ReadAll(privKeyReader)
	if err != nil {
		return Provider{}, fmt.Errorf("failed to read private key: %w", err)
	}
	pubKeyBytes, err := io.ReadAll(pubKeyReader)
	if err != nil {
		return Provider{}, fmt.Errorf("failed to read public key: %w", err)
	}

	privKey, err := jwt.ParseECPrivateKeyFromPEM(privKeyBytes)
	if err != nil {
		return Provider{}, fmt.Errorf("failed to parse private key: %w", err)
	}

	pubKey, err := jwt.ParseECPublicKeyFromPEM(pubKeyBytes)
	if err != nil {
		return Provider{}, fmt.Errorf("failed to parse public key: %w", err)
	}

	return Provider{PrivKey: privKey, PubKey: pubKey}, nil
}

func NewTestProvider(t *testing.T) Provider {
	privKeyBytes := []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIMPu+380curPEbzB5FmrrOAr6Th4ZmrbQfKmG1HvR4EBoAoGCCqGSM49
AwEHoUQDQgAEwgUlhc3KO/HMScHd8tzo9mX2eHKxLRY1mhTXLXsf/nmXddkJO6AV
35UALafcg5Pq0jLVAx90EPM26ANGzaMJEA==
-----END EC PRIVATE KEY-----
`)
	pubKeyBytes := []byte(`-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEwgUlhc3KO/HMScHd8tzo9mX2eHKx
LRY1mhTXLXsf/nmXddkJO6AV35UALafcg5Pq0jLVAx90EPM26ANGzaMJEA==
-----END PUBLIC KEY-----
`)

	privKey, err := jwt.ParseECPrivateKeyFromPEM(privKeyBytes)
	require.NoError(t, err)

	pubKey, err := jwt.ParseECPublicKeyFromPEM(pubKeyBytes)
	require.NoError(t, err)

	return Provider{PrivKey: privKey, PubKey: pubKey}
}

func InjectTestProvider(t *testing.T) func(i *do.Injector) (Provider, error) {
	return func(i *do.Injector) (Provider, error) { return NewTestProvider(t), nil }
}

func (p Provider) GenerateSignedToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	return token.SignedString(p.PrivKey)
}

func (p Provider) GetClaims(signedToken string) (jwt.Claims, error) {
	token, err := jwt.Parse(signedToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method (alg: %v)", t.Header["alg"])
		}
		return p.PubKey, nil
	})
	if err != nil {
		return nil, err
	}
	return token.Claims, nil
}
