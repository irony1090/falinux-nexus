// Package hold는 보관 프로세스(`worker -hold <소켓>`)와 worker 쪽 연결이다. 보관 프로세스가 PTY·화면 ring을 쥐어
// worker가 꺼져도 스크립트가 살아 있고, worker는 다시 켜지면 소켓으로 다시 붙는다 — REF-process-keep.md
package hold

import (
	"time"

	"nexus/internal/execute"
)

// Version은 보관 프로세스 <-> worker 통신 형식 버전이다. 이미 떠 있는 보관 프로세스는 옛 코드로 계속 돌므로
// 형식을 바꾸면 올리고, 옛 형식과 붙지 못하는 경우를 받아들여야 한다 — K5(통신 형식)
const Version = 1

// ScreenCap은 process당 화면 ring 크기다(REF-process-snapshot.md S5(ring 상한))
const ScreenCap = 256 * 1024

// DoneWait는 스크립트가 끝난 뒤 worker가 종료 코드를 가져가길 기다리는 시간이다 — K7(끝난 뒤 기다림)
const DoneWait = 10 * time.Minute

// 메시지 종류. 보관 -> worker: hello·data·snap·status / worker -> 보관: hello·input·resize·kill·snap·release
const (
	MsgHello   = "hello"
	MsgData    = "data"
	MsgSnap    = "snap"
	MsgStatus  = "status"
	MsgInput   = "input"
	MsgResize  = "resize"
	MsgKill    = "kill"
	MsgRelease = "release" // 종료 코드를 받았음 -> 보관 프로세스 끝냄
)

// Msg는 소켓 위 한 줄(JSON)이다. 종류마다 쓰는 필드만 채운다
type Msg struct {
	T      string                `json:"t"`
	V      int                   `json:"v,omitempty"`
	Err    string                `json:"err,omitempty"`
	Data   []byte                `json:"data,omitempty"`
	Off    int64                 `json:"off,omitempty"` // data·snap: Data 끝의 누적 바이트
	Alt    bool                  `json:"alt,omitempty"`
	Status execute.CommandStatus `json:"status,omitempty"`
	Pid    int                   `json:"pid,omitempty"` // 스크립트 pid
	Exit   int                   `json:"exit,omitempty"`
	Cols   uint16                `json:"cols,omitempty"`
	Rows   uint16                `json:"rows,omitempty"`
	Force  bool                  `json:"force,omitempty"`
}

// Spec은 보관 프로세스가 stdin으로 받는 실행 정보다(Env는 완전 대체 — worker가 조립해 넘김)
type Spec struct {
	UID  string   `json:"uid"`
	Cmd  string   `json:"cmd"`
	Args []string `json:"args"`
	Env  []string `json:"env"`
	Cwd  string   `json:"cwd"`
}

// ready는 보관 프로세스가 준비를 마치고 stdout에 쓰는 한 줄이다
type ready struct {
	Pid int    `json:"pid,omitempty"`
	Err string `json:"err,omitempty"`
}
