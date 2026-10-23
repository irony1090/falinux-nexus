package router

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"nexus/internal/protocol"
)

// 파일 읽기 — 전송-7(worker 송신), REF-transfer-down.md. supervisor가 FILE_READ로 조각을 끌어간다
const (
	readExpire   = 60 * time.Second // FILE_READ 사이 idle 허용(supervisor가 사라지면 닫음)
	readChunkMax = 1 << 20
)

// fileRead: 열린 읽기 하나. 파일 하나 = ReadAt(offset 자유) / zip = pipe(순서대로만)
type fileRead struct {
	mu      sync.Mutex
	file    *os.File
	pipe    *io.PipeReader
	pos     int64  // zip: 다음 읽을 위치
	last    []byte // zip: 바로 앞 조각 — 응답이 끊겨 같은 offset을 다시 물을 때
	lastAt  int64
	lastEOF bool
	timer   *time.Timer
}

// fileOpen: FILE_OPEN 핸들러. 실패는 RES 에러가 아니라 Err 코드로(supervisor가 HTTP 상태로 바꿈)
func (r *workerRouter) fileOpen(req protocol.Frame) (any, error) {
	var body protocol.FileOpenRequest
	if err := req.Bind(&body); err != nil {
		return nil, err
	}
	if !filepath.IsAbs(body.Path) {
		return openFail(dirError{protocol.DirErrBadPath, "절대 경로가 아닙니다: " + body.Path}), nil
	}

	var fr *fileRead
	var res protocol.FileOpenResponse
	if len(body.Names) == 0 {
		f, err := os.Open(body.Path)
		if err != nil {
			return openFail(err), nil
		}
		st, err := f.Stat()
		if err != nil {
			f.Close()
			return openFail(err), nil
		}
		if st.IsDir() {
			f.Close()
			return openFail(dirError{protocol.DirErrIsDir, "폴더입니다: " + body.Path}), nil
		}
		if !st.Mode().IsRegular() {
			f.Close()
			return openFail(dirError{protocol.DirErrOther, "일반 파일이 아닙니다: " + body.Path}), nil
		}
		fr = &fileRead{file: f}
		res = protocol.FileOpenResponse{Name: filepath.Base(body.Path), Size: st.Size(), Mode: uint32(st.Mode().Perm()), ModTime: st.ModTime().Unix()}
	} else {
		for _, name := range body.Names {
			if !protocol.ValidFileName(name) {
				return openFail(dirError{protocol.DirErrBadPath, "잘못된 이름: " + name}), nil
			}
			if _, err := os.Stat(filepath.Join(body.Path, name)); err != nil {
				return openFail(err), nil
			}
		}
		pr, pw := io.Pipe()
		go writeZip(pw, body.Path, body.Names)
		fr = &fileRead{pipe: pr}
		res = protocol.FileOpenResponse{Name: zipName(body.Path, body.Names), Size: -1}
	}

	fr.timer = time.AfterFunc(readExpire, func() { r.closeRead(body.ReadID) })
	if !r.reads.Append(body.ReadID, fr) {
		fr.close()
		return nil, fmt.Errorf("readId 중복: %s", body.ReadID)
	}
	return res, nil
}

func openFail(err error) protocol.FileOpenResponse {
	d := dirFail(err)
	return protocol.FileOpenResponse{Err: d.Err, Reason: d.Reason}
}

// zipName: 하나면 그 이름, 여러 개면 담은 폴더 이름(최상위면 files)
func zipName(dir string, names []string) string {
	base := names[0]
	if len(names) > 1 {
		base = filepath.Base(dir)
		if base == string(filepath.Separator) || base == "." || filepath.VolumeName(dir)+string(filepath.Separator) == dir {
			base = "files"
		}
	}
	return base + ".zip"
}

// writeZip: dir 안의 names를 zip으로 pw에 쓴다. 읽을 수 없는 하위 항목은 건너뛰고, 쓰기 실패(중단 포함)면 멈춘다
func writeZip(pw *io.PipeWriter, dir string, names []string) {
	zw := zip.NewWriter(pw)
	err := func() error {
		for _, name := range names {
			if err := addZipTree(zw, filepath.Join(dir, name), name); err != nil {
				return err
			}
		}
		return zw.Close()
	}()
	if err != nil && !errors.Is(err, io.ErrClosedPipe) {
		log.Printf("[WORKER-READ] zip 실패 %s: %v", dir, err)
	}
	pw.CloseWithError(err) // nil = 읽는 쪽 EOF
}

