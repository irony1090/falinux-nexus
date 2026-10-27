package protocol

import (
	"encoding/json"
	"fmt"
	"strings"

	"nexus/internal/execute"
)

// Nexus 도메인 메시지 어휘 — 봉투(Frame/Kind)는 protocol.go

// worker 등록 + 파일 전송(supervisor -> worker)
const (
	MsgRegister   MsgType = "REGISTER"
	MsgFileInit   MsgType = "FILE_INIT"   // 통보 → FileInitResponse
	MsgFileChunk  MsgType = "FILE_CHUNK"  // 데이터 조각 → FileChunkResponse
	MsgFileResult MsgType = "FILE_RESULT" // 전송 완료 통보 → FileResult(검증 결과)
	MsgFileAbort  MsgType = "FILE_ABORT"  // 진행중 중단
)

// RegisterRequest: worker -> supervisor. SubKey = 최초 접속이면 빈 값, 재접속이면 이전에 받은 값
type RegisterRequest struct {
	Key    string `json:"key"`              // 사전 지정 고유키 (여러 worker가 같을 수 있음)
	SubKey string `json:"subKey,omitempty"` // 재접속 시에만 채워짐
	OS     string `json:"os,omitempty"`     // worker runtime.GOOS(빈 값 = 옛 worker = linux) — WW5(OS 표시)
	APIKey string `json:"apiKey,omitempty"` // 장비 API 키 — 회원-4(worker 접속 인증), REF-member.md
}

// NormalizeOS는 빈 OS(옛 worker·이전 명부 행)를 linux로 본다.
func NormalizeOS(os string) string {
	if os == "" {
		return "linux"
	}
	return os
}

func (r RegisterRequest) InstanceKey() string {
	if r.SubKey == "" {
		return r.Key
	} else {
		return fmt.Sprintf("%s#%s", r.Key, r.SubKey)
	}
}

// RegisterResponse: supervisor -> worker. 최초면 새로 부여한 SubKey, 재접속이면 확인된 SubKey
type RegisterResponse struct {
	SubKey string `json:"subKey"`
}

func NewRegisterRequest(key, subKey string) RegisterRequest {
	return RegisterRequest{Key: key, SubKey: subKey}
}

// ===== 파일 전송 (supervisor -> worker) =====
// FileInit -> FileChunk × N -> FileResult, 중단 = FileAbort. 한 연결에 여러 전송이 동시에 오가므로 TransferID로 묶는다(Frame.ID와 별개)

// FileInitRequest: 전송 통보. (REQ → FileInitResponse)
type FileInitRequest struct {
	TransferID string `json:"transferId"`          // 전송마다 새로 발급된 유일 토큰
	Name       string `json:"name"`                // 원본 파일명(표시/로그용). Abs면 DestPath 안에 만들 파일 이름
	DestPath   string `json:"destPath"`            // 수신측 저장 상대경로(루트 기준 — traversal 검증 대상). Abs면 받을 폴더의 절대 경로
	Size       int64  `json:"size"`                // 전체 바이트 수
	Mode       uint32 `json:"mode"`                // 권한 비트(수신측에서 os.FileMode로 변환)
	Hash       string `json:"hash,omitempty"`      // 전체 무결성 해시(sha256 hex 등)
	Resume     bool   `json:"resume,omitempty"`    // 송신자의 이어받기 의향
	Abs        bool   `json:"abs,omitempty"`       // 장비 디렉토리 업로드: DestPath(폴더) + Name을 worker가 합침(OS 경로 규칙), ProcessRoot 검사 안 함, 폴더가 있어야 함
	Exclusive  bool   `json:"exclusive,omitempty"` // 이미 있으면 거절(Code = FileErrExists) — Q23(같은 이름 파일)
}

// FileInitResponse: 수락 여부 + ResumeOffset(받는 쪽이 이미 가진 바이트 수, 0 = 처음부터)
type FileInitResponse struct {
	Accept       bool   `json:"accept"`
	ResumeOffset int64  `json:"resumeOffset"`
	Reason       string `json:"reason,omitempty"` // 거부 사유(accept=false일 때)
	Code         string `json:"code,omitempty"`   // 거부 코드 FileErr*(빈 값 = 그 밖의 거부)
}

