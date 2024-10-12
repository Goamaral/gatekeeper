package helper

import (
	"crypto/rand"
	"encoding/base64"
	"path/filepath"
	"runtime"
	"strings"

	"braces.dev/errtrace"
	"github.com/google/uuid"
)

func RelativePath(relativePath string) string {
	_, file, _, _ := runtime.Caller(1)
	folderPath := filepath.Dir(file)
	return folderPath + "/" + relativePath
}

const ApiKeySuffixLength = 16

func GenerateApiKey() (string, error) {
	b := make([]byte, ApiKeySuffixLength)
	_, err := rand.Read(b)
	if err != nil {
		return "", errtrace.Wrap(err)
	}
	uuid, err := uuid.NewV7()
	if err != nil {
		return "", errtrace.Wrap(err)
	}
	return strings.ReplaceAll(uuid.String(), "-", "") + strings.ToLower(base64.URLEncoding.EncodeToString(b)), nil
}
