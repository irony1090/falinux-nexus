// Package updstore: 업데이트 app 파일 보관소(UPDATE_STORE_DIR). 내용 = 파일시스템, 메타 = DB — REF-update-deploy.md "릴리스 보관 구조"
//
//	devices/<메인키>/<버전>/<경로>             빌드 본체
//	devices/<메인키>/<버전>/@<서브키>/<경로>   한 인스턴스에만 줄 파일(P1은 자리만)
//	common/<이름>/<버전>/<경로>                공용 묶음
//
// 메인키 "common"과 부딪히지 않게 빌드는 devices/ 아래에 둔다
package updstore

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	KindBuild  = "build"
	KindCommon = "common"
	ScopeMain  = "main"
	ScopeSub   = "sub"

	markFile = ".nexus-db" // 폴더 주인 DB
	tmpDir   = ".tmp"
)

// NamePattern: 메인키·버전·공용 이름(메인키는 회원 장비 이름과 같은 규칙)
var NamePattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

type Ref struct{ Kind, Owner, Version string }

func (r Ref) String() string { return r.Owner + "@" + r.Version }

type Store struct{ dir string } // dir "" = 끔

// Open: 다른 DB의 supervisor가 같은 폴더를 쓰면 메타와 내용이 어긋난다 -> 표시 파일이 다르면 끔(viewCache와 같은 방식)
func Open(dir, owner string) *Store {
	if dir == "" {
		log.Printf("[update] 보관소 끔(UPDATE_STORE_DIR 빈 값)")
		return &Store{}
	}
	if err := os.MkdirAll(filepath.Join(dir, tmpDir), 0o755); err != nil {
		log.Printf("[update] 보관소 폴더를 만들 수 없음 — 끔: %v", err)
		return &Store{}
	}
	mark := filepath.Join(dir, markFile)
	if b, err := os.ReadFile(mark); err == nil {
		if got := strings.TrimSpace(string(b)); got != owner {
			log.Printf("[update] 경고: 보관소 %s 는 다른 DB(%s) 것 — 끔. UPDATE_STORE_DIR을 따로 지정할 것", dir, got)
			return &Store{}
		}
	} else if err := os.WriteFile(mark, []byte(owner+"\n"), 0o644); err != nil {
		log.Printf("[update] 보관소 표시 파일을 쓸 수 없음 — 끔: %v", err)
		return &Store{}
	}
	// 지난 실행에서 남은 받는 중 파일
	if ents, err := os.ReadDir(filepath.Join(dir, tmpDir)); err == nil {
		for _, e := range ents {
			_ = os.Remove(filepath.Join(dir, tmpDir, e.Name()))
		}
	}
	log.Printf("[update] 보관소 %s", dir)
	return &Store{dir: dir}
}

func (s *Store) Enabled() bool { return s.dir != "" }

func (s *Store) bundleDir(r Ref) string {
	top := "devices"
	if r.Kind == KindCommon {
		top = "common"
	}
	return filepath.Join(s.dir, top, r.Owner, r.Version)
}

// file: rel은 CleanRel을 거친 값이어야 한다
func (s *Store) file(r Ref, scope, sub, rel string) string {
	if scope == ScopeSub {
		return filepath.Join(s.bundleDir(r), "@"+sub, filepath.FromSlash(rel))
	}
	return filepath.Join(s.bundleDir(r), filepath.FromSlash(rel))
}

// CleanRel: 번들 안 상대 경로(`/` 구분). 빈 값·절대·`..`·첫 칸 `@`(서브키 자리)·숨김 받는 중 폴더 거절
func CleanRel(p string) (string, error) {
	p = strings.ReplaceAll(strings.TrimSpace(p), "\\", "/")
	if p == "" || strings.HasPrefix(p, "/") || strings.ContainsRune(p, 0) {
		return "", errors.New("경로가 비었거나 절대 경로입니다")
	}
	c := path.Clean(p)
	if c == "." || c == ".." || strings.HasPrefix(c, "../") {
		return "", errors.New("번들 밖을 가리키는 경로입니다")
	}
	if len(c) > 512 {
		return "", errors.New("경로가 너무 깁니다")
	}
	for _, seg := range strings.Split(c, "/") {
		if strings.HasSuffix(seg, " ") || strings.HasSuffix(seg, ".") {
			return "", fmt.Errorf("이름 끝에 공백·점을 쓸 수 없습니다: %q", seg)
		}
	}
	if strings.HasPrefix(c, "@") {
		return "", errors.New("첫 폴더 이름은 @로 시작할 수 없습니다(서브키 자리)")
	}
	return c, nil
}

// Temp: 받는 중 파일 — Commit으로 제자리에 옮기거나 Discard
func (s *Store) Temp() (*os.File, error) {
	return os.CreateTemp(filepath.Join(s.dir, tmpDir), "up-*")
}

func Discard(f *os.File) {
	f.Close()
	_ = os.Remove(f.Name())
}

