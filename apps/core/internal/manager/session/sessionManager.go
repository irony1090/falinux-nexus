package session

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
)

// NameFunc는 세션에서 표시용 이름(또는 식별자)을 도출하는 전략이다.
// NewSessionManager에 주입되어 SessionElement.Name()이 호출한다(nil이면 "" 반환).
type NameFunc[T any] func(data T, session *sessions.Session, key string) string

type SessionElement[T any] struct {
	key     string
	req     *http.Request
	res     *echo.Response
	Session *sessions.Session
	Data    *T
	nameFn  NameFunc[T]
}

func NewSessionElement[T any](req *http.Request, res *echo.Response, key string, session *sessions.Session, nameFn NameFunc[T]) *SessionElement[T] {
	var data T
	v := reflect.ValueOf(&data).Elem()

	// T가 struct일 때만 세션 값으로 필드를 복원한다(기본형/map 등은 NumField 패닉 방지).
	if v.Kind() == reflect.Struct {
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if !isGobSerializable(field.Type.Kind()) {
				continue
			}

			if val, ok := session.Values[field.Name]; ok {
				fv := v.Field(i)
				if fv.CanSet() && val != nil && reflect.TypeOf(val).AssignableTo(fv.Type()) {
					fv.Set(reflect.ValueOf(val))
				}
			}
		}
	}

	return &SessionElement[T]{
		key:     key,
		req:     req,
		res:     res,
		Session: session,
		Data:    &data,
		nameFn:  nameFn,
	}
}

func (se *SessionElement[T]) saveData() {
	v := reflect.ValueOf(se.Data).Elem()
	if v.Kind() != reflect.Struct {
		return // struct가 아니면 저장할 필드가 없다(NumField 패닉 방지)
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue // 비공개 필드는 .Interface()가 패닉 → 건너뜀(읽기 쪽 CanSet 가드와 대칭)
		}
		if !isGobSerializable(field.Type.Kind()) {
			continue
		}
		se.Session.Values[field.Name] = v.Field(i).Interface()
	}
}

// Name은 주입된 NameFunc로 표시용 이름/식별자를 도출한다(미주입이면 "").
func (se *SessionElement[T]) Name() string {
	if se.nameFn == nil {
		return ""
	}
	return se.nameFn(*se.Data, se.Session, se.key)
}

func (se *SessionElement[T]) Save() error {
	if se.res == nil {
		return fmt.Errorf("Response 객체가 없습니다")
	}
	se.saveData()
	err := se.Session.Save(se.req, se.res)
	// log.Printf("[SESS] SAVE %v", err)
	return err
}

type SessionManager[T any] struct {
	store  *sessions.CookieStore
	key    string
	nameFn NameFunc[T]
}

func NewSessionManager[T any](keyPaire string, sessionKey string, nameFn NameFunc[T]) *SessionManager[T] {
	// log.Printf("[NEW_SESSION_MANAGER] %s, %s", string(keyPaire), sessionKey)
	store := sessions.NewCookieStore([]byte(keyPaire))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
		Secure:   false,                // HTTP 환경
		SameSite: http.SameSiteLaxMode, // 같은 IP → same-site
	}
	return &SessionManager[T]{
		store:  store,
		key:    sessionKey,
		nameFn: nameFn,
	}
}

func (sm *SessionManager[T]) GetSession(req *http.Request, res *echo.Response) *SessionElement[T] {
	// store.Get은 쿠키 디코드 실패(손상/키 교체) 시에도 err와 함께 '새 세션'을 돌려준다.
	// 그 새 세션은 그대로 쓸 수 있으므로 err로 버리지 않는다(버리면 낡은 쿠키 = 영구 잠금).
	session, _ := sm.store.Get(req, sm.key)
	if session == nil {
		return nil
	}
	return NewSessionElement(req, res, sm.key, session, sm.nameFn)
}

func isGobSerializable(k reflect.Kind) bool {
	switch k {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String:
		return true
	default:
		return false
	}
}
