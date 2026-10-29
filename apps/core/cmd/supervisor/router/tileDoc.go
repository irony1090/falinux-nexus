package router

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	"nexus/internal/util"
	"nexus/internal/web"
)

// 타일 트리 문서(tileDoc) — 모양 검사·직렬화·타일 찾기

// 프론트 tileTree.store의 Tile 미러. 서버는 모양만 검사하고 순서·닫기 규칙은 프론트 몫
type tileSize struct {
	W float64 `json:"w"`
	H float64 `json:"h"`
}

type tileItem struct {
	ID           string   `json:"id"`
	Type         string   `json:"type"` // tileKinds의 키(tileKind.go)
	Parent       *string  `json:"parent"`
	Kids         []string `json:"kids"`
	Size         tileSize `json:"size"`
	NodeID       *int64   `json:"nodeId"`        // folder: null = 루트 목록 / terminal: 필수
	UID          string   `json:"uid,omitempty"` // terminal: process uid / folder: 서버가 발급하는 타일 uid(FU1), 공유로 생긴 받는 쪽 타일은 공유한 타일과 같은 값(FU2)
	Adopted      bool     `json:"adopted,omitempty"`
	Minimized    bool     `json:"minimized,omitempty"`    // 배치에서만 빠짐(트리·uid 유지) — SD3(실행 중 터미널 치우기)
	ShortcutsOff bool     `json:"shortcutsOff,omitempty"` // terminal: 이 타일이 active인 동안 앱 단축키 끔 — 이동-3(타일별 단축키 끔)
	ShownAt      int64    `json:"shownAt,omitempty"`      // folder: 지금 노드를 보여 주기 시작한 시각(unix ms, 서버만 씀) — BD3(저장 위치)
	Bound        bool     `json:"bound,omitempty"`        // folder: 터미널 받는 타일로 고정(서버만 씀) — BD4(설정 방법)
	Instance     string   `json:"instance,omitempty"`     // device·viewer: worker 인스턴스(main#sub)
	Path         string   `json:"path,omitempty"`         // device: 지금 경로("" = 최상위) / viewer: 파일 경로
	Back         []string `json:"back,omitempty"`         // 이동 기록(folder = 노드 id, "" = 루트 목록 / device = 경로) — 뒤로-1(타일 안 이동 기록)
	Fwd          []string `json:"fwd,omitempty"`
}

type tileDoc struct {
	RootID string              `json:"rootId"`
	Tiles  map[string]tileItem `json:"tiles"`
}

func newRootDoc() tileDoc {
	root := newFolderTile(nil, nil, "")
	return tileDoc{RootID: root.ID, Tiles: map[string]tileItem{root.ID: root}}
}

func validSpan(v float64) bool { return v == 0.5 || v == 1 }

// validate: 루트 1개 · parent/kids 쌍 일치 · 루트에서 전부 닿음(고아·순환 없음) · type/size · 능력(tileKind)과 필드
func (d *tileDoc) validate() error {
	root, ok := d.Tiles[d.RootID]
	if !ok || root.Parent != nil || !root.canRoot() {
		return errors.New("루트 타일이 없거나 올바르지 않습니다")
	}
	for key, t := range d.Tiles {
		if t.ID == "" || t.ID != key {
			return fmt.Errorf("타일 id가 키와 다릅니다: %q", key)
		}
		if !validSpan(t.Size.W) || !validSpan(t.Size.H) {
			return fmt.Errorf("타일 크기가 올바르지 않습니다: %s", key)
		}
		k, ok := tileKinds[t.Type]
		if !ok {
			return fmt.Errorf("타일 type이 올바르지 않습니다: %s", key)
		}
		if !k.holdsProc && (t.Adopted || t.Minimized || t.ShortcutsOff) {
			return fmt.Errorf("process를 담지 않는 타일에 터미널 필드가 있습니다: %s", key)
		}
		if len(t.Back)+len(t.Fwd) > tileHistMax {
			return fmt.Errorf("이동 기록이 너무 깁니다: %s", key)
		}
		if !k.canParent && len(t.Kids) > 0 {
			return fmt.Errorf("자식을 가질 수 없는 타일입니다: %s", key)
		}
		if k.check != nil {
			if err := k.check(t); err != nil {
				return fmt.Errorf("%w: %s", err, key)
			}
		}
		if t.Parent == nil {
			if key != d.RootID {
				return fmt.Errorf("루트가 아닌 타일에 parent가 없습니다: %s", key)
			}
			continue
		}
		p, ok := d.Tiles[*t.Parent]
		if !ok || !p.canParent() {
			return fmt.Errorf("parent가 없거나 자식을 가질 수 없는 타일입니다: %s", key)
		}
		n := 0
		for _, k := range p.Kids {
			if k == key {
				n++
			}
		}
		if n != 1 {
			return fmt.Errorf("parent의 kids에 정확히 한 번 있어야 합니다: %s", key)
		}
	}
	// 위에서 모든 parent->kid 쌍을 확인했으니, kids 쪽만 보면 루트에서 닿는지 알 수 있다
	seen := map[string]bool{d.RootID: true}
	queue := []string{d.RootID}
	for len(queue) > 0 {
		t := d.Tiles[queue[0]]
		queue = queue[1:]
		for _, k := range t.Kids {
			kid, ok := d.Tiles[k]
			if !ok || kid.Parent == nil || *kid.Parent != t.ID || seen[k] {
				return fmt.Errorf("kids가 올바르지 않습니다: %s -> %s", t.ID, k)
			}
			seen[k] = true
			queue = append(queue, k)
		}
	}
	if len(seen) != len(d.Tiles) {
		return errors.New("루트에서 닿지 않는 타일이 있습니다")
	}
	return nil
}

