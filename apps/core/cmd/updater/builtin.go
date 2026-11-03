package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// builtin 단계 — updater 안 함수(셸 안 거침). REF-update-deploy.md "steps.json" builtin 표
// 백업 = backups/<배포 id>/{files/, new-files.txt, manifest.json}, 최근 3개. manifest에 그때 TARGET_DIR을 남겨 경로가 바뀐 뒤에도 옛 경로로 복원(정-5(설치 경로))

const keepBackups = 3

type manifest struct {
	DeployID  string    `json:"deployId"`
	RunID     int64     `json:"runId"`
	Version   string    `json:"version"`
	TargetDir string    `json:"targetDir"`
	At        time.Time `json:"at"`
	// PrevInstalled: 백업 때 설치본 이름 — 되돌리기가 성공하면 이 이름으로 돌아감(P3)
	PrevInstalled string `json:"prevInstalled,omitempty"`
}

func (j *job) builtin(name string) error {
	switch name {
	case "backup":
		return j.backup()
	case "copy":
		return j.copyFiles()
	case "restore":
		return j.restore()
	}
	return fmt.Errorf("모르는 builtin %q", name)
}

func (j *job) targetDir() (string, error) {
	t := j.a.cfg.get().TargetDir
	if t == "" || !filepath.IsAbs(t) {
		return "", errors.New("설치 경로(TARGET_DIR)가 설정되지 않았습니다 — 인스턴스 설정에서 지정")
	}
	return t, nil
}

func (j *job) backupRoot() string { return filepath.Join(j.a.dir, "backups") }

// backupID: 배포 id, 개별 실행이면 r<기록 id>
func (j *job) backupID() string {
	if j.run.DeployID != "" {
		return j.run.DeployID
	}
	return fmt.Sprintf("r%d", j.run.RunID)
}

// findBackup: 배포 id의 백업, 개별 실행이면 가장 최근 백업
func (j *job) findBackup() (string, error) {
	if j.run.DeployID != "" {
		d := filepath.Join(j.backupRoot(), j.run.DeployID)
		if _, err := os.Stat(filepath.Join(d, "manifest.json")); err != nil {
			return "", fmt.Errorf("배포 %s의 백업이 없습니다", j.run.DeployID)
		}
		return d, nil
	}
	ents, _ := os.ReadDir(j.backupRoot())
	var best string
	var bestAt time.Time
	for _, e := range ents {
		m, err := readManifest(filepath.Join(j.backupRoot(), e.Name()))
		if err == nil && m.At.After(bestAt) {
			best, bestAt = filepath.Join(j.backupRoot(), e.Name()), m.At
		}
	}
	if best == "" {
		return "", errors.New("백업이 하나도 없습니다")
	}
	return best, nil
}

func readManifest(dir string) (manifest, error) {
	var m manifest
	b, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return m, err
	}
	return m, json.Unmarshal(b, &m)
}

// backup: 덮어쓸 파일 중 이미 있는 것 -> files/, 없는 것 -> new-files.txt
func (j *job) backup() error {
	target, err := j.targetDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(j.backupRoot(), j.backupID())
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dir, "files"), 0o755); err != nil {
		return err
	}
	var news []string
	saved := 0
	for _, f := range j.run.Files {
		if err := j.ctx.Err(); err != nil {
			return err
		}
		src := filepath.Join(target, filepath.FromSlash(f.Path))
		st, err := os.Lstat(src)
		if errors.Is(err, fs.ErrNotExist) {
			news = append(news, f.Path)
			continue
		}
		if err != nil {
			return err
		}
		if !st.Mode().IsRegular() {
			return fmt.Errorf("%s는 보통 파일이 아니라 백업할 수 없습니다", src)
		}
		if _, err := copyOne(src, filepath.Join(dir, "files", filepath.FromSlash(f.Path)), st.Mode().Perm()); err != nil {
			return fmt.Errorf("백업 %s: %w", f.Path, err)
		}
		saved++
	}
	if err := os.WriteFile(filepath.Join(dir, "new-files.txt"), []byte(strings.Join(news, "\n")+"\n"), 0o644); err != nil {
		return err
	}
	m := manifest{DeployID: j.backupID(), RunID: j.run.RunID, Version: j.run.Version, TargetDir: target, At: time.Now(), PrevInstalled: j.a.state().Installed}
	b, _ := json.MarshalIndent(m, "", "  ")
	if err := writeFileAtomic(filepath.Join(dir, "manifest.json"), append(b, '\n'), 0o644); err != nil {
		return err
	}
	j.say("backup %s — 있던 파일 %d개 저장, 새 파일 %d개 기록 (%s)", target, saved, len(news), dir)
	pruneBackups(j.backupRoot(), keepBackups, dir)
	return nil
}

func pruneBackups(root string, keep int, keepAlso string) {
	ents, _ := os.ReadDir(root)
	type b struct {
		dir string
		at  time.Time
	}
	var all []b
	for _, e := range ents {
		d := filepath.Join(root, e.Name())
		m, err := readManifest(d)
		if err != nil {
			continue
		}
		all = append(all, b{d, m.At})
	}
	sort.Slice(all, func(x, y int) bool { return all[x].at.After(all[y].at) })
	for i, x := range all {
		if i >= keep && x.dir != keepAlso {
			_ = os.RemoveAll(x.dir)
		}
	}
}