// addZipTree: root(선택한 항목)를 zip 안 prefix 아래로. 선택한 항목이 폴더 링크면 따라가고, 그 아래 폴더 링크는 따라가지 않는다(순환 방지)
func addZipTree(zw *zip.Writer, root, prefix string) error {
	walkRoot := root
	if st, err := os.Lstat(root); err == nil && st.Mode()&fs.ModeSymlink != 0 {
		if t, err := filepath.EvalSymlinks(root); err == nil {
			if tst, err := os.Stat(t); err == nil && tst.IsDir() {
				walkRoot = t
			}
		}
	}
	return filepath.WalkDir(walkRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if p == walkRoot {
				return err
			}
			return nil
		}
		rel, err := filepath.Rel(walkRoot, p)
		if err != nil {
			return nil
		}
		zname := path.Join(prefix, filepath.ToSlash(rel))
		info, err := os.Stat(p) // 파일 링크는 원본 기준
		if err != nil {
			return nil
		}
		switch {
		case d.IsDir():
			hdr, err := zip.FileInfoHeader(info)
			if err != nil {
				return nil
			}
			hdr.Name = zname + "/"
			_, err = zw.CreateHeader(hdr)
			return err
		case info.Mode().IsRegular():
			return addZipFile(zw, p, zname, info)
		}
		return nil // 폴더 링크·장치 파일 등
	})
}

func addZipFile(zw *zip.Writer, p, zname string, info fs.FileInfo) error {
	f, err := os.Open(p)
	if err != nil {
		return nil // 권한 없음 등은 건너뜀
	}
	defer f.Close()
	hdr, err := zip.FileInfoHeader(info)
	if err != nil {
		return nil
	}
	hdr.Name = zname
	hdr.Method = zip.Deflate
	w, err := zw.CreateHeader(hdr)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	return err
}

// fileReadChunk: FILE_READ 핸들러
func (r *workerRouter) fileReadChunk(req protocol.Frame) (any, error) {
	var body protocol.FileReadRequest
	if err := req.Bind(&body); err != nil {
		return nil, err
	}
	fr, ok := r.reads.Get(body.ReadID)
	if !ok {
		return nil, fmt.Errorf("알 수 없는 읽기: %s", body.ReadID)
	}
	fr.timer.Reset(readExpire)
	return fr.read(body.Offset, min(max(body.Max, 1), readChunkMax))
}

func (fr *fileRead) read(off int64, n int) (protocol.FileReadResponse, error) {
	fr.mu.Lock()
	defer fr.mu.Unlock()
	buf := make([]byte, n)
	if fr.file != nil {
		got, err := fr.file.ReadAt(buf, off)
		if err != nil && !errors.Is(err, io.EOF) {
			return protocol.FileReadResponse{}, err
		}
		return protocol.FileReadResponse{Data: buf[:got], EOF: errors.Is(err, io.EOF)}, nil
	}

	if off != fr.pos {
		if off == fr.lastAt && fr.last != nil {
			return protocol.FileReadResponse{Data: fr.last, EOF: fr.lastEOF}, nil
		}
		return protocol.FileReadResponse{}, fmt.Errorf("zip은 순서대로만 읽습니다(지금 %d, 요청 %d)", fr.pos, off)
	}
	got, err := io.ReadFull(fr.pipe, buf)
	eof := errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF)
	if err != nil && !eof {
		return protocol.FileReadResponse{}, err
	}
	fr.last, fr.lastAt, fr.lastEOF = buf[:got], fr.pos, eof
	fr.pos += int64(got)
	return protocol.FileReadResponse{Data: buf[:got], EOF: eof}, nil
}

// fileClose: FILE_CLOSE 핸들러
func (r *workerRouter) fileClose(req protocol.Frame) (any, error) {
	var body protocol.FileCloseRequest
	if err := req.Bind(&body); err != nil {
		return nil, err
	}
	r.closeRead(body.ReadID)
	return nil, nil
}

func (r *workerRouter) closeRead(id string) {
	if fr, ok := r.reads.Get(id); ok {
		r.reads.Remove(id)
		fr.close()
	}
}

// close: zip이면 pipe를 닫아 writeZip 고루틴이 쓰기 실패로 끝나게 한다. mu를 잡지 않는다(read가 pipe에서 기다리는 중일 수 있음)
func (fr *fileRead) close() {
	fr.timer.Stop()
	if fr.file != nil {
		fr.file.Close()
	}
	if fr.pipe != nil {
		fr.pipe.Close()
	}
}
