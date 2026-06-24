package protocol

import "fmt"

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
