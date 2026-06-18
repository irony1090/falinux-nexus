package store

import (
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
)

// Migrate는 주어진 마이그레이션 FS를 SQLite DB에 적용한다(goose.Up).
// InitStorePool과 분리된 별도 단계 — 연결 후 호출자가 명시적으로 부른다.
// 멱등: 매 실행마다 호출해도 미적용분만 반영하고 데이터는 보존한다.
//
//	store.InitStorePool(path, pragmas)
//	store.GetStorePool().Migrate(migrations.FS, ".")
func (s *StorePool) Migrate(fsys fs.FS, dir string) error {
	if s.err != nil {
		return s.err
	}
	goose.SetBaseFS(fsys)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}
	if err := goose.Up(s.db, dir); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}
