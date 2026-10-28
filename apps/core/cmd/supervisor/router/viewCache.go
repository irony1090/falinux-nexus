package router

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"nexus/cmd/supervisor/constants"
	"nexus/internal/supervisor/store"
	"nexus/internal/util"
)

// 이미지 사본 — 파트(보기 타일) 2단계, REF-view-tile.md. 보기 타일이 트리에 있는 동안만 보관(안 1(타일 존재 기준)), 영상은 사본 없음.
// 파일 = <폴더>/<userId>/<sha256(instance, path)> + 같은 이름 .json(크기·수정 시각), 받는 중 = .tmp-<임의>

const (
	viewCacheMax   = 20 << 20 // V3(사본 상한) — 넘으면 그때그때 전달
	viewSweepEvery = 10 * time.Minute
	viewSweepGrace = time.Minute // 막 만든 사본은 남김 — 트리 저장(묶어 보냄)보다 이미지 요청이 먼저 올 수 있음
	viewCacheMark  = ".nexus-db" // 폴더 주인 DB
)

type viewCache struct{ dir string } // dir "" = 끔

type viewMeta struct {
	Size    int64 `json:"size"`
	ModTime int64 `json:"modTime"`
}

// newViewCache: 다른 DB의 supervisor가 같은 폴더를 쓰면 정리가 "내 트리에 없는 사본"을 지워 서로의 사본을 지운다 -> 표시 파일이 다르면 끔
func newViewCache(env constants.EnvVars) *viewCache {
	dir := env.ViewCache
	if dir == "" {
		log.Printf("[view] 이미지 사본 끔(VIEW_CACHE_DIR 빈 값)")
		return &viewCache{}
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Printf("[view] 사본 폴더를 만들 수 없음 — 사본 끔: %v", err)
		return &viewCache{}
	}
	owner := fmt.Sprintf("%s:%d/%s", env.DBHost, env.DBPort, env.DBName)
	mark := filepath.Join(dir, viewCacheMark)
	if b, err := os.ReadFile(mark); err == nil {
		if got := strings.TrimSpace(string(b)); got != owner {
			log.Printf("[view] 경고: 사본 폴더 %s 는 다른 DB(%s) 것 — 사본 끔. VIEW_CACHE_DIR을 따로 지정할 것", dir, got)
			return &viewCache{}
		}
	} else if err := os.WriteFile(mark, []byte(owner+"\n"), 0o644); err != nil {
		log.Printf("[view] 사본 폴더 표시 파일을 쓸 수 없음 — 사본 끔: %v", err)
		return &viewCache{}
	}
	log.Printf("[view] 이미지 사본 폴더 %s", dir)
	return &viewCache{dir: dir}
}

// viewKey: 같은 사용자의 같은 파일 타일 여러 개 = 사본 하나
func viewKey(instance, path string) string {
	sum := sha256.Sum256([]byte(instance + "\x00" + path))
	return hex.EncodeToString(sum[:])
}

func (v *viewCache) base(userID int64, instance, path string) string {
	return filepath.Join(v.dir, strconv.FormatInt(userID, 10), viewKey(instance, path))
}

// get: 사본 본문 + 메타. 없거나 본문 크기가 메타와 다르면 ok = false
func (v *viewCache) get(userID int64, instance, path string) (*os.File, viewMeta, bool) {
	var m viewMeta
	if v.dir == "" {
		return nil, m, false
	}
	base := v.base(userID, instance, path)
	b, err := os.ReadFile(base + ".json")
	if err != nil || json.Unmarshal(b, &m) != nil {
		return nil, m, false
	}
	f, err := os.Open(base)
	if err != nil {
		return nil, m, false
	}
	if st, err := f.Stat(); err != nil || st.Size() != m.Size {
		f.Close()
		return nil, m, false
	}
	return f, m, true
}

// viewWriter: 받는 중인 사본. nil = 안 만듦(끔·상한 넘음·실패) — 메서드는 nil이면 아무것도 안 함
type viewWriter struct {
	f    *os.File
	base string
	meta viewMeta
}

