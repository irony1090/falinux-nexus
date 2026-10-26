package router

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"regexp"
	"strings"

	"nexus/cmd/supervisor/constants"
)

// 장비 API 키 = "nxk_" + base64url(32바이트). 붙여넣기(회원-Q14)도 이 형식만 받는다 → REF-member.md
const apiKeyPrefix = "nxk_"

// 장비 이름(메인키) — URL 경로·InstanceKey('#' 구분)에 그대로 쓰이므로 이 문자만
var mainKeyPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

func newAPIKey() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return apiKeyPrefix + base64.RawURLEncoding.EncodeToString(b)
}

func validAPIKey(k string) bool {
	raw, ok := strings.CutPrefix(k, apiKeyPrefix)
	if !ok {
		return false
	}
	b, err := base64.RawURLEncoding.DecodeString(raw)
	return err == nil && len(b) == 32
}

// 접속 검증용 — 키가 무작위 32바이트라 솔트·느린 해시가 필요 없다
func apiKeyHash(k string) string {
	return toHash(k)
}

func keyCipher() cipher.AEAD {
	sum := sha256.Sum256([]byte(constants.GetEnv().KeySecret))
	block, err := aes.NewCipher(sum[:])
	if err != nil {
		panic(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}
	return gcm
}

// sealAPIKey: nonce + 암호문 (다시 보기용, 회원-Q16)
func sealAPIKey(k string) []byte {
	gcm := keyCipher()
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		panic(err)
	}
	return gcm.Seal(nonce, nonce, []byte(k), nil)
}

// openAPIKey: KEY_SECRET이 바뀌었으면 실패한다
func openAPIKey(enc []byte) (string, error) {
	gcm := keyCipher()
	if len(enc) < gcm.NonceSize() {
		return "", errors.New("암호문이 짧습니다")
	}
	b, err := gcm.Open(nil, enc[:gcm.NonceSize()], enc[gcm.NonceSize():], nil)
	return string(b), err
}
