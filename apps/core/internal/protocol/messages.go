package protocol

import (
	"fmt"

	"nexus/internal/execute"
)

// 이 파일은 Nexus 도메인 메시지(어휘)를 모은다.
// 봉투(Frame/Kind)는 protocol.go에 있고, 여기엔 "무엇을 주고받는지"만 정의한다.

// MsgRegister는 worker가 supervisor에 접속하며 보내는 등록 요청이다.
// MsgFile*는 파일 전송(supervisor → worker)의 단계별 메시지다.
const (
	MsgRegister   MsgType = "REGISTER"
	MsgFileInit   MsgType = "FILE_INIT"   // 통보 → FileInitResponse
	MsgFileChunk  MsgType = "FILE_CHUNK"  // 데이터 조각 → FileChunkResponse
	MsgFileResult MsgType = "FILE_RESULT" // 전송 완료 통보 → FileResult(검증 결과)
	MsgFileAbort  MsgType = "FILE_ABORT"  // 진행중 중단
)

// RegisterRequest: worker → supervisor.
// 최초 접속이면 SubKey가 빈 값, 재접속(재부팅 후)이면 이전에 받은 SubKey를 함께 보낸다.
type RegisterRequest struct {
	Key    string `json:"key"`              // 사전 지정 고유키 (여러 worker가 같을 수 있음)
	SubKey string `json:"subKey,omitempty"` // 재접속 시에만 채워짐
}

func (r RegisterRequest) InstanceKey() string {
	if r.SubKey == "" {
		return r.Key
	} else {
		return fmt.Sprintf("%s#%s", r.Key, r.SubKey)
	}
}

// RegisterResponse: supervisor → worker.
// 최초면 supervisor가 새로 부여한 SubKey, 재접속이면 확인된 SubKey.
type RegisterResponse struct {
	SubKey string `json:"subKey"`
}

func NewRegisterRequest(key, subKey string) RegisterRequest {
	return RegisterRequest{Key: key, SubKey: subKey}
}

// ===== 파일 전송 (supervisor → worker) =====
//
// 흐름: FileInit(통보) → [FileChunk × N] → FileResult(완료/검증). 중단은 FileAbort.
// 모든 메시지는 TransferID로 한 전송 세션을 묶는다 — Frame.ID는 요청1↔응답1 짝짓기용이라 별개다.
// (한 연결에서 여러 파일이 동시에 오갈 수 있으므로 세션 식별자가 따로 필요)

// FileInitRequest: 전송 통보. (REQ → FileInitResponse)
type FileInitRequest struct {
	TransferID string `json:"transferId"`       // 전송마다 새로 발급된 유일 토큰
	Name       string `json:"name"`             // 원본 파일명(표시/로그용)
	DestPath   string `json:"destPath"`         // 수신측 저장 상대경로(루트 기준 — traversal 검증 대상)
	Size       int64  `json:"size"`             // 전체 바이트 수
	Mode       uint32 `json:"mode"`             // 권한 비트(수신측에서 os.FileMode로 변환)
	Hash       string `json:"hash,omitempty"`   // 전체 무결성 해시(sha256 hex 등)
	Resume     bool   `json:"resume,omitempty"` // 송신자의 이어받기 의향
}

// FileInitResponse: 수락 여부 + 이어받기 시작 위치.
// ResumeOffset은 "받는 쪽"이 정한다(이미 가진 바이트 수). 0이면 처음부터.
type FileInitResponse struct {
	Accept       bool   `json:"accept"`
	ResumeOffset int64  `json:"resumeOffset"`
	Reason       string `json:"reason,omitempty"` // 거부 사유(accept=false일 때)
}

// FileChunkRequest: 데이터 한 조각. (REQ → FileChunkResponse)
// Data는 Frame.Data(json) 봉투라 base64로 인코딩되어 나간다(~33% 증가).
type FileChunkRequest struct {
	TransferID string `json:"transferId"`
	Offset     int64  `json:"offset"` // 이 조각의 시작 바이트 위치(멱등 WriteAt용)
	Data       []byte `json:"data"`
}

