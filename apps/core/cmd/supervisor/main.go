// Command supervisor는 다수의 worker agent를 관리하는 총괄 서버 진입점이다.
package main

import "log"

func main() {
	log.Println("nexus supervisor: starting (scaffold)")
	// TODO: PostgreSQL 연결 (pgx/v5) + goose 마이그레이션
	// TODO: transport 서버 기동 (worker WebSocket 수신)
	// TODO: registry / bind / subscribe 매니저 초기화
}
