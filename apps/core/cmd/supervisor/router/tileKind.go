package router

import (
	"errors"

	"github.com/google/uuid"
)

// 타일 종류 명세 — 종류를 비교하지 말고 능력으로 묻는다. 새 종류 = tileKinds 줄 하나 + 그 종류만의 check (REF-tile-kinds.md)

const (
	tileFolder   = "folder"
	tileTerminal = "terminal"
)

type tileKind struct {
	canRoot   bool                   // 능력-1(루트 가능): 보드 루트·새 보드로 분리
	canParent bool                   // 능력-2(자식 가능): kids를 가짐
	holdsProc bool                   // 능력-3(process 담음): uid = process uid, adopted·minimized·shortcutsOff. 이런 종류를 더하면 SQL의 `type == "terminal"`(processes.sql·tileTrees.sql 3곳)도 같이
	check     func(t tileItem) error // 종류만의 필드 검사(능력 검사는 validate가 함)
}

var tileKinds = map[string]tileKind{
	tileFolder: {canRoot: true, canParent: true},
	tileTerminal: {holdsProc: true, check: func(t tileItem) error {
		if t.UID == "" || t.NodeID == nil {
			return errors.New("터미널 타일에 uid·nodeId가 없습니다")
		}
		return nil
	}},
}

var defaultTileSize = tileSize{W: 0.5, H: 0.5}

func (t tileItem) kind() tileKind { return tileKinds[t.Type] }

func (t tileItem) canRoot() bool   { return t.kind().canRoot }
func (t tileItem) canParent() bool { return t.kind().canParent }
func (t tileItem) holdsProc() bool { return t.kind().holdsProc }

// procUID: process를 담은 타일의 uid(아니면 "")
func (t tileItem) procUID() string {
	if t.holdsProc() {
		return t.UID
	}
	return ""
}

// newTerminalTile: 서버만 터미널 타일을 만든다(WS5(터미널 타일 추가 주체)) — parent·kids 연결은 호출부
func newTerminalTile(parent *string, nodeID int64, uid string, size tileSize) tileItem {
	return tileItem{ID: uuid.NewString(), Type: tileTerminal, Parent: parent, Kids: []string{}, Size: size, NodeID: &nodeID, UID: uid}
}

func newFolderTile(parent *string, nodeID *int64, uid string) tileItem {
	return tileItem{ID: uuid.NewString(), Type: tileFolder, Parent: parent, Kids: []string{}, Size: defaultTileSize, NodeID: nodeID, UID: uid}
}
