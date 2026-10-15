package router

import (
	"log"
	"time"
)

// ===== worker 종료 시 실행 중 process 정리 (REF-worker-orphan 안 다(정상 종료 정리)) =====

// admit은 종료 중이 아니면 process를 procs에 넣는다.
// closing 확인과 Append를 Shutdown과 같은 락 안에서 하므로, 여기서 들어간 process는
// 반드시 Shutdown의 killAll 대상에 잡힌다(락 밖에서 하면 killAll 직후 끼어든 process가 고아로 남음).
func (s *WorkerState) admit(uid string, entry *procEntry) (ok bool, reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closing {
		return false, "worker 종료 중"
	}
	if !s.procs.Append(uid, entry) {
		return false, "중복되는 uid"
	}
	return true, ""
}

// Shutdown은 새 exec를 막고 실행 중 process에 SIGTERM, grace 안에 안 끝나면 SIGKILL한다.
// 끝난 process는 기존 pumpStatus 경로로 최종 상태를 supervisor에 보고한다(종료-2(최종 상태 보고)).
func (s *WorkerState) Shutdown(grace time.Duration) {
	s.mu.Lock()
	s.closing = true
	s.mu.Unlock()

	s.KillAll(false)
	if s.waitEmpty(grace) {
		return
	}
	log.Printf("[worker shutdown] %v 안에 안 끝난 process 강제 종료", grace)
	s.KillAll(true)
	s.waitEmpty(time.Second)
}

// KillAll은 procs 전부에 Kill을 보낸다(force = SIGKILL).
func (s *WorkerState) KillAll(force bool) {
	for _, e := range s.procs.FindAll(func(string, *procEntry) bool { return true }) {
		if err := e.Val.inter.Kill(force); err != nil {
			log.Printf("[worker shutdown] kill uid=%s force=%v: %v", e.Key, force, err)
		}
	}
}

// waitEmpty는 procs가 빌 때까지(teardown이 Remove) 최대 d 동안 기다린다.
func (s *WorkerState) waitEmpty(d time.Duration) bool {
	deadline := time.Now().Add(d)
	for {
		if len(s.procs.FindAll(func(string, *procEntry) bool { return true })) == 0 {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(100 * time.Millisecond)
	}
}
