package process

import (
	"fmt"

	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
)

// WorkerNodePath는 worker 위 실행/저장 경로를 조립한다(치환 토큰 포함).
// supervisor는 worker의 실제 경로를 모르므로 {WORKER_BASE}를 박아 보내고, worker가
// 실행/저장 직전 자기 ProcessRoot로 치환한다(cmd/worker 핸들러). 경로 조립 책임은
// supervisor(여기), 지역화(placeholder 치환) 책임은 worker로 분리한다.
//
// 형식: {WORKER_BASE}/<node.ID>/<proc.Uid>
//   - node.ID  : 어떤 스크립트인지(같은 노드 재실행 시 같은 폴더).
//   - proc.Uid : 실행 인스턴스별 격리(동시 실행·재실행 충돌 방지).
//
// 이 문자열은 content 선배치 전송 DestPath와 실행 명세(EXEC Cmd / EDIT Args)가 함께
// 쓴다 — 반드시 같은 파일을 가리켜야 하므로 단일 출처로 둔다(불일치=버그).
func WorkerNodePath(node superdb.Node, proc superdb.Process) string {
	return fmt.Sprintf("%s/%d/%s", protocol.PlaceholderWorkerBase, node.ID, proc.Uid)
}