// FileChunkResponse: chunk 수신 확인.
type FileChunkResponse struct {
	Received int64 `json:"received"` // 수신측이 도달한 바이트 수(진행/검증용)
}

// FileResultRequest: 송신 완료 통보 → 수신측이 검증(크기·hash) 후 결과를 돌려준다. (REQ → FileResult)
type FileResultRequest struct {
	TransferID string `json:"transferId"`
}

// FileResult: 최종 결과(완료/실패).
type FileResult struct {
	OK        bool   `json:"ok"`
	Reason    string `json:"reason,omitempty"`    // 실패 사유
	Resumable bool   `json:"resumable,omitempty"` // 실패 시 partial 보존 → 이어받기 가능
}

// FileAbortRequest: 진행중 전송 중단. (REQ → 빈 응답)
type FileAbortRequest struct {
	TransferID string `json:"transferId"`
	Reason     string `json:"reason,omitempty"`
}

// ===== process 실행 (supervisor ↔ worker) =====
//
// 제어(REQ/RES): MsgExec / MsgResize / MsgKill — 성패 응답 필요.
// 스트림(EVENT): MsgData(입력·출력 양방향) / MsgStatus(worker→sup) — 짝 응답 없음.
// 실행 종류(ProcessSpec.Type): EXEC=일반 인터랙티브 실행 / EDIT=에디터로 편집 후 내용 회수.
//
//	둘 다 같은 PTY 엔진을 쓰며 Type만 다르다.
//	- EXEC: Cmd를 그대로 실행(argv = Cmd + Args).
//	- EDIT: 편집 대상 파일은 미리 깔아 둔다 — 초기 내용은 SendBuffer(메모리 전송)로 worker에
//	        먼저 보내거나 기존 파일을 그대로 쓴다(인라인 Seed 주입은 폐기). EDIT는 그 경로를
//	        에디터로 열고, 종료 시 MsgEditResult(REQ)로 파일 내용을 read-back한다.
//	        에디터 = Cmd(있으면) / 없으면 worker가 $EDITOR(→없으면 OS 기본) 해석.
//	★경로 치환: supervisor는 worker의 실제 경로를 모르므로 Cmd/Args/Cwd/Env에 치환 토큰
//	({WORKER_BASE})을 박아 보내고, worker가 실행 직전 실제 값으로 치환한다(알 수 없는 {..}는
//	그대로 둠). → 토큰 상수는 아래 Placeholder*.
//
// 모든 메시지는 UID로 한 process를 가리킨다. UID는 initiator(보통 supervisor)가 발급하는
// 시스템 전체 식별자고, 실제 OS PID와는 별개다(PID는 worker 로컬, StatusEvent로 보고된다).
const (
	MsgExec   MsgType = "EXEC"   // sup→worker REQ: ProcessSpec → ExecResponse
	MsgData   MsgType = "DATA"   // EVENT 양방향: 입력(sup→worker) / 출력(worker→sup)
	MsgResize MsgType = "RESIZE" // sup→worker REQ: ResizeRequest → 빈 응답
	MsgKill   MsgType = "KILL"   // sup→worker REQ: KillRequest → 빈 응답
	MsgStatus MsgType = "STATUS" // worker→sup EVENT: StatusEvent

	MsgEditResult MsgType = "EDIT_RESULT" // worker→sup REQ: EditResult → 빈 응답 (EDIT 전용)

	MsgSync MsgType = "SYNC" // worker→sup EVENT: SyncEvent (재접속 직후 자기 procs 스냅샷 자발 보고)
)

// ExecType은 실행 종류를 구분한다(MsgExec 공통 — 같은 PTY 엔진 위의 named recipe).
//
//	EXEC = 일반 인터랙티브 실행(산출물 없음, 종료코드만).
//	EDIT = 에디터로 띄워 편집 → 종료 시 파일 내용 회수(MsgEditResult). 초기 내용은 미리
//	       전송(SendBuffer)해 깔아 둔다(인라인 주입 아님).
//
// 빈 값은 EXEC로 간주한다(하위호환 기본값 — Kind() 사용).
type ExecType string

