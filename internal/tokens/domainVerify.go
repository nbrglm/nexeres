package tokens

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
)

func GenerateDomainVerifyCode(domain, orgId, secretKey string) string {
	verificationCode := sha256.Sum256([]byte(fmt.Sprintf("%s.%s.%s", secretKey, orgId, domain)))

	hash := base64.URLEncoding.EncodeToString(verificationCode[:])
	return hash
}

func ValidateDomainVerifyCode(domain, orgId, secretKey, code string) bool {
	// Implementation of the function to validate a domain verification code
	// This is a placeholder implementation and should be replaced with actual logic
	expectedCode := GenerateDomainVerifyCode(domain, orgId, secretKey)
	return subtle.ConstantTimeCompare([]byte(expectedCode), []byte(code)) == 1
}