// marshal: 모르는 필드는 버리고 서버 기준 모양으로 저장. kids는 null 대신 [], 폴더 타일의 빈 uid는 발급·빈 shownAt은 지금 — 모든 저장이 여기를 지난다
func (d *tileDoc) marshal() ([]byte, error) {
	now := time.Now().UnixMilli()
	for key, t := range d.Tiles {
		if t.Kids == nil {
			t.Kids = []string{}
		}
		if t.Type == tileFolder && t.UID == "" {
			uid, err := util.RandomKey(16, "", "")
			if err != nil {
				return nil, err
			}
			t.UID = uid
		}
		if t.Type == tileFolder && t.ShownAt == 0 {
			t.ShownAt = now
		}
		d.Tiles[key] = t
	}
	return json.Marshal(d)
}

// keepFolderMeta: 프론트가 보낸 폴더 uid·shownAt·bound는 안 믿고 저장본 값을 쓴다(없으면 marshal이 채움, FU2·FU3)
// shownAt·bound는 보여 주는 노드가 그대로일 때만 — 다른 폴더로 가면 "새로 연" 것(BD3(저장 위치))
func (d *tileDoc) keepFolderMeta(stored *tileDoc) {
	for key, t := range d.Tiles {
		t.ShownAt, t.Bound = 0, false
		if t.Type == tileFolder {
			t.UID = ""
			if s, ok := stored.Tiles[key]; ok && s.Type == tileFolder {
				t.UID = s.UID
				if sameNode(s.NodeID, t.NodeID) {
					t.ShownAt, t.Bound = s.ShownAt, s.Bound
				}
			}
		}
		d.Tiles[key] = t
	}
}

func sameNode(a, b *int64) bool {
	return (a == nil && b == nil) || (a != nil && b != nil && *a == *b)
}

func (d *tileDoc) mustMarshal() []byte {
	b, err := d.marshal()
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return b
}

func mustParseDoc(tree []byte) tileDoc {
	var doc tileDoc
	if err := json.Unmarshal(tree, &doc); err != nil {
		panic(web.Err(500, "저장된 타일 트리를 읽을 수 없습니다: %v", err))
	}
	return doc
}

// checkTerminalUIDs: 새 터미널 uid 추가·한 트리 안 중복을 막는다 — 터미널 타일은 서버만 넣으므로 계정 안 uid가 하나로 유지(WS5(터미널 타일 추가 주체))
func checkTerminalUIDs(next, stored *tileDoc) error {
	had := map[string]bool{}
	for _, t := range stored.Tiles {
		if t.holdsProc() {
			had[t.UID] = true
		}
	}
	seen := map[string]bool{}
	for _, t := range next.Tiles {
		if !t.holdsProc() {
			continue
		}
		if !had[t.UID] {
			return fmt.Errorf("터미널 타일은 서버만 추가할 수 있습니다: %s", t.UID)
		}
		if seen[t.UID] {
			return fmt.Errorf("같은 터미널이 두 번 있습니다: %s", t.UID)
		}
		seen[t.UID] = true
	}
	return nil
}

