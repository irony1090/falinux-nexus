package router

import (
	"encoding/json"
	"slices"
	"testing"
)

// 정리-3(서버 확인): 타일 종류 정리 전후 tileDoc 동작이 같은지 — REF-tile-kinds.md

func i64(v int64) *int64 { return new(v) }

func half() tileSize { return tileSize{W: 0.5, H: 0.5} }

// sampleDoc: 연 곳 root(루트 목록) -> [f1(노드 10) -> [t1(터미널), f2(노드 20) -> [t2]], t3], 화면 순서 = order
func sampleDoc() tileDoc {
	root, f1, f2 := "root", "f1", "f2"
	return tileDoc{RootID: root, Order: []string{"root", "t3", "f1", "t1", "f2", "t2"}, Tiles: map[string]tileItem{
		"root": {ID: root, Type: "folder", Size: half()},
		"f1":   {ID: f1, Type: "folder", Parent: &root, Size: half(), NodeID: i64(10), UID: "fu1"},
		"f2":   {ID: f2, Type: "folder", Parent: &f1, Size: half(), NodeID: i64(20), UID: "fu2"},
		"t1":   {ID: "t1", Type: "terminal", Parent: &f1, Size: half(), NodeID: i64(11), UID: "p1"},
		"t2":   {ID: "t2", Type: "terminal", Parent: &f2, Size: half(), NodeID: i64(21), UID: "p2", Minimized: true},
		"t3":   {ID: "t3", Type: "terminal", Parent: &root, Size: half(), NodeID: i64(31), UID: "p3", ShortcutsOff: true},
	}}
}

func setParent(d *tileDoc, id, parent string) {
	x := d.Tiles[id]
	x.Parent = &parent
	d.Tiles[id] = x
}

func TestValidateOK(t *testing.T) {
	d := sampleDoc()
	if err := d.validate(); err != nil {
		t.Fatal(err)
	}
	r := newRootDoc()
	if err := r.validate(); err != nil {
		t.Fatal(err)
	}
}

func TestValidateRejects(t *testing.T) {
	cases := map[string]func(d *tileDoc){
		"root is terminal": func(d *tileDoc) {
			d.RootID = "t3"
			d.Order = []string{"t3", "root", "f1", "t1", "f2", "t2"}
			x := d.Tiles["t3"]
			x.Parent = nil
			d.Tiles["t3"] = x
			setParent(d, "root", "t3")
		},
		"root has parent": func(d *tileDoc) {
			x := d.Tiles["root"]
			p := "f1"
			x.Parent = &p
			d.Tiles["root"] = x
		},
		"missing root": func(d *tileDoc) { d.RootID = "nope" },
		"id differs from key": func(d *tileDoc) {
			x := d.Tiles["t1"]
			x.ID = "other"
			d.Tiles["t1"] = x
		},
		"bad size": func(d *tileDoc) {
			x := d.Tiles["t1"]
			x.Size.W = 0.3
			d.Tiles["t1"] = x
		},
		"unknown type": func(d *tileDoc) {
			x := d.Tiles["t1"]
			x.Type = "viewer"
			d.Tiles["t1"] = x
		},
		"folder with minimized": func(d *tileDoc) {
			x := d.Tiles["f2"]
			x.Minimized = true
			d.Tiles["f2"] = x
		},
		"folder with shortcutsOff": func(d *tileDoc) {
			x := d.Tiles["f2"]
			x.ShortcutsOff = true
			d.Tiles["f2"] = x
		},
		"terminal without uid": func(d *tileDoc) {
			x := d.Tiles["t1"]
			x.UID = ""
			d.Tiles["t1"] = x
		},
		"opened by terminal": func(d *tileDoc) { setParent(d, "t2", "t1") },
		"opened by missing":  func(d *tileDoc) { setParent(d, "t2", "ghost") },
		"non-root without parent": func(d *tileDoc) {
			x := d.Tiles["t1"]
			x.Parent = nil
			d.Tiles["t1"] = x
		},
		"order missing tile": func(d *tileDoc) { d.Order = []string{"root", "t3", "f1", "f2", "t2"} },
		"order listed twice": func(d *tileDoc) { d.Order = []string{"root", "t3", "f1", "t1", "f2", "t1"} },
		"order unknown id":   func(d *tileDoc) { d.Order = []string{"root", "t3", "f1", "t1", "f2", "ghost"} },
		"root not first":     func(d *tileDoc) { d.Order = []string{"t3", "root", "f1", "t1", "f2", "t2"} },
		"order nil":          func(d *tileDoc) { d.Order = nil },
		"parent cycle": func(d *tileDoc) {
			a, b := "a", "b"
			d.Tiles["a"] = tileItem{ID: a, Type: "folder", Parent: &b, Size: half()}
			d.Tiles["b"] = tileItem{ID: b, Type: "folder", Parent: &a, Size: half()}
			d.Order = append(d.Order, "a", "b")
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			d := sampleDoc()
			mutate(&d)
			if err := d.validate(); err == nil {
				t.Fatal("validate passed, want error")
			}
		})
	}
}

