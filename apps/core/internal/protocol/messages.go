package protocol

import "fmt"

// 이 파일은 Nexus 도메인 메시지(어휘)를 모은다.
// 봉투(Frame/Kind)는 protocol.go에 있고, 여기엔 "무엇을 주고받는지"만 정의한다.

// MsgRegister는 worker가 supervisor에 접속하며 보내는 등록 요청이다.
const MsgRegister MsgType = "REGISTER"

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
