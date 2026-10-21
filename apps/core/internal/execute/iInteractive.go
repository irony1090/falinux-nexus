package execute

type IInteractive interface {
	OutputAll() ([]byte, error) // 대기 중인 출력을 한 번에(배치 드레인 — 락 경합·프레임 수 줄임)
	Write(data []byte) error
	Status() (CommandStatus, int, error)
	Kill(force bool) error          // force = SIGKILL, 아니면 SIGTERM(정리 기회)
	Layout(cols, rows uint16) error // 그 자리에서 확인된 결과(로컬 = ioctl 에러, 원격 = worker 응답 에러)
}
