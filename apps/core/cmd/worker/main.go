// Command worker는 장치에 배포되어 자기 자신의 process를 관리하는 에이전트 진입점이다.
package main

import "log"

func main() {
	log.Println("nexus worker: starting (scaffold)")
	// TODO: SQLite 연결 (modernc.org/sqlite, WAL) + goose 마이그레이션
	// TODO: supervisor에 WebSocket 연결 (transport)
	// TODO: process manager 초기화 (휘발성, 메모리 관리)
}