// FileInitResponse.Code
const (
	FileErrExists   = "exists"    // 같은 이름 파일이 있음(Exclusive)
	FileErrIsDir    = "is_dir"    // 같은 이름 폴더가 있음
	FileErrNoParent = "no_parent" // 받을 폴더 없음(Abs)
	FileErrBadPath  = "bad_path"  // 경로 이탈·절대 경로 아님·잘못된 파일 이름
)

// ValidFileName: 장비 디렉토리에 만들 파일 이름 — 경로 구분자(`/`·`\`)·`.`·`..`·빈 값 금지(OS 공통 규칙)
func ValidFileName(name string) bool {
	return name != "" && name != "." && name != ".." && !strings.ContainsAny(name, `/\`)
}

// FileChunkRequest: 데이터 한 조각(REQ -> FileChunkResponse). Data는 json 봉투라 base64(~33% 증가)
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

// ===== 파일 읽기 (worker -> supervisor, supervisor가 끌어옴) — REF-transfer-down.md 전송-7(worker 송신) =====
// FileOpen -> FileRead × N -> FileClose. 다음 조각은 supervisor가 요청할 때만 읽으므로 받는 쪽 속도에 맞춰진다
const (
	MsgFileOpen  MsgType = "FILE_OPEN"  // sup→worker REQ: FileOpenRequest → FileOpenResponse
	MsgFileRead  MsgType = "FILE_READ"  // sup→worker REQ: FileReadRequest → FileReadResponse
	MsgFileClose MsgType = "FILE_CLOSE" // sup→worker REQ: FileCloseRequest → 빈 응답
)

// FileOpenRequest: Names가 비면 Path = 파일 하나(그대로) / 있으면 Path = 폴더, 그 안의 Names(파일·폴더)를 zip 하나로
type FileOpenRequest struct {
	ReadID string   `json:"readId"`
	Path   string   `json:"path"`
	Names  []string `json:"names,omitempty"`
}

// FileOpenResponse: Err가 있으면 나머지는 빈 값(DirErr*). Size = -1이면 모름(zip)
type FileOpenResponse struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	Mode    uint32 `json:"mode"`
	ModTime int64  `json:"modTime"` // Unix 초(zip = 0)
	Err     string `json:"err,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

// FileReadRequest: zip은 순서대로만 — Offset = 지금 위치 또는 바로 앞 조각의 시작(응답이 끊겨 다시 묻는 경우)
type FileReadRequest struct {
	ReadID string `json:"readId"`
	Offset int64  `json:"offset"`
	Max    int    `json:"max"`
}

// FileReadResponse: EOF = 이 조각 뒤로 더 없음(Data가 있어도 됨)
type FileReadResponse struct {
	Data []byte `json:"data"`
	EOF  bool   `json:"eof,omitempty"`
}

type FileCloseRequest struct {
	ReadID string `json:"readId"`
}

// ===== worker 디렉토리 목록 (supervisor -> worker) — REF-worker-dir.md 폴더-1(목록 요청) =====
// 첫 요청(Path)에 worker가 전부 읽어 정렬해 보관하고 Cursor를 준다. 다음 페이지 = Cursor + Offset(안 다(worker 캐시))
const MsgDirList MsgType = "DIR_LIST" // sup→worker REQ: DirListRequest → DirListResponse

// DirListRequest: Cursor가 있으면 Path는 무시
type DirListRequest struct {
	Path   string `json:"path,omitempty"` // 빈 값 = 최상위(Linux `/`, Windows 드라이브 목록)
	Cursor string `json:"cursor,omitempty"`
	Offset int    `json:"offset,omitempty"`
}

// DirEntry: 항목 하나. Path는 worker가 이은 전체 경로(프론트는 구분자를 다루지 않음)
type DirEntry struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	Kind       string `json:"kind"` // DirKind*
	Size       int64  `json:"size"`
	ModTime    int64  `json:"modTime"`              // Unix 초
	Exec       bool   `json:"exec,omitempty"`       // 실행 비트(Linux)
	LinkTarget string `json:"linkTarget,omitempty"` // 링크에 적힌 원문
	TargetPath string `json:"targetPath,omitempty"` // 링크 원본 절대 경로(끊긴 링크면 빈 값)
	TargetDir  bool   `json:"targetDir,omitempty"`  // 링크 원본이 폴더
}

// DirCrumb: 경로 줄 한 칸. Path "" = 최상위(Windows 드라이브 목록)
type DirCrumb struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

const (
	DirKindDir   = "dir"
	DirKindFile  = "file"
	DirKindLink  = "link"
	DirKindOther = "other"
)

// 같은 이름 확인 — 올리기 전 Q28(여러 파일 겹침) 창의 근거(목록은 500개씩 페이지라 화면 목록으로는 모자람)
const MsgDirExists MsgType = "DIR_EXISTS" // sup→worker REQ: DirExistsRequest → DirExistsResponse

type DirExistsRequest struct {
	Path  string   `json:"path"` // 절대 폴더
	Names []string `json:"names"`
}

// DirExistsResponse: Found = 이미 있는 이름만(Kind = DirKindDir | DirKindFile, 링크는 원본 기준). 잘못된 이름은 빠짐(올릴 때 거절됨)
type DirExistsResponse struct {
	Err    string     `json:"err,omitempty"` // DirErr*
	Reason string     `json:"reason,omitempty"`
	Found  []DirEntry `json:"found"`
}

// DirListResponse: Err가 있으면 나머지는 빈 값. Parent = nil이면 위가 없음(Linux `/`), ""이면 최상위(Windows 드라이브 목록)
type DirListResponse struct {
	Path    string     `json:"path"`
	Parent  *string    `json:"parent"`
	Crumbs  []DirCrumb `json:"crumbs"` // 최상위부터 지금 폴더까지(경로 줄)
	Cursor  string     `json:"cursor"`
	Total   int        `json:"total"`
	Offset  int        `json:"offset"`
	Entries []DirEntry `json:"entries"`
	Err     string     `json:"err,omitempty"` // DirErr*
	Reason  string     `json:"reason,omitempty"`
}

// 경로 자동완성 — 경로-1(후보 요청), REF-path-input.md. 폴더만, 앞글자 대소문자 무시(목록 캐시 없이 매번 읽음)
const MsgDirSuggest MsgType = "DIR_SUGGEST" // sup→worker REQ: DirSuggestRequest → DirSuggestResponse

// DirSuggestRequest: Max 0 = 후보 없이 Dir가 폴더인지만(Enter 검사)
type DirSuggestRequest struct {
	Dir    string `json:"dir"` // 빈 값 = 최상위(Linux `/`, Windows 드라이브 목록)
	Prefix string `json:"prefix,omitempty"`
	Max    int    `json:"max,omitempty"`
}

// DirSuggestResponse: Dir = worker가 정리한 경로. Items의 Path는 worker가 이은 전체 경로
type DirSuggestResponse struct {
	Dir    string     `json:"dir"`
	Items  []DirCrumb `json:"items"`
	More   bool       `json:"more,omitempty"` // Max를 넘어 잘림
	Err    string     `json:"err,omitempty"`  // DirErr*
	Reason string     `json:"reason,omitempty"`
}

const (
	DirErrNotFound = "not_found"
	DirErrDenied   = "denied"
	DirErrNotDir   = "not_dir"
	DirErrBadPath  = "bad_path"
	DirErrExpired  = "expired" // 보관 목록이 만료됨 -> 처음부터 다시
	DirErrIsDir    = "is_dir"  // FILE_OPEN: 파일 하나를 열려는데 폴더
	DirErrOther    = "other"
)

// ===== process 실행 (supervisor <-> worker) =====
// 제어(REQ/RES) = EXEC·RESIZE·KILL / 스트림(EVENT) = DATA(양방향)·STATUS(worker -> sup)
// UID = supervisor가 발급하는 전체 식별자(OS PID와 별개, PID는 STATUS로 보고)
const (
	MsgExec   MsgType = "EXEC"   // sup→worker REQ: ProcessSpec → ExecResponse
	MsgData   MsgType = "DATA"   // EVENT 양방향: 입력(sup→worker) / 출력(worker→sup)
	MsgResize MsgType = "RESIZE" // sup→worker REQ: ResizeRequest → 빈 응답
	MsgKill   MsgType = "KILL"   // sup→worker REQ: KillRequest → 빈 응답
	MsgStatus MsgType = "STATUS" // worker→sup EVENT: StatusEvent

	MsgEditResult MsgType = "EDIT_RESULT" // worker→sup REQ: EditResult → 빈 응답 (EDIT 전용)

	MsgSync MsgType = "SYNC" // worker→sup EVENT: SyncEvent (재접속 직후 자기 procs 스냅샷 자발 보고)

	MsgScreen MsgType = "SCREEN" // worker→sup EVENT: ScreenEvent (보관 프로세스 화면 스냅샷 — 재접속 뒤 supervisor ring 채우기)

	MsgScreenReq MsgType = "SCREEN_REQ" // sup→worker REQ: ScreenRequest → 빈 응답, 화면은 MsgScreen으로 (가져오기-1(화면 다시 요청), REF-supervisor-move.md)
)

// ScreenRequest: 이 uid 보관 프로세스의 화면 스냅샷을 다시 보내 달라 — 미등록 process를 가져온 직후 그 전 화면을 채움
type ScreenRequest struct {
	UID string `json:"uid"`
}

// ExecType: EXEC = 실행 / EDIT = 미리 보낸 파일을 에디터로 열고 종료 때 내용 회수(MsgEditResult). 빈 값 = EXEC(Kind())
type ExecType string

const (
	ExecTypeExec ExecType = "EXEC"
	ExecTypeEdit ExecType = "EDIT"
)

// Placeholder*: supervisor가 Cmd/Args/Cwd/Env/DestPath에 넣어 보내면 worker가 실행·저장 직전 실제 값으로 바꾼다(모르는 {..}는 그대로)
const (
	PlaceholderWorkerBase   = "{WORKER_BASE}"   // worker 베이스 디렉토리(ProcessRoot)
	PlaceholderWorkerEditor = "{WORKER_EDITOR}" // worker 기본 텍스트 에디터($VISUAL>$EDITOR>OS 기본값)
	PlaceholderWorkerShell  = "{WORKER_SHELL}"  // worker 대화형 셸 절대 경로(Linux $SHELL>bash / Windows powershell) — 장비 타일 "여기서 셸"
)

// ProcessSpec: 실행 명세(MsgExec REQ). EDIT 초기 내용은 별도 전송(SendBuffer)이라 필드 없음
type ProcessSpec struct {
	UID  string   `json:"uid"`
	Type ExecType `json:"type,omitempty"` // EXEC(기본·빈값) | EDIT
	Cmd  string   `json:"cmd,omitempty"`  // EDIT: 비우면 worker가 $EDITOR 해석
	Args []string `json:"args,omitempty"`
	Cwd  string   `json:"cwd,omitempty"`
	Env  []string `json:"env,omitempty"`  // "KEY=VALUE" 형식
	Rows uint16   `json:"rows,omitempty"` // 초기 터미널 창 크기
	Cols uint16   `json:"cols,omitempty"`

	Direct bool `json:"direct,omitempty"` // 장비 파일을 그대로 실행(형식 변환·확장자 강제 없음, 실행 가능한 파일만) — 폴더-4(실행)

	Label string `json:"label,omitempty"` // 표시용 이름(스크립트 노드 이름) — worker가 보관 목록에 두고 미등록 process 표시에 씀(주소-4)
}

// Kind: 빈 값 = EXEC
func (s ProcessSpec) Kind() ExecType {
	if s.Type == ExecTypeEdit {
		return ExecTypeEdit
	}
	return ExecTypeExec
}

// ExecResponse: 실행 수락/거부(MsgExec RES). RUNNING(+PID)은 뒤에 STATUS로
type ExecResponse struct {
	Accept bool   `json:"accept"`
	Reason string `json:"reason,omitempty"`
}

// DataEvent: 입출력 바이트(MsgData, 방향 = 보내는 쪽). Data는 json 봉투라 base64
type DataEvent struct {
	UID  string `json:"uid"`
	Data []byte `json:"data"`
	Off  int64  `json:"off,omitempty"` // 출력만: 이 묶음까지의 누적 바이트(보관 프로세스가 매김, 화면 복원 이음매용)
}

// ScreenEvent: 보관 프로세스 화면 스냅샷(MsgScreen). Off = Data 끝의 누적 바이트 — REF-process-keep.md 살리기-6(화면 내용 이전)
type ScreenEvent struct {
	UID  string `json:"uid"`
	Data []byte `json:"data"`
	Off  int64  `json:"off"`
	Alt  bool   `json:"alt"`
}

// ResizeRequest: 터미널 창 크기(MsgResize REQ)
type ResizeRequest struct {
	UID  string `json:"uid"`
	Rows uint16 `json:"rows"`
	Cols uint16 `json:"cols"`
}

// KillRequest: process 종료. (MsgKill REQ) Force = SIGKILL, 아니면 SIGTERM — REF-process-kill.md
type KillRequest struct {
	UID   string `json:"uid"`
	Force bool   `json:"force,omitempty"`
}

// StatusEvent: 상태 변화(MsgStatus, worker -> sup). PID = 시작 때(보고 전용)
type StatusEvent struct {
	UID      string                `json:"uid"`
	Status   execute.CommandStatus `json:"status"`
	PID      int                   `json:"pid,omitempty"` // 0 = 아직 모름
	ExitCode int                   `json:"exitCode"`      // Completed/Failed 시에만 의미(0이 정상종료라 omitempty 금지)
}

// EditResult: EDIT 종료 때 회수한 파일 내용(MsgEditResult REQ). vi는 :wq·:q! 모두 정상 종료라 supervisor가 content와 비교해 저장 여부를 가린다
type EditResult struct {
	UID     string `json:"uid"`
	Content []byte `json:"content"`
}

// ===== 재접속 재바인딩 (worker -> sup) =====
// 재접속한 worker가 자기 procs를 MsgSync로 보고 -> supervisor가 DB와 대조(REF-process "재접속 재바인딩")

// SyncEntry: worker가 보고하는 개별 process 상태 한 줄.
type SyncEntry struct {
	UID    string                `json:"uid"`
	Status execute.CommandStatus `json:"status"`
	PID    int                   `json:"pid,omitempty"`
	// 아래는 미등록 process(붙은 supervisor DB에 행이 없음) 표시용 — 주소-4(미등록 process), REF-supervisor-move.md
	Kind      ExecType `json:"kind,omitempty"`
	Label     string   `json:"label,omitempty"` // 스크립트 노드 이름(장비 파일 실행은 빈 값)
	Cmd       string   `json:"cmd,omitempty"`   // 명령 + 인자(표시용 한 줄). 이 기능 전에 띄운 process는 빈 값
	Cwd       string   `json:"cwd,omitempty"`
	StartedAt int64    `json:"startedAt,omitempty"` // Unix 초
}

// SyncEvent: MsgSync payload — register 성공 직후 1회
type SyncEvent struct {
	Procs []SyncEntry `json:"procs"`
}

// ===== node 카탈로그 (supervisor -> browser) =====
// 볼 수 있는 접속 계정의 USER:<userId> 토픽, CRUD 3종 + payload = 전체 node(이동·이름변경 = UPDATE, REF-node-list-sync.md)
const (
	MsgNodeCreate MsgType = "NODE:CREATE" // sup→browser EVENT: 전체 node 구조체
	MsgNodeUpdate MsgType = "NODE:UPDATE" // sup→browser EVENT: 전체 node 구조체 (이동/이름변경 포함)
	MsgNodeDelete MsgType = "NODE:DELETE" // sup→browser EVENT: 전체 node 구조체(삭제 직전 스냅샷)
)

// ===== process 변경 통지 (supervisor -> browser) =====
// PROCESS:<uid> 토픽에서 DATA·STATUS와 별개로 process 레코드 변경(크기·종료 요청)을 전체 DTO로
const (
	MsgProcessUpdate MsgType = "PROCESS:UPDATE" // sup→browser EVENT: 전체 process 구조체

	// 토픽 발행 아님 — 새 크기 소유자 탭의 소켓에만 직접 Emit
	MsgProcessSizeOwner MsgType = "PROCESS:SIZE_OWNER" // sup→browser EVENT: SizeOwnerEvent

	// browser→sup EVENT: DataEvent(키 입력 바이트), 응답 없음 — REF-process-input.md I1(전달 경로)
	MsgProcessInput MsgType = "PROCESS:INPUT"
)

// SizeOwnerEvent: 받는 탭이 크기 소유자가 됨 / Lost = 잃음(크기-2(잃음 알림)), By = 가져간 다른 계정 닉네임(K5(빼앗긴 탭 안내))
type SizeOwnerEvent struct {
	UID  string `json:"uid"`
	Lost bool   `json:"lost,omitempty"`
	By   string `json:"by,omitempty"`
}

// 소켓 연결 직후 서버가 그 소켓에 가장 먼저 보내는 탭 id (REF-process-sync.md V(탭 id 전달))
const MsgTabID MsgType = "TAB:ID" // sup→browser EVENT: TabIDEvent

type TabIDEvent struct {
	TabID string `json:"tabId"`
}

// 탭이 지금 보는 작업 공간 — 라우트가 바뀔 때마다·연결 직후 1번. Hash "" = 기본(`/`) — WS7(보는 작업 공간 전달)
const MsgTabView MsgType = "TAB:VIEW" // browser→sup EVENT: TabViewEvent

type TabViewEvent struct {
	Hash string `json:"hash"`
}

// 탭의 active 타일 대상 — 바뀔 때만·연결 직후 1번. 둘 다 없음 = 해제. 터미널 = UID / 폴더 = NodeID (REF-node-ui-active.md)
const MsgTabActive MsgType = "TAB:ACTIVE" // browser→sup EVENT: TabActiveEvent

type TabActiveEvent struct {
	UID    string `json:"uid,omitempty"`
	NodeID *int64 `json:"nodeId,omitempty"`
}

// 다른 계정 탭의 active 변경 — USER:<userId> 토픽. 받는 계정이 볼 수 없는 대상이면 비워서(= 지움) 보낸다
const MsgActiveUpdate MsgType = "ACTIVE:UPDATE" // sup→browser EVENT: ActiveUpdateEvent

// ActiveUpdateEvent: Key = 탭 id 해시 앞 8자(탭 id 원문 = 크기 우선권 열쇠라 내보내지 않음). GET /tabs/active 응답 항목과 같은 모양
type ActiveUpdateEvent struct {
	Key      string `json:"key"`
	UserID   int64  `json:"userId"`
	Nickname string `json:"nickname"`
	UID      string `json:"uid,omitempty"`
	NodeID   *int64 `json:"nodeId,omitempty"`
}

// 계정의 타일 트리 저장 통지 — USER:<userId> 토픽, 4 O(계정 동기화) 화면 (REF-node-ui-sync.md)
const MsgTilesUpdate MsgType = "TILES:UPDATE" // sup→browser EVENT: TilesUpdateEvent

// TilesUpdateEvent: 트리 전체 + version. TabID = 저장한 탭(자기 것 거름, 4②(내 변경 되받기)), Hash = 보는 보드가 아니면 무시(WS6(TILES 이벤트 구분))
type TilesUpdateEvent struct {
	Hash    string          `json:"hash"`
	Tree    json.RawMessage `json:"tree"`
	Version int64           `json:"version"`
	TabID   string          `json:"tabId"`
}

// 폴더 타일 고정 지도 — USER:<userId> 토픽, 트리 저장마다 계정 보드 전체로 다시 계산(개편-1(고정 지도), REF-tile-binding-rework.md)
const MsgTilesPins MsgType = "TILES:PINS" // sup→browser EVENT: TilesPinsEvent

// TilePin: 같은 폴더 노드의 타일이 내 보드를 통틀어 2개 이상일 때 그 노드의 고정 타일 — R4(고정 표시 남발)
type TilePin struct {
	NodeID int64  `json:"nodeId"`
	Hash   string `json:"hash"`
	TileID string `json:"tileId"`
}

type TilesPinsEvent struct {
	Pins []TilePin `json:"pins"`
}

// 타일 위치 알림 — USER:<userId> 토픽, 토스트·알림 내역용(개편-5(알림 이벤트), REF-tile-binding-rework.md "위치 알림")
const MsgTilesNotice MsgType = "TILES:NOTICE" // sup→browser EVENT: TilesNoticeEvent

const (
	NoticeBind     = "BIND"      // 현재 타일에 바인딩 — M1(내가 바인딩)·M3(다른 곳으로 빠짐)은 프론트가 보는 보드로 가름
	NoticeAttach   = "ATTACH"    // 다른 계정 실행으로 붙음 — M2(자동으로 붙음)
	NoticeNoPlace  = "NO_PLACE"  // 다른 계정 실행, 안 붙음 — M4(위치 없음)
	NoticePinMoved = "PIN_MOVED" // 타일 고정이 다른 타일로 — M5(고정 이동)
)

type TileSpot struct {
	Hash   string `json:"hash"`
	TileID string `json:"tileId"`
}

// TilesNoticeEvent: UID = 터미널(PIN_MOVED는 없음), NodeID·Name = 스크립트(PIN_MOVED는 폴더)
type TilesNoticeEvent struct {
	Kind   string    `json:"kind"`
	UID    string    `json:"uid,omitempty"`
	NodeID int64     `json:"nodeId"`
	Name   string    `json:"name"`
	From   *TileSpot `json:"from,omitempty"`
	To     *TileSpot `json:"to,omitempty"`
	At     int64     `json:"at"` // unix ms
}

// 공유 변경(걸기·풀기·단계 변경·노드 이동·삭제) — USER:<userId> 토픽(SD4), 받은 쪽이 다시 조회한다(SD6). REF-process-share-reconcile.md
const MsgShareChange MsgType = "SHARE:CHANGE" // sup→browser EVENT: ShareChangeEvent

type ShareChangeEvent struct{}

// 작업 공간 목록 변경(만들기·이름·삭제) — USER:<userId> 토픽, 목록 전체를 보낸다 (REF-workspace-impl.md)
const MsgWorkspacesUpdate MsgType = "WORKSPACES:UPDATE" // sup→browser EVENT: WorkspacesUpdateEvent

type WorkspaceItem struct {
	Hash      string `json:"hash"`
	Name      string `json:"name"`
	IsDefault bool   `json:"isDefault"`
	Ord       int32  `json:"ord"`
}

type WorkspacesUpdateEvent struct {
	Workspaces []WorkspaceItem `json:"workspaces"`
}

// 전송 목록 — USER:<userId> 토픽, 파트(파일 전송 확장) 폴더-8(전송 상태) (REF-transfer-ext.md)
const (
	MsgTransferUpdate MsgType = "TRANSFER:UPDATE" // sup→browser EVENT: TransferInfo(새로 생김·진행·끝 모두)
	MsgTransferGone   MsgType = "TRANSFER:GONE"   // sup→browser EVENT: TransferGoneEvent(목록에서 빠짐)
)

const (
	TransferKindUpload   = "upload"
	TransferKindDownload = "download" // 전송-8·9(내려받기)
	TransferKindCopy     = "copy"     // 전송-10(worker 간 중계)

	TransferPhaseUpload   = "upload"   // 브라우저 -> supervisor — Q25(진행률 구간)
	TransferPhaseSend     = "send"     // supervisor -> worker(worker 간 중계도 이 단계 하나)
	TransferPhaseDownload = "download" // worker -> 브라우저

	TransferRunning = "running"
	TransferDone    = "done"
	TransferFailed  = "failed"
	TransferAborted = "aborted"
)

// TransferInfo: 전송 항목 하나. GET /transfers 응답 항목과 같은 모양
type TransferInfo struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	TileID      string `json:"tileId,omitempty"` // 시작한 장비 타일 — Q27(타일 막대 범위)
	InstanceKey string `json:"instanceKey"`      // 받는 장비(내려받기 = 보내는 장비)
	Dir         string `json:"dir"`              // 받을 폴더(내려받기 = 내려받는 폴더)
	Name        string `json:"name"`
	Size        int64  `json:"size"`               // -1 = 모름(zip)
	FromKey     string `json:"fromKey,omitempty"`  // worker 간 중계: 보내는 장비
	FromPath    string `json:"fromPath,omitempty"` // worker 간 중계: 보내는 파일
	Done        int64  `json:"done"`               // 지금 단계에서 옮긴 바이트
	Phase       string `json:"phase"`              // TransferPhase*
	State       string `json:"state"`              // TransferRunning | Done | Failed | Aborted
	Code        string `json:"code,omitempty"`     // 실패 코드(FileErr* 등)
	Reason      string `json:"reason,omitempty"`
	StartedAt   int64  `json:"startedAt"` // unix ms
}

type TransferGoneEvent struct {
	ID string `json:"id"`
}

// ===== supervisor 주소 옮기기 (supervisor -> worker) — 주소-2(옮기기 명령), REF-supervisor-move.md =====
// worker는 주소 목록(최대 3)을 순위대로 시도한다. 옮기기 = 새 주소를 1순위로 저장, Now면 지금 연결도 끊어 바로 옮김
const (
	MsgMove  MsgType = "MOVE"  // sup→worker REQ: MoveRequest → AddrsResponse
	MsgAddrs MsgType = "ADDRS" // sup→worker REQ: 빈 요청 → AddrsResponse
)

type MoveRequest struct {
	URL string `json:"url"` // scheme://host (ws·wss)
	Now bool   `json:"now"` // 주소-Q9(즉시 옮김 여부): true = 바로 옮기기 / false = 저장만
}

// AddrsResponse: 저장된 목록(1순위부터)
type AddrsResponse struct {
	Addrs []string `json:"addrs"`
}
