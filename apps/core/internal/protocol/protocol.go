package protocol

import "encoding/json"

// Kind는 프레임의 성격을 구분한다. 이 한 필드가 "요청인지 응답인지"를
// 프로토콜 레벨에서 알려주어, 요청↔응답 추적을 쉽게 한다.
type Kind uint8

const (
	REQ Kind = iota // 응답을 기대하는 요청 (ID 채번)
	RES             // REQ에 대한 1:1 응답 (REQ의 ID를 그대로 반사)
)

// MsgType은 도메인 메시지 종류다(예제에선 "ADD").
type MsgType string

// Frame은 supervisor ↔ worker가 주고받는 단일 봉투다.
// Data(payload)는 json.RawMessage라 어떤 요청/응답 데이터든 자유롭게 실린다.
type Frame struct {
	Kind Kind            `json:"k"`
	ID   uint64          `json:"id"`          // REQ가 채번, RES가 반사
	Type MsgType         `json:"t"`
	Err  string          `json:"e,omitempty"` // RES 전용: 비어있지 않으면 실패
	Data json.RawMessage `json:"d,omitempty"`
}

// Bind는 Data를 v로 역직렬화한다(받은 쪽에서 payload 꺼낼 때).
func (f Frame) Bind(v any) error {
	if len(f.Data) == 0 {
		return nil
	}
	return json.Unmarshal(f.Data, v)
}

// NewRequest는 채번된 id로 REQ 프레임을 만든다.
func NewRequest(id uint64, t MsgType, data any) (Frame, error) {
	raw, err := marshalData(data)
	if err != nil {
		return Frame{}, err
	}
	return Frame{Kind: REQ, ID: id, Type: t, Data: raw}, nil
}

// NewResponse는 req의 id를 반사한 성공 RES 프레임을 만든다.
func NewResponse(id uint64, t MsgType, data any) (Frame, error) {
	raw, err := marshalData(data)
	if err != nil {
		return Frame{}, err
	}
	return Frame{Kind: RES, ID: id, Type: t, Data: raw}, nil
}

// NewError는 req의 id를 반사한 실패 RES 프레임을 만든다.
func NewError(id uint64, t MsgType, err error) Frame {
	return Frame{Kind: RES, ID: id, Type: t, Err: err.Error()}
}

// Encode/Decode는 프레임 ↔ 바이트(WS 메시지) 변환이다.
func Encode(f Frame) ([]byte, error) { return json.Marshal(f) }

func Decode(b []byte) (Frame, error) {
	var f Frame
	err := json.Unmarshal(b, &f)
	return f, err
}

func marshalData(data any) (json.RawMessage, error) {
	if data == nil {
		return nil, nil
	}
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}