// 장비 파일 실행 터미널은 nodeId가 없다 — 폴더-4(실행)
func TestValidateTerminalWithoutNode(t *testing.T) {
	d := sampleDoc()
	x := d.Tiles["t1"]
	x.NodeID = nil
	d.Tiles["t1"] = x
	if err := d.validate(); err != nil {
		t.Fatal(err)
	}
}

// 순서는 연 곳과 무관 — 연 타일보다 앞에 와도 된다(끌어서 옮기기)
func TestValidateOrderFree(t *testing.T) {
	d := sampleDoc()
	d.Order = []string{"root", "t2", "f2", "t1", "t3", "f1"}
	if err := d.validate(); err != nil {
		t.Fatal(err)
	}
}

func TestMarshalIssuesFolderFields(t *testing.T) {
	d := sampleDoc()
	f2 := d.Tiles["f2"]
	f2.UID = ""
	d.Tiles["f2"] = f2
	b, err := d.marshal()
	if err != nil {
		t.Fatal(err)
	}
	var out tileDoc
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(out.Order, d.Order) {
		t.Errorf("order = %v", out.Order)
	}
	for id, x := range out.Tiles {
		switch x.Type {
		case "folder":
			if x.UID == "" || x.ShownAt == 0 {
				t.Errorf("%s folder uid=%q shownAt=%d", id, x.UID, x.ShownAt)
			}
		case "terminal":
			if x.ShownAt != 0 || x.UID != d.Tiles[id].UID {
				t.Errorf("%s terminal changed: %+v", id, x)
			}
		}
	}
	if out.Tiles["f1"].UID != "fu1" {
		t.Errorf("existing folder uid replaced: %q", out.Tiles["f1"].UID)
	}
}

func TestKeepFolderMeta(t *testing.T) {
	stored := sampleDoc()
	for _, id := range []string{"f1", "f2"} {
		x := stored.Tiles[id]
		x.ShownAt, x.Bound = 111, true
		stored.Tiles[id] = x
	}
	next := sampleDoc()
	f1 := next.Tiles["f1"]
	f1.UID, f1.ShownAt, f1.Bound = "forged", 999, false // 프론트 값은 버려짐
	next.Tiles["f1"] = f1
	f2 := next.Tiles["f2"]
	f2.NodeID = i64(99) // 다른 노드로 감 = 새로 연 것
	next.Tiles["f2"] = f2
	t1 := next.Tiles["t1"]
	t1.ShownAt, t1.Bound = 5, true
	next.Tiles["t1"] = t1
	next.Tiles["fnew"] = tileItem{ID: "fnew", Type: "folder", UID: "forged2", ShownAt: 7, Bound: true}

	next.keepFolderMeta(&stored)

	if g := next.Tiles["f1"]; g.UID != "fu1" || g.ShownAt != 111 || !g.Bound {
		t.Errorf("f1 = %+v", g)
	}
	if g := next.Tiles["f2"]; g.UID != "fu2" || g.ShownAt != 0 || g.Bound {
		t.Errorf("f2 = %+v", g)
	}
	if g := next.Tiles["t1"]; g.UID != "p1" || g.ShownAt != 0 || g.Bound {
		t.Errorf("t1 = %+v", g)
	}
	if g := next.Tiles["fnew"]; g.UID != "" || g.ShownAt != 0 || g.Bound {
		t.Errorf("fnew = %+v", g)
	}
}

