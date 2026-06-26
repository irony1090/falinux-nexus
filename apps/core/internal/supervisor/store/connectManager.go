// Package store는 supervisor(PostgreSQL) DB 연결 풀과 트랜잭션을 관리한다.
//
// 구 PortBridge internal/store.connectManager(pgx)를 이식하면서 worker(SQLite)판과
// 동일한 수정을 반영했다:
//   - afterUnsfae 오타 → afterUnsafe + 커밋/롤백 결과 err을 리스너에 그대로 전달
//     (구 코드는 인자를 무시하고 store.err을 넘기던 버그)
//   - InitStorePool 재호출 시 실패한 이전 연결을 정리하고 재시도(모호한 "return nil" 제거)
//   - sqlc 생성 패키지(superdb)를 참조하도록 New/Queries 타입 교체
package store

import (
	"context"
	"fmt"
	"sync"

	superdb "nexus/internal/supervisor/db/gen"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StorePool struct {
	pool *pgxpool.Pool
	dsn  string // 마이그레이션용 throwaway *sql.DB를 여는 데 재사용
	err  error
}

var (
	storePool *StorePool
	mutex     sync.Mutex
)

// InitStorePool은 PostgreSQL 풀을 열고 Ping으로 연결을 확인한다.
func InitStorePool(user, pass, host string, port int, name string) error {
	mutex.Lock()
	defer mutex.Unlock()

	if storePool != nil {
		if storePool.err == nil {
			return fmt.Errorf("이미 연결되어 있습니다")
		}
		// 이전 연결이 실패했으면 정리하고 재시도
		if storePool.pool != nil {
			storePool.pool.Close()
		}
		storePool = nil
	}

	dbInfo := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", user, pass, host, port, name)

	pool, err := pgxpool.New(context.Background(), dbInfo)
	if err == nil {
		err = pool.Ping(context.Background())
	}

	storePool = &StorePool{pool: pool, dsn: dbInfo, err: err}
	return err
}

func GetStorePool() *StorePool {
	mutex.Lock()
	defer mutex.Unlock()
	return storePool
}

func (s *StorePool) Close() {
	mutex.Lock()
	defer mutex.Unlock()
	if s.pool != nil {
		s.pool.Close()
	}
	storePool = nil
}

func (s *StorePool) Queries() *superdb.Queries {
	return superdb.New(s.pool)
}

func (s *StorePool) Transaction() *Transaction {
	return &Transaction{
		store:     s,
		listeners: make([]func(error), 0),
	}
}

type Transaction struct {
	store     *StorePool
	tx        pgx.Tx
	q         *superdb.Queries
	start     bool
	mutex     sync.Mutex
	listeners []func(error)
}

func (t *Transaction) AddAfterRelease(f func(error)) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.listeners = append(t.listeners, f)
}

// afterUnsafe는 커밋/롤백 결과(err)를 등록된 리스너들에게 알린다.
func (t *Transaction) afterUnsafe(err error) {
	for _, f := range t.listeners {
		f(err)
	}
}

func (t *Transaction) isStartUnsafe() bool {
	return t.start
}

func (t *Transaction) IsStart() bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.isStartUnsafe()
}

func (t *Transaction) Begin(ctx context.Context) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.isStartUnsafe() {
		return fmt.Errorf("이미 트랜잭션 중입니다")
	} else if t.store.err != nil {
		return t.store.err
	}

	tx, err := t.store.pool.Begin(ctx)
	if err != nil {
		return err
	}

	t.setTxUnsafe(tx)
	return nil
}

func (t *Transaction) setTxUnsafe(tx pgx.Tx) {
	t.tx = tx
	if tx == nil {
		t.q = nil
		t.start = false
	} else {
		t.q = superdb.New(tx)
		t.start = true
	}
}

func (t *Transaction) Queries() (*superdb.Queries, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.store.err != nil {
		return nil, t.store.err
	} else if !t.isStartUnsafe() {
		return nil, fmt.Errorf("트랜잭션이 시작되지 않았습니다")
	}

	return t.q, nil
}

func (t *Transaction) QueriesPanic() *superdb.Queries {
	q, err := t.Queries()
	if err != nil {
		panic(err)
	}
	return q
}

func (t *Transaction) Commit(ctx context.Context) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.isStartUnsafe() {
		return fmt.Errorf("트랜잭션이 시작되지 않았습니다")
	} else if t.store.err != nil {
		return t.store.err
	}

	defer t.setTxUnsafe(nil)
	defer func() {
		go t.afterUnsafe(nil)
	}()

	return t.tx.Commit(ctx)
}

func (t *Transaction) Rollback(ctx context.Context, err error) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.isStartUnsafe() {
		return nil
	} else if t.store.err != nil {
		return t.store.err
	}

	defer t.setTxUnsafe(nil)
	defer func() {
		go t.afterUnsafe(err)
	}()

	return t.tx.Rollback(ctx)
}
