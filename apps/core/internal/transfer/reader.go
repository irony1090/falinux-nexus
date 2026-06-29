package transfer

// Reader는 송신부(SendFile/SendBuffer/send)가 청크 전송에 필요로 하는 순수
// 바이트 스트림 계약이다. 디스크 파일(ReadFile)이든 메모리 버퍼(ReadBuffer)든
// 같은 인터페이스로 보내기 위함이다.
//
// 이름(destPath)·권한(perm) 같은 메타데이터는 여기 두지 않는다 — 그건 전송
// 세션(router의 sendJob)이 들고 있고, 각 래퍼가 구체 reader에서 뽑아 넘긴다.
//
// SetOnClose만 예외적으로 둔다: OnClose는 "필드"라 인터페이스로 노출할 수 없는데,
// transferId를 캡처하는 정리 콜백은 transferId가 생성되는 send() 안에서만 걸 수
// 있다(구체 래퍼에선 아직 transferId가 없다). 그래서 세터 메서드로 우회한다.
type Reader interface {
	Size() int64
	Hash() (string, error)
	SeekTo(offset int64) error
	Read(chunkSize int64) ([]byte, error)
	Close()
	SetOnClose(fn func())
}
