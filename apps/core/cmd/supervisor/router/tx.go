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

// 요청 경계 트랜잭션: 요청마다 txScope 하나, 미들웨어가 끝에 커밋/롤백

const txContextKey = "txScope"

// txScope: 요청 1건의 트랜잭션 수명. 요청마다 새로 만들어 단일 goroutine 전용(락 없음)
type txScope struct {
	pool  *store.StorePool
	tx    *store.Transaction
	err   error
	hooks []func() // AfterCommit
}

// ensureTx: 처음 부를 때만 Begin(안 부른 읽기 핸들러는 트랜잭션 없음). 실패 = panic(web.Err(500))
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

// release: 요청 끝에 한 번 — err면 롤백, 아니면 커밋 뒤 hooks. 트랜잭션을 안 열었으면 아무것도 안 함
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
		return
	}
	for _, fn := range s.hooks {
		fn()
	}
}

// txMiddleware: 요청마다 txScope를 싣고 끝에 커밋/롤백. panic·error 반환 = 롤백(panic은 다시 던짐)
// 순서 의존: PanicMiddleware보다 안쪽(나중 Use)에 등록해야 다시 던진 panic이 그쪽에 잡힌다
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

// Tx: 현재 요청의 트랜잭션(lazy Begin)
func Tx(c echo.Context) *store.Transaction {
	return c.Get(txContextKey).(*txScope).ensureTx()
}

// TxQueries: 현재 요청 트랜잭션의 sqlc Queries
func TxQueries(c echo.Context) *superdb.Queries {
	return Tx(c).QueriesPanic()
}

// AfterCommit: fn을 커밋 성공 뒤에만 실행한다(롤백·커밋 실패면 안 함). 등록 순서대로, 핸들러 반환 후 동기 실행.
//
// 소켓 발행 규칙 — 롤백된 변경이 나가지 않게:
//   - publish*·reconcileAccess·sendKill은 호출자가 AfterCommit으로 감싼다(트랜잭션 밖 호출은 그대로)
//   - 이름이 ...AfterCommit으로 끝나는 함수만 스스로 감싼다(publishWorkspacesAfterCommit·reconcileShareAfterCommit)
//   - 예외: resizeProcess의 먼저 발행(밀림-가(먼저 발행)) — DB가 아니라 worker에 보낼 값을 미리 알린다
func AfterCommit(c echo.Context, fn func()) {
	scope := c.Get(txContextKey).(*txScope)
	scope.hooks = append(scope.hooks, fn)
}