func TestCheckTerminalUIDs(t *testing.T) {
	stored := sampleDoc()
	next := sampleDoc()
	if err := checkTerminalUIDs(&next, &stored); err != nil {
		t.Fatal(err)
	}
	added := sampleDoc()
	root := "root"
	added.Tiles["t9"] = tileItem{ID: "t9", Type: "terminal", Parent: &root, NodeID: i64(1), UID: "new"}
	if err := checkTerminalUIDs(&added, &stored); err == nil {
		t.Error("new terminal accepted")
	}
	dup := sampleDoc()
	dup.Tiles["t9"] = tileItem{ID: "t9", Type: "terminal", Parent: &root, NodeID: i64(1), UID: "p1"}
	if err := checkTerminalUIDs(&dup, &stored); err == nil {
		t.Error("duplicate terminal accepted")
	}
	// 폴더 uid는 검사 대상 아님
	fol := sampleDoc()
	fol.Tiles["f9"] = tileItem{ID: "f9", Type: "folder", Parent: &root, UID: "p1"}
	if err := checkTerminalUIDs(&fol, &stored); err != nil {
		t.Errorf("folder uid checked: %v", err)
	}
}

func TestClosedTerminalsAndRemove(t *testing.T) {
	stored := sampleDoc()
	next := sampleDoc()
	next.remove("t2")
	next.remove("t3")
	if _, ok := next.Tiles["t2"]; ok {
		t.Fatal("t2 still in tiles")
	}
	if !slices.Equal(next.Order, []string{"root", "f1", "t1", "f2"}) {
		t.Errorf("order = %v", next.Order)
	}
	if err := next.validate(); err != nil {
		t.Fatal(err)
	}
	got := closedTerminals(&next, &stored)
	slices.Sort(got)
	if !slices.Equal(got, []string{"p2", "p3"}) {
		t.Errorf("closed = %v", got)
	}
}

// 폴더를 빼면 그 폴더가 연 타일은 한 단계 위가 연 것으로, 화면 순서는 그 타일만 빠짐
func TestRemoveLiftsOpened(t *testing.T) {
	d := sampleDoc()
	d.remove("f1")
	if p := d.Tiles["t1"].Parent; p == nil || *p != "root" {
		t.Errorf("t1 parent = %v", p)
	}
	if p := d.Tiles["f2"].Parent; p == nil || *p != "root" {
		t.Errorf("f2 parent = %v", p)
	}
	if p := d.Tiles["t2"].Parent; p == nil || *p != "f2" {
		t.Errorf("t2 parent = %v", p)
	}
	if !slices.Equal(d.Order, []string{"root", "t3", "t1", "f2", "t2"}) {
		t.Errorf("order = %v", d.Order)
	}
	if err := d.validate(); err != nil {
		t.Fatal(err)
	}
}

// 새 타일 = 연 곳 바로 뒤(최신이 앞). 연 곳이 없거나 자식 불가면 루트 바로 뒤(연곳-B)
func TestOpenTile(t *testing.T) {
	d := sampleDoc()
	a := d.openTile("f1", newTerminalTile(i64(1), "pa", defaultTileSize))
	b := d.openTile("f1", newTerminalTile(i64(1), "pb", defaultTileSize))
	c := d.openTile("ghost", newTerminalTile(i64(1), "pc", defaultTileSize))
	e := d.openTile("t1", newTerminalTile(i64(1), "pe", defaultTileSize))
	want := []string{"root", e.ID, c.ID, "t3", "f1", b.ID, a.ID, "t1", "f2", "t2"}
	if !slices.Equal(d.Order, want) {
		t.Errorf("order = %v, want %v", d.Order, want)
	}
	if *a.Parent != "f1" || *c.Parent != "root" || *e.Parent != "root" {
		t.Errorf("parents a=%s c=%s e=%s", *a.Parent, *c.Parent, *e.Parent)
	}
	if err := d.validate(); err != nil {
		t.Fatal(err)
	}
}

