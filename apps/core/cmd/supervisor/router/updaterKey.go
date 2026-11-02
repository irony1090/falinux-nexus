package router

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"time"

	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/store"
	"nexus/internal/web"

	"github.com/labstack/echo/v4"
)

// updater 접속 키 — 정-1(접속 인증), REF-update-p2.md. 저장·다시 보기는 worker 장비 키(deviceKey.go)와 같은 함수, 접두어만 달라 두 키를 섞어 넣으면 바로 보인다
const updaterKeyPrefix = "nxu_"

func newUpdaterKey() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return updaterKeyPrefix + base64.RawURLEncoding.EncodeToString(b)
}

type updKeyResponse struct {
	MainKey string `json:"mainKey"`
	APIKey  string `json:"apiKey"`
}

func (r *supervisorRouter) updShowKey(c echo.Context) error {
	r.requireAdmin(c)
	d := requireUpdDevice(c, updName(c.Param("owner"), "메인키"))
	if len(d.ApiKeyEnc) == 0 {
		panic(web.Err(404, "키가 아직 없습니다. 발급하세요"))
	}
	key, err := openAPIKey(d.ApiKeyEnc)
	if err != nil {
		panic(web.Err(500, "키를 풀 수 없습니다(KEY_SECRET이 바뀌었으면 재발급하세요)"))
	}
	return c.JSON(200, updKeyResponse{MainKey: d.MainKey, APIKey: key})
}

// updReissueKey: 발급·재발급. 이미 붙은 updater는 그대로, 다음 접속부터 새 키
func (r *supervisorRouter) updReissueKey(c echo.Context) error {
	r.requireAdmin(c)
	main := updName(c.Param("owner"), "메인키")
	key := newUpdaterKey()
	n, err := TxQueries(c).UpdateUpdateDeviceKey(c.Request().Context(), superdb.UpdateUpdateDeviceKeyParams{MainKey: main, ApiKeyHash: apiKeyHash(key), ApiKeyEnc: sealAPIKey(key)})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	if n == 0 {
		panic(web.Err(404, "장비 %s가 없습니다", main))
	}
	return c.JSON(200, updKeyResponse{MainKey: main, APIKey: key})
}

var errUpdaterAuth = errors.New("등록되지 않은 메인키이거나 키가 다릅니다")

// checkUpdaterKey: 장비 루트가 있고 키가 맞아야 — 접속(HELLO)·파일 받기 공용
func checkUpdaterKey(mainKey, key string) error {
	if mainKey == "" || key == "" {
		return errUpdaterAuth
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	hash, err := store.GetStorePool().Queries().GetUpdateDeviceAuth(ctx, mainKey)
	if err != nil || hash == "" {
		return errUpdaterAuth
	}
	if subtle.ConstantTimeCompare([]byte(hash), []byte(apiKeyHash(key))) != 1 {
		return errUpdaterAuth
	}
	return nil
}
