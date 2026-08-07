package execute

type IInteractive interface {
	Output() ([]byte, error)
	// OutputAll은 대기 중인 출력을 한 번에(연결된 바이트로) 반환한다. 고빈도 스트림에서
	// 항목당 락을 잡는 Output 대신 배치 드레인해 락 경합과 프레임 수를 줄인다.
	OutputAll() ([]byte, error)
	Write(data []byte) error
	Status() (CommandStatus, int, error)
	Kill() error
	// Layout은 터미널 창 크기를 바꾼다. 로컬 PTY는 ioctl 결과(syscall.Errno, error 인터페이스
	// 구현체라 그대로 반환 가능)를, 원격(AgentInteractive)은 worker.Call의 실제 응답 에러를
	// 반환한다 — 둘 다 "그 자리에서 확인된 진짜 결과"를 돌려준다는 계약은 동일.
	Layout(cols, rows uint16) error
	ExitCode() int
}
