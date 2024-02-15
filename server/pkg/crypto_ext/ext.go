package crypto_ext

import (
	"crypto/ecdsa"
	"strconv"

	"github.com/ethereum/go-ethereum/crypto"
)

// https://eips.ethereum.org/EIPS/eip-191
func PersonalSign(data []byte, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	hash := crypto.Keccak256Hash([]byte("\x19Ethereum Signed Message:\n"+strconv.Itoa(len(data))), data).Bytes()
	return crypto.Sign(hash, privateKey)
}
