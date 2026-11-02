// Package updproto: updater <-> supervisor 메시지(봉투·전송은 protocol·transport 재사용). worker 메시지와 섞지 않는다 — REF-update-p2.md
package updproto

import "nexus/internal/protocol"

// WS·파일 받기 경로(브라우저 API가 아니라 /api 밖 — worker 경로와 같은 이유)
const (
	WSPath   = "/updater/ws"
	FilePath = "/updater/file"
)

// 파일 받기 인증 머리(HTTP)
const (
	HeaderMain = "X-Nexus-Main"
	HeaderKey  = "X-Nexus-Key"
)

const (
	MsgHello  protocol.MsgType = "UPD_HELLO"   // REQ u->s  Hello -> HelloResponse
	MsgSetDir protocol.MsgType = "UPD_SET_DIR" // REQ s->u  SetDir -> 빈 응답(실행 중이면 오류)
	MsgRun    protocol.MsgType = "UPD_RUN"     // REQ s->u  Run -> 빈 응답(받음). 받기·실행은 뒤에서, 끝 = MsgEnd
	MsgInput  protocol.MsgType = "UPD_INPUT"   // EVENT s->u Input(interactive 단계만)
	MsgResize protocol.MsgType = "UPD_RESIZE"  // EVENT s->u Resize
	MsgCancel protocol.MsgType = "UPD_CANCEL"  // EVENT s->u Cancel
	MsgOutput protocol.MsgType = "UPD_OUTPUT"  // EVENT u->s Output
	MsgEnd    protocol.MsgType = "UPD_END"     // REQ u->s  End -> 빈 응답(받을 때까지 재접속 뒤에도 다시 보냄)
)

// Hello: SubKey 빈 값 = 첫 접속(supervisor가 발급). Running = 아직 끝나지 않은 실행(재접속)
type Hello struct {
	MainKey   string  `json:"mainKey"`
	SubKey    string  `json:"subKey,omitempty"`
	APIKey    string  `json:"apiKey"`
	OS        string  `json:"os"`
	Arch      string  `json:"arch"`
	TargetDir string  `json:"targetDir"`
	Installed string  `json:"installed"` // 마지막으로 끝까지 성공한 배포 버전(P3부터 채움)
	Version   string  `json:"version"`   // updater 빌드 태그
	Running   []int64 `json:"running,omitempty"`
}

type HelloResponse struct {
	SubKey string `json:"subKey"`
}

type SetDir struct {
	TargetDir string `json:"targetDir"`
}

// StageFile: 합친 결과 한 파일. URL = FilePath 아래 상대 주소(쿼리 포함)
type StageFile struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
	Sha  string `json:"sha"`
	Mode uint32 `json:"mode"`
	URL  string `json:"url"`
}

// Step 종류
const (
	KindScript  = "script"
	KindBuiltin = "builtin"
)

type Step struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Script      string `json:"script,omitempty"`
	Builtin     string `json:"builtin,omitempty"`
	TimeoutSec  int64  `json:"timeoutSec,omitempty"` // 0 = updater 기본값
	Interactive bool   `json:"interactive,omitempty"`
}

// Run: 단계 하나. Stage = 합친 파일 목록의 sha(같으면 받은 staging을 다시 씀). DeployID 빈 값 = 개별 실행
type Run struct {
	RunID    int64       `json:"runId"`
	DeployID string      `json:"deployId,omitempty"`
	Version  string      `json:"version"`
	Stage    string      `json:"stage"`
	Files    []StageFile `json:"files"`
	Step     Step        `json:"step"`
	Cols     uint16      `json:"cols,omitempty"`
	Rows     uint16      `json:"rows,omitempty"`
}

type Input struct {
	RunID int64  `json:"runId"`
	Data  []byte `json:"data"`
}

type Resize struct {
	RunID int64  `json:"runId"`
	Cols  uint16 `json:"cols"`
	Rows  uint16 `json:"rows"`
}

type Cancel struct {
	RunID int64 `json:"runId"`
}

type Output struct {
	RunID int64  `json:"runId"`
	Data  []byte `json:"data"`
}

// 결과(update_step_runs.result와 같은 값)
const (
	ResultOK      = "ok"
	ResultFail    = "fail"
	ResultTimeout = "timeout"
	ResultCancel  = "cancel"
)

type End struct {
	RunID    int64  `json:"runId"`
	Result   string `json:"result"`
	ExitCode *int   `json:"exitCode,omitempty"` // script만
	Msg      string `json:"msg,omitempty"`
}
