package router

import (
	"encoding/json"
	"slices"
	"testing"
)

// 정리-3(서버 확인): 타일 종류 정리 전후 tileDoc 동작이 같은지 — REF-tile-kinds.md

func i64(v int64) *int64 { return new(v) }

func half() tileSize { return tileSize{W: 0.5, H: 0.5} }

// sampleDoc: root(루트 목록) -> f1(노드 10) -> [t1(터미널), f2(노드 20) -> [t2]], root -> t3
func sampleDoc() tileDoc {
	root, f1, f2 := "root", "f1", "f2"
	return tileDoc{RootID: root, Tiles: map[string]tileItem{
		"root": {ID: root, Type: "folder", Kids: []string{"f1", "t3"}, Size: half()},
		"f1":   {ID: f1, Type: "folder", Parent: &root, Kids: []string{"t1", "f2"}, Size: half(), NodeID: i64(10), UID: "fu1"},
		"f2":   {ID: f2, Type: "folder", Parent: &f1, Kids: []string{"t2"}, Size: half(), NodeID: i64(20), UID: "fu2"},
		"t1":   {ID: "t1", Type: "terminal", Parent: &f1, Kids: []string{}, Size: half(), NodeID: i64(11), UID: "p1"},
		"t2":   {ID: "t2", Type: "terminal", Parent: &f2, Kids: []string{}, Size: half(), NodeID: i64(21), UID: "p2", Minimized: true},
		"t3":   {ID: "t3", Type: "terminal", Parent: &root, Kids: []string{}, Size: half(), NodeID: i64(31), UID: "p3", Adopted: true, ShortcutsOff: true},
	}}
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
			x := d.Tiles["t3"]
			x.Parent = nil
			d.Tiles["t3"] = x
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
		"folder with adopted": func(d *tileDoc) {
			x := d.Tiles["f2"]
			x.Adopted = true
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
		"terminal without nodeId": func(d *tileDoc) {
			x := d.Tiles["t1"]
			x.NodeID = nil
			d.Tiles["t1"] = x
		},
		"terminal with kids": func(d *tileDoc) {
			x := d.Tiles["t1"]
			x.Kids = []string{"t2"}
			d.Tiles["t1"] = x
			y := d.Tiles["t2"]
			p := "t1"
			y.Parent = &p
			d.Tiles["t2"] = y
			f2 := d.Tiles["f2"]
			f2.Kids = []string{}
			d.Tiles["f2"] = f2
		},
		"non-root without parent": func(d *tileDoc) {
			x := d.Tiles["t1"]
			x.Parent = nil
			d.Tiles["t1"] = x
		},
		"parent not in kids": func(d *tileDoc) {
			f1 := d.Tiles["f1"]
			f1.Kids = []string{"f2"}
			d.Tiles["f1"] = f1
		},
		"listed twice": func(d *tileDoc) {
			f1 := d.Tiles["f1"]
			f1.Kids = []string{"t1", "f2", "t1"}
			d.Tiles["f1"] = f1
		},
		"kid missing": func(d *tileDoc) {
			f1 := d.Tiles["f1"]
			f1.Kids = append(slices.Clone(f1.Kids), "ghost")
			d.Tiles["f1"] = f1
		},
		"cycle unreachable": func(d *tileDoc) {
			a, b := "a", "b"
			d.Tiles["a"] = tileItem{ID: a, Type: "folder", Parent: &b, Kids: []string{"b"}, Size: half()}
			d.Tiles["b"] = tileItem{ID: b, Type: "folder", Parent: &a, Kids: []string{"a"}, Size: half()}
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

// 부모가 터미널인 타일: 폴더가 아닌 부모 아래 붙으면 거부
func TestValidateParentMustParent(t *testing.T) {
	d := sampleDoc()
	t1 := "t1"
	f2 := d.Tiles["f2"]
	f2.Kids = []string{}
	d.Tiles["f2"] = f2
	x := d.Tiles["t2"]
	x.Parent = &t1
	d.Tiles["t2"] = x
	y := d.Tiles["t1"]
	y.Kids = []string{"t2"}
	d.Tiles["t1"] = y
	if err := d.validate(); err == nil {
		t.Fatal("validate passed, want error")
	}
}

func TestMarshalIssuesFolderFields(t *testing.T) {
	d := sampleDoc()
	f2 := d.Tiles["f2"]
	f2.UID, f2.Kids = "", nil
	d.Tiles["f2"] = f2
	b, err := d.marshal()
	if err != nil {
		t.Fatal(err)
	}
	var out tileDoc
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	for id, x := range out.Tiles {
		if x.Kids == nil {
			t.Errorf("%s kids nil", id)
		}
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

func TestClosedTerminalsAndRemoveLeaf(t *testing.T) {
	stored := sampleDoc()
	next := sampleDoc()
	next.removeLeaf("t2")
	next.removeLeaf("t3")
	if _, ok := next.Tiles["t2"]; ok {
		t.Fatal("t2 still in tiles")
	}
	if k := next.Tiles["f2"].Kids; len(k) != 0 {
		t.Errorf("f2 kids = %v", k)
	}
	if k := next.Tiles["root"].Kids; !slices.Equal(k, []string{"f1"}) {
		t.Errorf("root kids = %v", k)
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

	sub := d.detachFolder("f1")
	if sub.RootID != "f1" || sub.Tiles["f1"].Parent != nil {
		t.Fatalf("sub root = %q parent=%v", sub.RootID, sub.Tiles["f1"].Parent)
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
	d.attachUnderRoot(sub)
	if err := d.validate(); err != nil {
		t.Errorf("reattached invalid: %v", err)
	}
}
