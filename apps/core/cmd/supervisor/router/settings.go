package router

import (
	"encoding/json"
	"net/http"
	"regexp"

	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/web"

	"github.com/labstack/echo/v4"
)

// 계정 설정 — 설정 두 겹 중 계정 겹. 서버는 키 이름을 모르고 값만 보관(모양 검사 = 프론트 SPECS) — REF-settings-footer.md "설정 두 겹"

const (
	settingMaxKeys  = 64   // 계정마다 키 수 — 프론트가 아는 키보다 넉넉히, 아무 키나 쌓이지 않게
	settingMaxBytes = 4096 // 값 하나의 JSON 크기
)

var settingKeyRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]{0,63}$`)

func (r *supervisorRouter) mountSettings(e *echo.Group) {
	g := e.Group("/settings")
	g.GET("", r.listSettings)
	g.PUT("/:key", r.saveSetting)
}

type settingSaveRequest struct {
	Value json.RawMessage `json:"value"`
}

func readSettings(c echo.Context, userID int64) protocol.SettingsUpdateEvent {
	rows, err := TxQueries(c).ListUserSettings(c.Request().Context(), userID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	out := protocol.SettingsUpdateEvent{Settings: make(map[string]json.RawMessage, len(rows))}
	for _, row := range rows {
		out.Settings[row.Key] = row.Value
		out.Rev = max(out.Rev, row.UpdatedAt.Time.UnixMicro())
	}
	return out
}

// GET /settings — 계정 값 전체 {settings: {키: 값}, rev}
func (r *supervisorRouter) listSettings(c echo.Context) error {
	sess := r.requireSession(c)
	return c.JSON(http.StatusOK, readSettings(c, sess.Data.ID))
}

// PUT /settings/:key {value} — 저장 뒤 계정 값 전체를 응답하고 같은 값을 모든 탭에 push
func (r *supervisorRouter) saveSetting(c echo.Context) error {
	sess := r.requireSession(c)
	userID := sess.Data.ID
	key := c.Param("key")
	if !settingKeyRe.MatchString(key) {
		panic(web.Err(400, "잘못된 설정 이름입니다"))
	}
	body := bindBody[settingSaveRequest](c)
	if len(body.Value) == 0 || len(body.Value) > settingMaxBytes || !json.Valid(body.Value) {
		panic(web.Err(400, "잘못된 설정 값입니다"))
	}
	if cur := readSettings(c, userID).Settings; len(cur) >= settingMaxKeys {
		if _, ok := cur[key]; !ok {
			panic(web.Err(400, "설정 항목이 너무 많습니다"))
		}
	}
	if err := TxQueries(c).SaveUserSetting(c.Request().Context(), superdb.SaveUserSettingParams{
		UserID: userID, Key: key, Value: body.Value,
	}); err != nil {
		panic(web.Err(500, "%v", err))
	}

	all := readSettings(c, userID)
	AfterCommit(c, func() {
		r.publishUser(userID, protocol.MsgSettingsUpdate, all)
	})
	return c.JSON(http.StatusOK, all)
}