const (
	ExecTypeExec ExecType = "EXEC"
	ExecTypeEdit ExecType = "EDIT"
)

// Placeholder*는 명령 치환 토큰이다. supervisor가 Cmd/Args/Cwd/Env에 그대로 박아 보내면
// worker가 실행 직전 실제 값으로 치환한다(매칭 안 되는 {..}는 그대로 둔다). 양쪽이 같은
// 문자열을 참조하도록 여기에 둔다 — worker는 치환, supervisor는 명령 빌드에 사용.
// 치환 토큰. supervisor가 Cmd/Args/Cwd/Env/전송 DestPath에 그대로 박아 보내면 worker가
// 실행/저장 직전 실제 값으로 치환한다(매칭 안 되는 {..}는 그대로). 경로 조립은 supervisor가
// process/node 구조체로 하고(→ process.WorkerNodePath), worker는 placeholder만 지역화한다.
const (
	PlaceholderWorkerBase   = "{WORKER_BASE}"   // worker 베이스 디렉토리(ProcessRoot)
	PlaceholderWorkerEditor = "{WORKER_EDITOR}" // worker 기본 텍스트 에디터($VISUAL>$EDITOR>vi)
)

// ProcessSpec: 실행 명세. (MsgExec REQ)
// UID는 initiator가 발급(RandomKey). 와이어 데이터라 execute가 아닌 여기(protocol)에 둔다
// — execute는 Linux 전용(PTY ioctl)이므로 protocol/transport의 이식성을 위해 분리.
// 공통 필드(UID/Cmd/Args/...) + 종류 구분(Type). EDIT의 초기 내용은 별도 전송(SendBuffer)이라
// 여기에 인라인 필드를 두지 않는다.
type ProcessSpec struct {
	UID  string   `json:"uid"`
	Type ExecType `json:"type,omitempty"` // EXEC(기본·빈값) | EDIT
	Cmd  string   `json:"cmd,omitempty"`  // EDIT: 비우면 worker가 $EDITOR 해석
	Args []string `json:"args,omitempty"`
	Cwd  string   `json:"cwd,omitempty"`
	Env  []string `json:"env,omitempty"`  // "KEY=VALUE" 형식
	Rows uint16   `json:"rows,omitempty"` // 초기 터미널 창 크기
	Cols uint16   `json:"cols,omitempty"`
}

// Kind는 Type을 정규화해 반환한다(빈 값 → EXEC). worker/supervisor는 이걸로 분기한다.
func (s ProcessSpec) Kind() ExecType {
	if s.Type == ExecTypeEdit {
		return ExecTypeEdit
	}
	return ExecTypeExec
}

// ExecResponse: 실행 수락/거부. (MsgExec RES)
// 수락 후 실제 RUNNING(+PID)은 별도 MsgStatus(EVENT)로 비동기 통보된다.
type ExecResponse struct {
	Accept bool   `json:"accept"`
	Reason string `json:"reason,omitempty"`
}

// DataEvent: 입출력 바이트. (MsgData EVENT, 양방향)
// 보내는 쪽에 따라 입력(sup→worker)/출력(worker→sup)이 결정된다(받는 쪽 On 핸들러가 구분).
// Data는 Frame.Data(json) 봉투라 base64로 인코딩되어 나간다.
type DataEvent struct {
	UID  string `json:"uid"`
	Data []byte `json:"data"`
}

// ResizeRequest: 터미널 창 크기 변경. (MsgResize REQ) → Interactive.Layout 매핑.
type ResizeRequest struct {
	UID  string `json:"uid"`
	Rows uint16 `json:"rows"`
	Cols uint16 `json:"cols"`
}

// KillRequest: process 종료. (MsgKill REQ)
type KillRequest struct {
	UID string `json:"uid"`
}

