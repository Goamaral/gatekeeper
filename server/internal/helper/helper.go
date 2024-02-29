package helper

import (
	"crypto/rand"
	"encoding/base64"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/uuid"
)

func RelativePath(relativePath string) string {
	_, file, _, _ := runtime.Caller(1)
	folderPath := filepath.Dir(file)
	return folderPath + "/" + relativePath
}

const ApiKeySuffixLength = 16

func GenerateApiKey(companyUuid uuid.UUID) (string, error) {
	b := make([]byte, ApiKeySuffixLength)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	compactAccoutnUuid := strings.ReplaceAll(companyUuid.String(), "-", "")
	return compactAccoutnUuid + strings.ToLower(base64.URLEncoding.EncodeToString(b)), nil
}