func (v *viewCache) create(userID int64, instance, path string, m viewMeta) *viewWriter {
	if v.dir == "" || m.Size > viewCacheMax {
		return nil
	}
	base := v.base(userID, instance, path)
	if err := os.MkdirAll(filepath.Dir(base), 0o755); err != nil {
		log.Printf("[view] 사본 폴더 생성 실패: %v", err)
		return nil
	}
	id, err := util.RandomKey(8, "", "")
	if err != nil {
		return nil
	}
	f, err := os.Create(base + ".tmp-" + id)
	if err != nil {
		log.Printf("[view] 사본 임시 파일 생성 실패: %v", err)
		return nil
	}
	return &viewWriter{f: f, base: base, meta: m}
}

// write: 실패하면 사본만 포기(표시와 무관) — 이후 호출은 아무것도 안 함
func (w *viewWriter) write(b []byte) *viewWriter {
	if w == nil {
		return nil
	}
	if _, err := w.f.Write(b); err != nil {
		log.Printf("[view] 사본 쓰기 실패: %v", err)
		w.abort()
		return nil
	}
	return w
}

func (w *viewWriter) abort() {
	if w == nil {
		return
	}
	w.f.Close()
	os.Remove(w.f.Name())
}

// commit 순서: 옛 메타 삭제 -> 본문 이름 바꿈 -> 새 메타. 사이에 읽는 쪽은 메타가 없어 사본 없음으로 본다(새 본문 + 옛 ETag 방지)
func (w *viewWriter) commit() {
	if w == nil {
		return
	}
	if err := w.f.Close(); err != nil {
		os.Remove(w.f.Name())
		return
	}
	os.Remove(w.base + ".json")
	if err := os.Rename(w.f.Name(), w.base); err != nil {
		log.Printf("[view] 사본 저장 실패: %v", err)
		os.Remove(w.f.Name())
		return
	}
	b, _ := json.Marshal(w.meta)
	if err := os.WriteFile(w.base+".json", b, 0o644); err != nil {
		log.Printf("[view] 사본 메타 저장 실패: %v", err)
	}
}

// runViewSweep: 시작 때 한 번 + 주기 정리 — 트리를 쓰는 곳이 여러 곳(PUT·배치·정리·보드)이라 저장마다가 아니라 주기로(REF-view-tile.md)
func (r *supervisorRouter) runViewSweep() {
	if r.view.dir == "" {
		return
	}
	for {
		if err := r.view.sweep(); err != nil {
			log.Printf("[view] 사본 정리 실패: %v", err)
		}
		time.Sleep(viewSweepEvery)
	}
}

// sweep: 어느 트리에도 보기 타일이 없는 사본 + 남은 임시 파일을 지운다(viewSweepGrace 안에 만든 것은 남김)
func (v *viewCache) sweep() error {
	rows, err := store.GetStorePool().Queries().ListViewerTiles(context.Background())
	if err != nil {
		return err
	}
	keep := map[string]bool{}
	for _, t := range rows {
		keep[filepath.Join(strconv.FormatInt(t.UserID, 10), viewKey(t.Instance, t.Path))] = true
	}
	users, err := os.ReadDir(v.dir)
	if err != nil {
		return err
	}
	removed := 0
	for _, u := range users {
		if !u.IsDir() {
			continue
		}
		udir := filepath.Join(v.dir, u.Name())
		files, err := os.ReadDir(udir)
		if err != nil {
			continue
		}
		for _, f := range files {
			name := f.Name()
			key, _, _ := strings.Cut(name, ".")
			tmp := strings.Contains(name, ".tmp-")
			if !tmp && keep[filepath.Join(u.Name(), key)] {
				continue
			}
			if info, err := f.Info(); err != nil || time.Since(info.ModTime()) < viewSweepGrace {
				continue
			}
			if os.Remove(filepath.Join(udir, name)) == nil {
				removed++
			}
		}
		os.Remove(udir) // 비었을 때만 지워짐
	}
	if removed > 0 {
		log.Printf("[view] 사본 정리: 파일 %d개 삭제", removed)
	}
	return nil
}
