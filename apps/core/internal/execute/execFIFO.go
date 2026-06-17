package execute

import (
	"bufio"
	"bytes"
	"fmt"
	"nexus/internal/syncProcess"
	"os"
	"os/user"
	"path/filepath"
	"syscall"
)

type Fifo struct {
	id     string
	path   string
	output *syncProcess.SyncData[[]byte]
	status *syncProcess.SyncData[CommandStatus]

	err error

	closeBytes []byte
	done       chan struct{} // 완료 대기용 채널
}

func (f *Fifo) createFifo() error {

	dir := filepath.Dir(f.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if _, err := os.Stat(f.path); os.IsNotExist(err) {
		if err := syscall.Mkfifo(f.path, 0666); err != nil {
			return err
		}
	}
	return nil
}
func (f *Fifo) pending() {
	err := f.createFifo()
	// log.Printf("생성 %v", err)
	f.output = syncProcess.NewSyncData([][]byte{})
	f.status = syncProcess.NewSyncData([]CommandStatus{})
	f.done = make(chan struct{}) // 완료 채널 초기화
	if err == nil {
		f.err = nil
		f.status.Push(CommandPending)
	} else {
		f.err = err
		f.status.Push(CommandFailed)
	}
}

func (f *Fifo) process() {

	go func() {
		for {
			// log.Print("OPEN")
			file, err := os.Open(f.path)
			// log.Print("OPENED")
			if err != nil {
				f.complete(err)
				return
			}
			scanner := bufio.NewScanner(file)
			var flag bool
			for scanner.Scan() {
				b := scanner.Bytes()
				flag = bytes.Equal(b, f.closeBytes)
				if flag {
					break
				}
				f.output.Push(bytes.Clone(b))
			}
			file.Close()
			if err := scanner.Err(); err != nil {
				f.complete(err)
				return
			}
			if flag {
				break
			}
		}
		f.complete(nil)
		// f.status.Push(CommandCompleted)
	}()

	f.status.Push(CommandProcess)
}

func (f *Fifo) complete(err error) {

	if _, er := os.Stat(f.path); er == nil {
		// go func() {
		// 	time.Sleep(1 * time.Second)
		// 	log.Printf("[삭제] %s %v", f.path, er)

		// }()
		er := os.Remove(f.path)
		f.err = er
	} else {
		f.err = err
	}
	if f.err == nil {
		f.status.Push(CommandCompleted)
	} else {
		f.status.Push(CommandFailed)
	}

	f.output.Close()
	f.status.Close()
	close(f.done) // 완료 신호
}

func (f *Fifo) Connect() error {
	f.pending()
	if f.err != nil {
		return f.err
	}
	f.process()
	return nil
}

func (f *Fifo) Kill() error {
	if f.done == nil {
		return fmt.Errorf("fifo not connected")
	}

	err := f.Write(f.closeBytes)
	// 완료될 때까지 대기
	<-f.done
	return err
}

func (f *Fifo) Status() (CommandStatus, error) {
	if f.done == nil {
		return CommandPending, fmt.Errorf("fifo not connected")
	}

	sts, _ := f.status.Shift()
	return sts, f.err
}

func (f *Fifo) Output() ([]byte, error) {
	if f.done == nil {
		return nil, fmt.Errorf("fifo not connected")
	}

	return f.output.Shift()
}
func (f *Fifo) Write(data []byte) error {
	if f.done == nil {
		return fmt.Errorf("fifo not connected")
	}

	file, err := os.OpenFile(f.path, os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	// closeBytes + 개행 전송 (scanner.Scan()은 줄 단위)
	_, err = file.Write(append(data, '\n'))
	file.Close()
	return nil
}

func (f *Fifo) Path() string {
	return f.path
}
func (f *Fifo) GetStatus() CommandStatus {
	if f.done == nil {
		return CommandPending
	}
	return f.status.Last()
}

func NewFifo(id string, closeBytes []byte) *Fifo {
	user, err := user.Current()
	var uid string
	if err != nil {
		uid = "-9999"
	} else {
		uid = user.Uid
	}
	// user = -9999
	// if err != nil {
	// 	return nil
	// }
	fifoPath := fmt.Sprintf("%s/%s/%s.fifo", os.TempDir(), uid, id)

	return &Fifo{
		id:         id,
		path:       fifoPath,
		closeBytes: closeBytes,
		// output:     syncProcess.NewSyncData([][]byte{}),
		// status:     syncProcess.NewSyncData([]CommandStatus{}),
	}
}