// HashFile: 크기·sha256(hex)
func HashFile(name string) (int64, string, error) {
	f, err := os.Open(name)
	if err != nil {
		return 0, "", err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil {
		return 0, "", err
	}
	return n, hex.EncodeToString(h.Sum(nil)), nil
}

// Commit: 받은 임시 파일을 제자리로(같은 파일시스템 안 이름 변경). 덮어쓰기 = 이름 변경이라 하드링크로 나눈 복사본에는 영향 없음
func (s *Store) Commit(tmpName string, r Ref, scope, sub, rel string) error {
	dst := s.file(r, scope, sub, rel)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.Rename(tmpName, dst)
}

// WriteBytes: 텍스트 저장 — 임시 파일에 쓰고 Commit
func (s *Store) WriteBytes(r Ref, scope, sub, rel string, b []byte) error {
	f, err := s.Temp()
	if err != nil {
		return err
	}
	if _, err := f.Write(b); err != nil {
		Discard(f)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(f.Name())
		return err
	}
	if err := s.Commit(f.Name(), r, scope, sub, rel); err != nil {
		_ = os.Remove(f.Name())
		return err
	}
	return nil
}

func (s *Store) OpenFile(r Ref, scope, sub, rel string) (*os.File, error) {
	return os.Open(s.file(r, scope, sub, rel))
}

// Remove: 파일 하나 + 비게 된 폴더(번들 폴더까지는 남김)
func (s *Store) Remove(r Ref, scope, sub, rel string) error {
	name := s.file(r, scope, sub, rel)
	if err := os.Remove(name); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	root := s.bundleDir(r)
	for d := filepath.Dir(name); d != root && strings.HasPrefix(d, root); d = filepath.Dir(d) {
		if os.Remove(d) != nil {
			break
		}
	}
	return nil
}

// RemoveBundle: 번들 폴더 통째로 + 비게 된 주인 폴더
func (s *Store) RemoveBundle(r Ref) error {
	dir := s.bundleDir(r)
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	_ = os.Remove(filepath.Dir(dir))
	return nil
}

// RemoveOwner: 장비 루트 폴더(빌드가 없을 때만 부른다)
func (s *Store) RemoveOwner(kind, owner string) error {
	return os.RemoveAll(filepath.Dir(s.bundleDir(Ref{Kind: kind, Owner: owner, Version: "x"})))
}

// MkBundle: 빈 번들 폴더
func (s *Store) MkBundle(r Ref) error { return os.MkdirAll(s.bundleDir(r), 0o755) }

// CopyBundle: 하드링크(안 되면 복사). 이미 dst가 있으면 거절
func (s *Store) CopyBundle(src, dst Ref) error {
	from, to := s.bundleDir(src), s.bundleDir(dst)
	if _, err := os.Stat(to); err == nil {
		return fmt.Errorf("보관소에 %s 폴더가 이미 있습니다", dst)
	}
	if _, err := os.Stat(from); errors.Is(err, fs.ErrNotExist) {
		return os.MkdirAll(to, 0o755) // 빈 번들
	}
	return filepath.WalkDir(from, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(from, p)
		target := filepath.Join(to, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if os.Link(p, target) == nil {
			return nil
		}
		return copyFile(p, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// Entry: 폴더 sha 계산 재료
type Entry struct{ Scope, Sub, Path, Sha string }

// FolderSha: 정렬한 (범위, 서브키, 경로, 파일 sha) 목록의 sha256. 빈 번들 = ""
func FolderSha(es []Entry) string {
	if len(es) == 0 {
		return ""
	}
	es = append([]Entry(nil), es...)
	sort.Slice(es, func(i, j int) bool {
		a, b := es[i], es[j]
		if a.Scope != b.Scope {
			return a.Scope < b.Scope
		}
		if a.Sub != b.Sub {
			return a.Sub < b.Sub
		}
		return a.Path < b.Path
	})
	h := sha256.New()
	for _, e := range es {
		fmt.Fprintf(h, "%s\x00%s\x00%s\x00%s\n", e.Scope, e.Sub, e.Path, e.Sha)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// TextMax: 웹에서 열어 고칠 수 있는 텍스트 파일 상한
const TextMax = 1 << 20

// IsText: 앞 8KB가 UTF-8이고 NUL이 없으면 텍스트(크기 상한 안)
func (s *Store) IsText(r Ref, scope, sub, rel string, size int64) bool {
	if size > TextMax {
		return false
	}
	f, err := s.OpenFile(r, scope, sub, rel)
	if err != nil {
		return false
	}
	defer f.Close()
	buf := make([]byte, 8<<10)
	n, _ := io.ReadFull(f, buf)
	return LooksText(buf[:n], int64(n) < size)
}

// LooksText: NUL이 없고 제어 문자(탭·줄바꿈·ESC 등 글에 흔한 것 제외)가 1% 이하인 UTF-8. cut = 잘라 읽음(끝의 깨진 UTF-8 한 글자는 봐준다)
func LooksText(b []byte, cut bool) bool {
	ctrl := 0
	for _, c := range b {
		switch {
		case c == 0:
			return false
		case c < 0x20 && c != '\t' && c != '\n' && c != '\r' && c != '\f' && c != '\b' && c != 0x1b:
			ctrl++
		}
	}
	if ctrl*100 > len(b) {
		return false
	}
	if cut {
		for i := 0; i < 4 && len(b) > 0 && !utf8.Valid(b); i++ {
			b = b[:len(b)-1]
		}
	}
	return utf8.Valid(b)
}