// checkVideoTiles: 영상 보기 타일은 서버만 넣는다(openVideoTile, MEMORY "타일 추가 주체") — 저장본에 같은 id·파일로 있던 것만 통과
func checkVideoTiles(next, stored *tileDoc) error {
	for id, t := range next.Tiles {
		if t.Type != tileViewer || !isVideoPath(t.Path) {
			continue
		}
		if s, ok := stored.Tiles[id]; !ok || s.Type != tileViewer || s.Instance != t.Instance || s.Path != t.Path {
			return fmt.Errorf("영상 보기 타일은 서버만 추가할 수 있습니다: %s", id)
		}
	}
	return nil
}

// closedVideos: 저장본에 있었는데 새 트리에서 빠진 영상 보기 타일 — PUT에서 빠짐 = 닫기(보드 간 이동은 서버 동작이라 PUT으로 오지 않음)
func closedVideos(next, stored *tileDoc) []tileItem {
	var out []tileItem
	for id, t := range stored.Tiles {
		if t.Type != tileViewer || !isVideoPath(t.Path) {
			continue
		}
		if _, ok := next.Tiles[id]; !ok {
			out = append(out, t)
		}
	}
	return out
}

// closedTerminals: 저장본에 있었는데 새 트리에서 빠진 터미널 uid(사용자 저장에서 빠짐 = 닫기, checkTerminalUIDs가 추가는 막음)
func closedTerminals(next, stored *tileDoc) []string {
	keep := map[string]bool{}
	for _, uid := range next.terminalUIDs() {
		keep[uid] = true
	}
	var out []string
	for _, uid := range stored.terminalUIDs() {
		if !keep[uid] {
			out = append(out, uid)
		}
	}
	return out
}

// removeLeaf: 자식 없는 타일(process 타일)을 부모 kids와 트리에서 뺀다
func (d *tileDoc) removeLeaf(id string) {
	t := d.Tiles[id]
	if t.Parent != nil {
		p := d.Tiles[*t.Parent]
		p.Kids = slices.DeleteFunc(slices.Clone(p.Kids), func(k string) bool { return k == id })
		d.Tiles[p.ID] = p
	}
	delete(d.Tiles, id)
}

func (d *tileDoc) hasFolderUID(uid string) bool {
	for _, t := range d.Tiles {
		if t.Type == tileFolder && t.UID == uid {
			return true
		}
	}
	return false
}

// folderTileOf: 그 노드를 보여 주는 폴더 타일 중 루트에서 너비 우선으로 처음 닿는 것(map 순서에 기대지 않게)
func (d *tileDoc) folderTileOf(nodeID int64) (string, bool) {
	ids := d.folderTilesOf(nodeID)
	if len(ids) == 0 {
		return "", false
	}
	return ids[0], true
}

// folderTilesOf: 그 노드를 보여 주는 폴더 타일 전부(너비 우선 순)
func (d *tileDoc) folderTilesOf(nodeID int64) []string {
	var out []string
	for _, t := range d.folderTiles() {
		if *t.NodeID == nodeID {
			out = append(out, t.ID)
		}
	}
	return out
}

// folderTiles: 노드를 보여 주는 폴더 타일 전부(루트 목록 제외, 루트에서 너비 우선 순)
func (d *tileDoc) folderTiles() []tileItem {
	var out []tileItem
	queue := []string{d.RootID}
	for len(queue) > 0 {
		t := d.Tiles[queue[0]]
		queue = queue[1:]
		if t.Type == tileFolder && t.NodeID != nil {
			out = append(out, t)
		}
		queue = append(queue, t.Kids...)
	}
	return out
}

// terminalOf: uid의 process 타일(능력-3(process 담음))
func (d *tileDoc) terminalOf(uid string) (tileItem, bool) {
	for _, t := range d.Tiles {
		if t.holdsProc() && t.UID == uid {
			return t, true
		}
	}
	return tileItem{}, false
}

// terminalUIDs: 문서 안 process uid 전부
func (d *tileDoc) terminalUIDs() []string {
	var uids []string
	for _, t := range d.Tiles {
		if uid := t.procUID(); uid != "" {
			uids = append(uids, uid)
		}
	}
	return uids
}