// copyFiles: staging(합친 결과) -> TARGET_DIR. 임시 이름에 쓴 뒤 이름 변경 — 실행 중 파일도 Linux는 이름 변경으로 바뀜(옛 프로세스는 옛 inode). 지우지 않음
func (j *job) copyFiles() error {
	target, err := j.targetDir()
	if err != nil {
		return err
	}
	for _, f := range j.run.Files { // 하나라도 걸리면 아무것도 안 씀
		if err := j.guardSelf(filepath.Join(target, filepath.FromSlash(f.Path))); err != nil {
			return err
		}
	}
	if _, err := j.findBackup(); err != nil {
		j.say("주의: 이 배포의 backup이 없어 copy를 되돌릴 수 없습니다")
	}
	for _, f := range j.run.Files {
		if err := j.ctx.Err(); err != nil {
			return err
		}
		src := filepath.Join(j.stage, filepath.FromSlash(f.Path))
		dst := filepath.Join(target, filepath.FromSlash(f.Path))
		old, err := copyOne(src, dst, os.FileMode(f.Mode))
		if err != nil {
			return fmt.Errorf("copy %s: %w", f.Path, err)
		}
		j.say("copy %s (%04o)%s", f.Path, f.Mode, movedNote(old))
	}
	j.say("copy %s — 파일 %d개", target, len(j.run.Files))
	return nil
}

// restore: 백업 files/를 그때 경로에 되돌리고 new-files.txt의 파일만 지움
func (j *job) restore() error {
	dir, err := j.findBackup()
	if err != nil {
		return err
	}
	m, err := readManifest(dir)
	if err != nil {
		return err
	}
	if m.TargetDir == "" {
		return errors.New("백업에 설치 경로가 없습니다")
	}
	if now := j.a.cfg.get().TargetDir; now != m.TargetDir {
		j.say("백업 때 경로 %s로 복원합니다(지금 설정 %s)", m.TargetDir, now)
	}
	files := filepath.Join(dir, "files")
	if err := j.guardSelf(m.TargetDir); err != nil {
		return err
	}
	n := 0
	err = filepath.WalkDir(files, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if err := j.ctx.Err(); err != nil {
			return err
		}
		rel, _ := filepath.Rel(files, p)
		st, err := os.Stat(p)
		if err != nil {
			return err
		}
		dst := filepath.Join(m.TargetDir, rel)
		if err := j.guardSelf(dst); err != nil {
			return err
		}
		old, err := copyOne(p, dst, st.Mode().Perm())
		if err != nil {
			return fmt.Errorf("복원 %s: %w", rel, err)
		}
		if old != "" {
			j.say("복원 %s%s", rel, movedNote(old))
		}
		n++
		return nil
	})
	if err != nil {
		return err
	}
	removed := 0
	if fh, err := os.Open(filepath.Join(dir, "new-files.txt")); err == nil {
		sc := bufio.NewScanner(fh)
		for sc.Scan() {
			rel := strings.TrimSpace(sc.Text())
			if rel == "" || !filepath.IsLocal(filepath.FromSlash(rel)) {
				continue
			}
			dst := filepath.Join(m.TargetDir, filepath.FromSlash(rel))
			if j.guardSelf(dst) != nil {
				continue
			}
			if old, err := removeFile(dst); err == nil {
				if old != "" {
					j.say("새 파일 %s 지우기 거절 — 이름만 바꿈%s", rel, movedNote(old))
				}
				removed++
			}
		}
		fh.Close()
	}
	j.say("restore %s (백업 %s, 버전 %s) — 되돌림 %d개, 새 파일 지움 %d개", m.TargetDir, filepath.Base(dir), m.Version, n, removed)
	return nil
}

// copyOne: dst 폴더 안 임시 파일 -> 권한 -> 이름 변경. 돌려주는 값 = 물린 옛 파일(Windows 실행 중 파일, 없으면 "")
func copyOne(src, dst string, mode os.FileMode) (string, error) {
	in, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp(filepath.Dir(dst), "."+filepath.Base(dst)+".nexus-*")
	if err != nil {
		return "", err
	}
	_, err = io.Copy(tmp, in)
	if err == nil {
		err = tmp.Sync()
	}
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Chmod(tmp.Name(), mode)
	}
	var old string
	if err == nil {
		old, err = replaceFile(tmp.Name(), dst)
	}
	if err != nil {
		os.Remove(tmp.Name())
	}
	return old, err
}

// guardSelf: 보완-1(자기 폴더 보호) — TARGET_DIR 안에 updater 폴더가 있을 때(예: C:\Nexus\updater) 빌드 파일이 updater 자기 파일을 덮어쓰지 않게.
// 폴더 전체가 아니라 updater가 쓰는 것만 막는다(설치 경로가 updater 폴더 안이어도 동작하게)
func (j *job) guardSelf(dst string) error {
	own := []string{j.a.cfg.path, j.a.statePath(), "staging", "backups", "logs", "updater.log", "updater.log.1"}
	if exe, err := os.Executable(); err == nil {
		own = append(own, exe)
	}
	for _, p := range own {
		if !filepath.IsAbs(p) {
			p = filepath.Join(j.a.dir, p)
		}
		if rel, err := filepath.Rel(p, dst); err == nil && (rel == "." || filepath.IsLocal(rel)) {
			return fmt.Errorf("%s는 updater가 쓰는 파일이라 쓰지 않습니다(%s)", dst, p)
		}
	}
	return nil
}

func movedNote(old string) string {
	if old == "" {
		return ""
	}
	return " — 실행 중이던 옛 파일 -> " + filepath.Base(old)
}
