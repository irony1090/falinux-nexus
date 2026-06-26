package router

import (
	"context"
	"fmt"
	"log"

	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/store"
	"nexus/internal/web"

	"github.com/labstack/echo/v4"
)

// 구 PortBridge web/requestContext.go(전역 map[*http.Request]any + RequestScope 인터페이스 +
// 도메인별 *Context 4종)를 대체한다. 그 기계장치의 실체는 "요청 경계 자동 커밋/롤백"이었고,
// echo.Context가 요청 수명 저장소(c.Set/c.Get)를 이미 주므로 전역 map은 불필요해진다.
// 4개 중복 Context는 아래 단일 txScope + 미들웨어 하나로 합쳐진다.

const txContextKey = "txScope"

// txScope는 요청 1건의 트랜잭션 수명을 담는다.
// 요청마다 새로 만들어 echo.Context에 실리므로 단일 goroutine 전용 → 락 불필요.
type txScope struct {
	pool *store.StorePool
	tx   *store.Transaction
	err  error
}

// ensureTx는 트랜잭션을 lazy하게 시작한다(처음 호출 때만 Begin).
// 한 번도 부르지 않은 읽기전용 핸들러는 트랜잭션을 열지 않는다(구 *Context.Transaction 동작 보존).
// Begin 실패는 panic(web.Err(500))으로 던져 PanicMiddleware가 응답한다.
func (s *txScope) ensureTx() *store.Transaction {
	if s.tx == nil {
		t := s.pool.Transaction()
		if err := t.Begin(context.Background()); err != nil {
			panic(web.Err(500, "트랜잭션 시작 실패: %v", err))
		}
		s.tx = t
	}
	return s.tx
}

// release는 요청 종료 시 한 번 호출된다. err 유무로 커밋/롤백을 가른다.
// 트랜잭션을 한 번도 안 열었으면(tx==nil) 아무것도 하지 않는다.
func (s *txScope) release() {
	if s.tx == nil {
		return
	}
	if s.err != nil {
		if rbErr := s.tx.Rollback(context.Background(), s.err); rbErr != nil {
			log.Printf("[tx] rollback 실패: %v (원인: %v)", rbErr, s.err)
		}
		return
	}
	if cErr := s.tx.Commit(context.Background()); cErr != nil {
		log.Printf("[tx] commit 실패: %v", cErr)
	}
}

// txMiddleware는 요청마다 txScope를 만들어 echo.Context에 싣고, 종료 시 커밋/롤백한다.
//   - 핸들러 panic(web.Err(...) 등 기대된 에러 + 예기치 못한 버그): err 기록 → 롤백 → 재-panic(바깥 PanicMiddleware가 잡아 응답)
//   - 핸들러 error 반환(드묾): err 기록 → 롤백 → 그대로 반환
//   - 정상: 커밋
//
// PanicMiddleware보다 안쪽(나중 Use)에 등록해야 재-panic이 그쪽으로 전파된다.
func txMiddleware(pool *store.StorePool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			scope := &txScope{pool: pool}
			c.Set(txContextKey, scope)

			defer func() {
				if r := recover(); r != nil {
					scope.err = fmt.Errorf("panic: %v", r)
					scope.release()
					panic(r) // PanicMiddleware로 전파
				}
				scope.err = err // 핸들러가 error를 반환해도 롤백
				scope.release()
			}()

			return next(c)
		}
	}
}

// Tx는 핸들러에서 현재 요청의 트랜잭션을 꺼낸다(lazy Begin).
// 구 web.GetRequestScopePanic[model.XxxContext](req).Transaction() 자리를 대체.
func Tx(c echo.Context) *store.Transaction {
	return c.Get(txContextKey).(*txScope).ensureTx()
}

// TxQueries는 현재 요청 트랜잭션의 sqlc Queries를 바로 꺼내는 단축 헬퍼.
func TxQueries(c echo.Context) *superdb.Queries {
	return Tx(c).QueriesPanic()
}
