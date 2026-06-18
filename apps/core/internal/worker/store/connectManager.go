// Package store는 worker(SQLite) DB 연결 풀과 트랜잭션을 관리한다.
//
// 구 PortBridge agentStore.connectManager를 이식하면서 두 가지를 바꿨다:
//   - PRAGMA를 하드코딩 대신 옵션(map)으로 받아 DSN(_pragma=)에 실음
//     → foreign_keys 등 "커넥션별" 설정이 풀의 모든 커넥션에 적용되도록(Exec는 1개에만 먹음)
//   - sqlc 생성 패키지(workerdb)를 참조하도록 New/Queries 타입 교체
package store

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"sync"

	workerdb "nexus/internal/worker/db/gen"

	_ "modernc.org/sqlite"
)

type StorePool struct {
	db  *sql.DB
	err error
}

var (
	storePool *StorePool
	mutex     sync.Mutex
)

// InitStorePool은 SQLite 파일을 열고 PRAGMA를 적용한다.
// pragmas는 DSN 쿼리파라미터(_pragma=)로 실려 매 커넥션에 적용된다.
// 예: InitStorePool("worker.db", map[string]string{"journal_mode":"WAL","foreign_keys":"1"})
func InitStorePool(dbFile string, pragmas map[string]string) error {
	mutex.Lock()
	defer mutex.Unlock()

	if storePool != nil {
		if storePool.err == nil {
			return fmt.Errorf("이미 연결되어 있습니다")
		}
		// 이전 연결이 실패했으면 정리하고 재시도
		if storePool.db != nil {
			storePool.db.Close()
		}
		storePool = nil
	}

	db, err := sql.Open("sqlite", buildDSN(dbFile, pragmas))
	if err == nil {
		err = db.Ping()
	}

	storePool = &StorePool{db: db, err: err}
	return err
}

// buildDSN은 파일 경로에 PRAGMA들을 _pragma= 쿼리로 붙인다.
// 예: worker.db?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)
func buildDSN(dbFile string, pragmas map[string]string) string {
	if len(pragmas) == 0 {
		return dbFile
	}
	keys := make([]string, 0, len(pragmas))
	for k := range pragmas {
		keys = append(keys, k)
	}
	sort.Strings(keys) // DSN을 결정적으로(로그/비교 안정)
	ps := make([]string, 0, len(keys))
	for _, k := range keys {
		ps = append(ps, fmt.Sprintf("_pragma=%s(%s)", k, pragmas[k]))
	}
	return dbFile + "?" + strings.Join(ps, "&")
}

func GetStorePool() *StorePool {
	mutex.Lock()
	defer mutex.Unlock()
	return storePool
}

func (s *StorePool) Close() {
	mutex.Lock()
	defer mutex.Unlock()
	if s.db != nil {
		s.db.Close()
	}
	storePool = nil
}

func (s *StorePool) Queries() *workerdb.Queries {
	return workerdb.New(s.db)
}

func (s *StorePool) Transaction() *Transaction {
	return &Transaction{
		store:     s,
		listeners: make([]func(error), 0),
	}
}

type Transaction struct {
	store     *StorePool
	tx        *sql.Tx
	q         *workerdb.Queries
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
// (구 코드는 인자를 무시하고 store.err을 넘기던 버그 → 결과 err을 그대로 전달하도록 수정)
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

	tx, err := t.store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	t.setTxUnsafe(tx)
	return nil
}

func (t *Transaction) setTxUnsafe(tx *sql.Tx) {
	t.tx = tx
	if tx == nil {
		t.q = nil
		t.start = false
	} else {
		t.q = workerdb.New(tx)
		t.start = true
	}
}

func (t *Transaction) Queries() (*workerdb.Queries, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.store.err != nil {
		return nil, t.store.err
	} else if !t.isStartUnsafe() {
		return nil, fmt.Errorf("트랜잭션이 시작되지 않았습니다")
	}

	return t.q, nil
}

func (t *Transaction) QueriesPanic() *workerdb.Queries {
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

	return t.tx.Commit()
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

	return t.tx.Rollback()
}
