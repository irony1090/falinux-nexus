// Package store는 worker(SQLite) DB 연결 풀을 연다.
// PRAGMA는 DSN(_pragma=)에 싣는다 — foreign_keys 같은 커넥션별 설정이 풀의 모든 커넥션에 적용되게(Exec는 1개에만 먹음)
package store

import (
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

// InitStorePool: SQLite 파일을 열고 PRAGMA 적용(예: {"journal_mode":"WAL","foreign_keys":"1"})
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

// buildDSN: worker.db?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)
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

func (s *StorePool) Queries() *workerdb.Queries {
	return workerdb.New(s.db)
}
