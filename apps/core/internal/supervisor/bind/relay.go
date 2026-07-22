package bind

import (
	"log"
	"sync"

	"nexus/internal/execute"
	"nexus/internal/protocol"
)

// Publish는 한 토픽으로 이벤트를 내보내는 전략이다. Relay가 subscribe.Hub의 제네릭
// 타입(Conn/MsgType)에 직접 묶이지 않도록, 호출자가 토픽 키를 클로저로 감싸 주입한다.
//
//	예) bind.NewRelay(uid, inter, func(k protocol.MsgType, p any) error {
//	        return hub.Publish("PROCESS:"+uid, k, p)
//	    })
type Publish func(kind protocol.MsgType, payload any) error

// Relay는 한 process(IInteractive)의 출력/상태를 드레인해 구독자에게 중계한다.
// IInteractive 위에서 동작하므로 supervisor 프록시(AgentInteractive)든 로컬 PTY든 재사용된다.
// (REF-process: "허브는 sink 모름 — bind 계층이 배선".)
//
// 생명주기: Start로 pump 2개 기동 → AgentInteractive.Done()이 채널을 닫으면
// 두 pump가 err을 받고 자연 종료 → Wait로 합류.
type Relay struct {
	uid     string
	inter   execute.IInteractive
	publish Publish

	// TODO(후순위): ring buffer(SNAPSHOT용) / DB sink(EXEC만 on) 훅 배선 지점.

	wg   sync.WaitGroup
	once sync.Once
}

// NewRelay는 릴레이를 만든다(아직 goroutine 미기동 — Start에서 기동).
func NewRelay(uid string, inter execute.IInteractive, publish Publish) *Relay {
	return &Relay{uid: uid, inter: inter, publish: publish}
}

// Start는 출력/상태 pump goroutine을 각각 기동한다. 중복 호출은 무시된다.
func (r *Relay) Start() {
	r.once.Do(func() {
		r.wg.Add(2)
		go r.pumpOutput()
		go r.pumpStatus()
	})
}

// Wait는 두 pump가 끝날 때까지(=process 종료) 블록한다.
func (r *Relay) Wait() { r.wg.Wait() }

// pumpOutput은 OutputAll()로 대기분을 배치 드레인해 MsgData(EVENT ↓) 한 프레임으로 발행한다.
// OutputAll()이 err(채널 close)을 주면 종료한다.
func (r *Relay) pumpOutput() {
	defer r.wg.Done()
	for {
		data, err := r.inter.OutputAll()
		if err != nil {
			return // Done() 시 output 채널 close → 정상 종료
		}
		// TODO: ring buffer append / DB sink(EXEC만) 배선 지점.
		if err := r.publish(protocol.MsgData, protocol.DataEvent{UID: r.uid, Data: data}); err != nil {
			log.Printf("[bind] publish data uid=%s: %v", r.uid, err)
		}
	}
}

// pumpStatus는 Status()를 드레인해 MsgStatus(EVENT ↓)로 발행한다.
// 주의: AgentInteractive.Status()는 PID를 담지 않는다(PID는 worker 원본 StatusEvent에만).
// 따라서 pool 상태 갱신(MarkProcessRunning[+pid]/MarkProcessDone)은 PID를 아는
// router의 On(MsgStatus) 핸들러에서 처리하고, 여기선 frontend fan-out만 맡는 걸 권장.
func (r *Relay) pumpStatus() {
	defer r.wg.Done()
	for {
		st, code, err := r.inter.Status()
		if err != nil {
			return
		}
		if err := r.publish(protocol.MsgStatus, protocol.StatusEvent{
			UID:      r.uid,
			Status:   st,
			ExitCode: code,
		}); err != nil {
			log.Printf("[bind] publish status uid=%s: %v", r.uid, err)
		}
	}
}
