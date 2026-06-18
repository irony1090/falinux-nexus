// Package migrations는 worker(SQLite) goose 마이그레이션 파일을 임베드한다.
// 배포된 worker 바이너리가 단독으로(원본 .sql 없이) 마이그레이션을 적용할 수 있게 한다.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