// StatusEvent: process 상태 변화 통보. (MsgStatus EVENT, worker→sup)
// Status는 execute.CommandStatus를 그대로 사용한다(execute가 이식 가능해져 미러 불필요).
// PID는 PROCESS(시작) 시 채워진다(worker OS pid, 보고 전용). ExitCode는 종료(Completed/Failed) 시에만 유효.
type StatusEvent struct {
	UID      string                `json:"uid"`
	Status   execute.CommandStatus `json:"status"`
	PID      int                   `json:"pid,omitempty"` // 0 = 아직 모름
	ExitCode int                   `json:"exitCode"`      // Completed/Failed 시에만 의미(0이 정상종료라 omitempty 금지)
}

// EditResult: EDIT 종료 시 worker가 회수한 최종 파일 내용. (MsgEditResult REQ, worker→sup → 빈 응답)
// 저장 여부 판별은 supervisor가 한다: 노드의 현재 content와 비교해 바뀌었을 때만 UPDATE.
// (vi는 :wq·:q! 모두 정상종료라 종료코드로 저장여부 못 가림 → 무조건 read-back 후 supervisor가 diff.)
// 에디터 비정상 종료(예: :cq)는 별도 MsgStatus(Failed)로 통보되니 supervisor가 취소로 처리할 수 있다.
type EditResult struct {
	UID     string `json:"uid"`
	Content []byte `json:"content"`
}

// ===== 재접속 재바인딩 (worker→sup) =====
//
// worker 끊김 시 supervisor는 관련 process를 전부 PENDING으로 낙관적 전이한다(REF-process
// "worker 끊김→PENDING"). 재접속하면 worker가 자기 로컬 procs를 스냅샷해 MsgSync로 자발적
// 보고하고, supervisor는 DB(PENDING/PROCESS)와 3-way 대조해 재바인딩한다. RegisterRequest에
// 얹지 않고 별도 메시지로 분리한다 — 식별자 핸드셰이크와 도메인 상태 동기화는 관심사가 다르다.

// SyncEntry: worker가 보고하는 개별 process 상태 한 줄.
type SyncEntry struct {
	UID    string                `json:"uid"`
	Status execute.CommandStatus `json:"status"`
	PID    int                   `json:"pid,omitempty"`
}

// SyncEvent: MsgSync payload. (worker→sup EVENT, register 성공 직후 1회 자발 전송)
type SyncEvent struct {
	Procs []SyncEntry `json:"procs"`
}

// ===== node 카탈로그 (supervisor → browser) =====
//
// worker는 관여하지 않는다(node는 순수 supervisor PG 영속). REST(CRUD) 커밋 성공 후
// subscribeHub가 NODE:<parentId> 토픽 구독자에게 EVENT로 push한다. 세분화 kind(created/
// moved/renamed 등) 대신 CRUD 3종 + payload=항상 전체 node 구조체로 통일한다(REF-realtime.md
// "Kind(MsgType) 어휘 — node 도메인 확정"). 이동/이름변경은 별도 kind 없이 UPDATE에 흡수.
const (
	MsgNodeCreate MsgType = "NODE:CREATE" // sup→browser EVENT: 전체 node 구조체
	MsgNodeUpdate MsgType = "NODE:UPDATE" // sup→browser EVENT: 전체 node 구조체 (이동/이름변경 포함)
	MsgNodeDelete MsgType = "NODE:DELETE" // sup→browser EVENT: 전체 node 구조체(삭제 직전 스냅샷)
)

// ===== process 객체 변경 통지 (supervisor → browser) =====
//
// PROCESS:<uid> 토픽 위에서 기존 MsgData(출력 바이트)/MsgStatus(상태전이, worker 원본 그대로
// 중계)와 별개로 쓰는 채널이다 — "process 레코드 자체가 바뀌었다"를 알릴 땐 이 kind로 전체
// processResponse DTO를 실어보낸다(node 도메인과 동일 원칙: payload=항상 전체 구조체).
// 첫 사용처는 resize(rows/cols 변경) — worker 확인 후 DB+memory 동기화가 끝난 시점에만 발행.
const (
	MsgProcessUpdate MsgType = "PROCESS:UPDATE" // sup→browser EVENT: 전체 process 구조체
)