// 옛 문서(kids·adopted) -> 옛 화면 순서 그대로 — 변환-B(시작 때 한 번)
func TestLegacyOrder(t *testing.T) {
	old := `{"rootId":"root","tiles":{
		"root":{"id":"root","type":"folder","parent":null,"kids":["f1","t3","t4"],"size":{"w":0.5,"h":0.5}},
		"f1":{"id":"f1","type":"folder","parent":"root","kids":["v1","t1","f2"],"size":{"w":0.5,"h":0.5},"nodeId":10},
		"f2":{"id":"f2","type":"folder","parent":"f1","kids":["t2"],"size":{"w":0.5,"h":0.5},"nodeId":20},
		"v1":{"id":"v1","type":"viewer","parent":"f1","kids":[],"size":{"w":0.5,"h":0.5},"nodeId":10,"instance":"m#s","path":"/a.png"},
		"t1":{"id":"t1","type":"terminal","parent":"f1","kids":[],"size":{"w":0.5,"h":0.5},"nodeId":11,"uid":"p1"},
		"t2":{"id":"t2","type":"terminal","parent":"f2","kids":[],"size":{"w":0.5,"h":0.5},"nodeId":21,"uid":"p2"},
		"t3":{"id":"t3","type":"terminal","parent":"root","kids":[],"size":{"w":0.5,"h":0.5},"nodeId":31,"uid":"p3","adopted":true},
		"t4":{"id":"t4","type":"terminal","parent":"root","kids":[],"size":{"w":0.5,"h":0.5},"nodeId":41,"uid":"p4"}}}`
	order, err := legacyOrder([]byte(old))
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"root", "t4", "t3", "f1", "t1", "v1", "f2", "t2"}; !slices.Equal(order, want) {
		t.Fatalf("order = %v, want %v", order, want)
	}
	var doc tileDoc
	if err := json.Unmarshal([]byte(old), &doc); err != nil {
		t.Fatal(err)
	}
	doc.Order = order
	if err := doc.validate(); err != nil {
		t.Fatal(err)
	}
	b, err := doc.marshal()
	if err != nil {
		t.Fatal(err)
	}
	var raw struct {
		Tiles map[string]map[string]any `json:"tiles"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatal(err)
	}
	for id, x := range raw.Tiles {
		if _, ok := x["kids"]; ok {
			t.Errorf("%s kids kept", id)
		}
		if _, ok := x["adopted"]; ok {
			t.Errorf("%s adopted kept", id)
		}
	}
}

func TestFinders(t *testing.T) {
	d := sampleDoc()
	var order []string
	for _, x := range d.folderTiles() {
		order = append(order, x.ID)
	}
	if !slices.Equal(order, []string{"f1", "f2"}) {
		t.Errorf("folderTiles = %v", order)
	}
	if id, ok := d.folderTileOf(20); !ok || id != "f2" {
		t.Errorf("folderTileOf(20) = %q %v", id, ok)
	}
	if _, ok := d.folderTileOf(11); ok {
		t.Error("folderTileOf(11) found a terminal's node")
	}
	if !d.hasFolderUID("fu2") || d.hasFolderUID("p1") {
		t.Error("hasFolderUID wrong")
	}
	if x, ok := d.terminalOf("p2"); !ok || x.ID != "t2" {
		t.Errorf("terminalOf(p2) = %+v %v", x, ok)
	}
	if _, ok := d.terminalOf("fu1"); ok {
		t.Error("terminalOf found a folder uid")
	}
	uids := d.terminalUIDs()
	slices.Sort(uids)
	if !slices.Equal(uids, []string{"p1", "p2", "p3"}) {
		t.Errorf("terminalUIDs = %v", uids)
	}
}

func mustPanic(t *testing.T, name string, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Errorf("%s: no panic", name)
		}
	}()
	f()
}

func TestDetachFolder(t *testing.T) {
	d := sampleDoc()
	mustPanic(t, "terminal", func() { d.detachFolder("t1") })
	mustPanic(t, "root", func() { d.detachFolder("root") })
	mustPanic(t, "missing", func() { d.detachFolder("nope") })

	// 연 타일이 연 곳보다 앞에 끌려가 있어도 새 보드는 떼어낸 타일이 맨 앞
	d.Order = []string{"root", "t2", "t3", "f1", "t1", "f2"}
	sub := d.detachFolder("f1")
	if sub.RootID != "f1" || sub.Tiles["f1"].Parent != nil {
		t.Fatalf("sub root = %q parent=%v", sub.RootID, sub.Tiles["f1"].Parent)
	}
	if !slices.Equal(sub.Order, []string{"f1", "t2", "t1", "f2"}) {
		t.Errorf("sub order = %v", sub.Order)
	}
	if !slices.Equal(d.Order, []string{"root", "t3"}) {
		t.Errorf("rest order = %v", d.Order)
	}
	ids := make([]string, 0, len(sub.Tiles))
	for id := range sub.Tiles {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	if !slices.Equal(ids, []string{"f1", "f2", "t1", "t2"}) {
		t.Errorf("sub tiles = %v", ids)
	}
	if err := sub.validate(); err != nil {
		t.Errorf("sub invalid: %v", err)
	}
	if err := d.validate(); err != nil {
		t.Errorf("rest invalid: %v", err)
	}
}

// 보드-1(보드 간 이동 API): 다른 보드로 = 연 타일까지 묶음째, after 뒤(""= 루트 뒤, 없는 타일 = 맨 뒤), 맨 위 연 곳 = 도착 루트
func TestMoveOpenedBetweenBoards(t *testing.T) {
	cases := map[string]struct {
		after string
		want  []string
	}{
		"after root": {"", []string{"droot", "t1", "f1", "f2", "t2", "t3"}},
		"after t3":   {"t3", []string{"droot", "t3", "t1", "f1", "f2", "t2"}},
		"stale":      {"gone", []string{"droot", "t3", "t1", "f1", "f2", "t2"}},
	}
	for name, tc := range cases {
		src := sampleDoc()
		src.Order = []string{"root", "t1", "t3", "f1", "f2", "t2"} // 연 터미널 t1이 폴더 f1보다 앞
		dst := tileDoc{RootID: "droot", Order: []string{"droot"}, Tiles: map[string]tileItem{}}
		dst.Tiles["droot"] = tileItem{ID: "droot", Type: "folder", Size: half()}
		moved := src.takeOpened("f1")
		if !slices.Equal(src.Order, []string{"root", "t3"}) {
			t.Errorf("%s: src order = %v", name, src.Order)
		}
		if err := src.validate(); err != nil {
			t.Errorf("%s: src invalid: %v", name, err)
		}
		// 도착 보드에 t3 하나를 미리 둔다(after 확인용)
		t3 := src.Tiles["t3"]
		setParentTo := "droot"
		t3.Parent = &setParentTo
		dst.Tiles["t3"] = t3
		dst.Order = append(dst.Order, "t3")
		dst.putOpened(tc.after, moved)
		if !slices.Equal(dst.Order, tc.want) {
			t.Errorf("%s: dst order = %v", name, dst.Order)
		}
		if p := dst.Tiles["f1"].Parent; p == nil || *p != "droot" {
			t.Errorf("%s: f1 parent = %v", name, p)
		}
		if p := dst.Tiles["t1"].Parent; p == nil || *p != "f1" {
			t.Errorf("%s: t1 parent = %v", name, p)
		}
		if err := dst.validate(); err != nil {
			t.Errorf("%s: dst invalid: %v", name, err)
		}
	}
}

// 폴더-2(장비 타일): f1 아래 장비 타일 d1 -> [t4] — REF-worker-dir-list.md
func deviceDoc() tileDoc {
	d := sampleDoc()
	f1, d1 := "f1", "d1"
	d.Tiles[d1] = tileItem{ID: d1, Type: tileDevice, Parent: &f1, Size: half(), NodeID: i64(10), Instance: "dev#ab", Path: "/usr", Back: []string{"", "/"}}
	d.Tiles["t4"] = tileItem{ID: "t4", Type: "terminal", Parent: &d1, Size: half(), NodeID: i64(41), UID: "p4"}
	d.Order = append(d.Order, d1, "t4")
	return d
}

func TestValidateDevice(t *testing.T) {
	ok := deviceDoc()
	if err := ok.validate(); err != nil {
		t.Fatal(err)
	}
	cases := map[string]func(d *tileDoc){
		"device without nodeId":   func(d *tileDoc) { x := d.Tiles["d1"]; x.NodeID = nil; d.Tiles["d1"] = x },
		"device without instance": func(d *tileDoc) { x := d.Tiles["d1"]; x.Instance = "dev"; d.Tiles["d1"] = x },
		"device with uid":         func(d *tileDoc) { x := d.Tiles["d1"]; x.UID = "u"; d.Tiles["d1"] = x },
		"folder with path":        func(d *tileDoc) { x := d.Tiles["f2"]; x.Path = "/"; d.Tiles["f2"] = x },
		"terminal with instance":  func(d *tileDoc) { x := d.Tiles["t1"]; x.Instance = "dev#ab"; d.Tiles["t1"] = x },
		"history too long": func(d *tileDoc) {
			x := d.Tiles["d1"]
			x.Back, x.Fwd = make([]string, 15), make([]string, 6)
			d.Tiles["d1"] = x
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			d := deviceDoc()
			mutate(&d)
			if err := d.validate(); err == nil {
				t.Fatal("validate passed, want error")
			}
		})
	}
}

// 장비 타일은 폴더 전용(uid 발급·folderTiles)에 안 잡힌다
func TestDeviceNotFolder(t *testing.T) {
	d := deviceDoc()
	b, err := d.marshal()
	if err != nil {
		t.Fatal(err)
	}
	var back tileDoc
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if x := back.Tiles["d1"]; x.UID != "" || x.Instance != "dev#ab" || x.Path != "/usr" || len(x.Back) != 2 {
		t.Fatalf("device tile changed: %+v", x)
	}
	if ids := back.folderTilesOf(10); !slices.Equal(ids, []string{"f1"}) {
		t.Fatalf("folderTilesOf(10) = %v", ids)
	}
}

func TestCheckVideoTiles(t *testing.T) {
	root := "root"
	stored := sampleDoc()
	stored.Tiles["v1"] = tileItem{ID: "v1", Type: "viewer", Parent: &root, NodeID: i64(1), Instance: "m#s", Path: "/a.mp4"}
	same := sampleDoc()
	same.Tiles["v1"] = stored.Tiles["v1"]
	if err := checkVideoTiles(&same, &stored); err != nil {
		t.Fatal(err)
	}
	added := sampleDoc()
	added.Tiles["v2"] = tileItem{ID: "v2", Type: "viewer", Parent: &root, NodeID: i64(1), Instance: "m#s", Path: "/b.MP4"}
	if err := checkVideoTiles(&added, &stored); err == nil {
		t.Error("new video tile accepted")
	}
	// 있던 타일의 경로를 다른 영상으로 바꾸기 = 추가로 봄
	img := sampleDoc()
	img.Tiles["v1"] = tileItem{ID: "v1", Type: "viewer", Parent: &root, NodeID: i64(1), Instance: "m#s", Path: "/c.webm"}
	if err := checkVideoTiles(&img, &stored); err == nil {
		t.Error("path change to other video accepted")
	}
	// 이미지 보기 타일은 프론트가 넣음
	pic := sampleDoc()
	pic.Tiles["v3"] = tileItem{ID: "v3", Type: "viewer", Parent: &root, NodeID: i64(1), Instance: "m#s", Path: "/a.png"}
	if err := checkVideoTiles(&pic, &stored); err != nil {
		t.Errorf("image tile rejected: %v", err)
	}
}

func TestClosedVideos(t *testing.T) {
	root := "root"
	stored := sampleDoc()
	stored.Tiles["v1"] = tileItem{ID: "v1", Type: "viewer", Parent: &root, NodeID: i64(1), Instance: "m#s", Path: "/a.mp4"}
	stored.Tiles["i1"] = tileItem{ID: "i1", Type: "viewer", Parent: &root, NodeID: i64(1), Instance: "m#s", Path: "/a.png"}
	next := sampleDoc()
	got := closedVideos(&next, &stored)
	if len(got) != 1 || got[0].ID != "v1" {
		t.Errorf("closed = %v", got)
	}
	next.Tiles["v1"] = stored.Tiles["v1"]
	if got := closedVideos(&next, &stored); len(got) != 0 {
		t.Errorf("kept video reported closed: %v", got)
	}
}
