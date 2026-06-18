package hunyuan

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const tc3Service = "hunyuan"

// tc3Signer signs Tencent Cloud API requests with TC3-HMAC-SHA256.
type tc3Signer struct {
	secretID  string
	secretKey string
	region    string
	service   string
}

func newTC3Signer(secretID, secretKey, region string) *tc3Signer {
	if region == "" {
		region = "ap-guangzhou"
	}
	return &tc3Signer{
		secretID:  secretID,
		secretKey: secretKey,
		region:    region,
		service:   tc3Service,
	}
}

func (s *tc3Signer) sign(req *http.Request, payload []byte, timestamp int64) {
	host := req.URL.Host
	action := req.Header.Get("X-TC-Action")
	if action == "" {
		action = "ChatCompletions"
	}
	version := req.Header.Get("X-TC-Version")
	if version == "" {
		version = "2023-09-01"
	}

	date := time.Unix(timestamp, 0).UTC().Format("2006-01-02")
	credentialScope := fmt.Sprintf("%s/%s/tc3_request", date, s.service)

	hashedPayload := sha256Hex(payload)
	canonicalHeaders := fmt.Sprintf("content-type:%s\nhost:%s\nx-tc-action:%s\n",
		strings.ToLower(req.Header.Get("Content-Type")),
		strings.ToLower(host),
		strings.ToLower(action),
	)
	signedHeaders := "content-type;host;x-tc-action"
	canonicalRequest := strings.Join([]string{
		req.Method,
		req.URL.Path,
		req.URL.RawQuery,
		canonicalHeaders,
		signedHeaders,
		hashedPayload,
	}, "\n")

	stringToSign := strings.Join([]string{
		"TC3-HMAC-SHA256",
		fmt.Sprintf("%d", timestamp),
		credentialScope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")

	secretDate := hmacSHA256([]byte("TC3"+s.secretKey), date)
	secretService := hmacSHA256(secretDate, s.service)
	secretSigning := hmacSHA256(secretService, "tc3_request")
	signature := hex.EncodeToString(hmacSHA256(secretSigning, stringToSign))

	req.Header.Set("Authorization", fmt.Sprintf(
		"TC3-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		s.secretID, credentialScope, signedHeaders, signature,
	))
	req.Header.Set("Host", host)
	req.Header.Set("X-TC-Timestamp", fmt.Sprintf("%d", timestamp))
	req.Header.Set("X-TC-Version", version)
	req.Header.Set("X-TC-Region", s.region)
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func hmacSHA256(key []byte, msg string) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(msg))
	return mac.Sum(nil)
}
