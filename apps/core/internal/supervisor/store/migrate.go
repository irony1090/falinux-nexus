package store

import (
	"database/sql"
	"fmt"
	"io/fs"

	_ "github.com/jackc/pgx/v5/stdlib" // database/sql용 "pgx" 드라이버 등록
	"github.com/pressly/goose/v3"
)

// Migrate는 주어진 마이그레이션 FS를 PostgreSQL DB에 적용한다(goose.Up).
// InitStorePool과 분리된 별도 단계 — 연결 후 호출자가 명시적으로 부른다.
// 멱등: 매 실행마다 호출해도 미적용분만 반영하고 데이터는 보존한다.
//
// goose는 database/sql(*sql.DB)을 요구하는데 풀은 pgxpool이라,
// 같은 DSN으로 마이그레이션 1회용 *sql.DB를 열고 끝나면 닫는다(풀과 무관).
//
//	store.InitStorePool(user, pass, host, name, port)
//	store.GetStorePool().Migrate(migrations.FS, ".")
func (s *StorePool) Migrate(fsys fs.FS, dir string) error {
	if s.err != nil {
		return s.err
	}

	db, err := sql.Open("pgx", s.dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	goose.SetBaseFS(fsys)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	if err := goose.Up(db, dir); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}
