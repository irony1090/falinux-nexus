package router

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

// 타일 종류 명세 — 종류를 비교하지 말고 능력으로 묻는다. 새 종류 = tileKinds 줄 하나 + 그 종류만의 check (REF-tile-kinds.md)

const (
	tileFolder   = "folder"
	tileTerminal = "terminal"
	tileDevice   = "device" // 장비 타일 — worker 디렉토리(폴더-2(장비 타일), REF-worker-dir-list.md)
	tileViewer   = "viewer" // 보기 타일 — 장비 파일 하나(파트(보기 타일), REF-file-view.md)
)

// tileHistMax: 이동 기록 뒤로·앞으로 합친 수 — Q16(기록 보관 위치)
const tileHistMax = 20

type tileKind struct {
	canRoot   bool                   // 능력-1(루트 가능): 보드 루트·새 보드로 분리
	canParent bool                   // 능력-2(자식 가능): 다른 타일을 연다(연 곳 = parent가 될 수 있음)
	holdsProc bool                   // 능력-3(process 담음): uid = process uid, minimized·shortcutsOff. 이런 종류를 더하면 SQL의 `type == "terminal"`(processes.sql·tileTrees.sql 3곳)도 같이
	check     func(t tileItem) error // 종류만의 필드 검사(능력 검사는 validate가 함)
}

var tileKinds = map[string]tileKind{
	tileFolder: {canRoot: true, canParent: true, check: noDeviceFields},
	// nodeId = 스크립트 노드, 장비 파일 실행이면 null(폴더-4(실행))
	tileTerminal: {holdsProc: true, check: func(t tileItem) error {
		if t.UID == "" {
			return errors.New("터미널 타일에 uid가 없습니다")
		}
		return noDeviceFields(t)
	}},
	// nodeId = 연 폴더(Q15(장비 디렉토리 id)), instance = main#sub, path = 지금 경로("" = 최상위)
	tileDevice: {canRoot: true, canParent: true, check: func(t tileItem) error {
		if t.NodeID == nil || !strings.Contains(t.Instance, "#") {
			return errors.New("장비 타일에 nodeId·instance가 없습니다")
		}
		if t.UID != "" {
			return errors.New("장비 타일에는 uid가 없습니다")
		}
		return nil
	}},
	// 장비 타일과 같은 필드, path = 파일 경로. 이미지·영상은 확장자로 나눔(프론트)
	tileViewer: {check: func(t tileItem) error {
		if t.NodeID == nil || !strings.Contains(t.Instance, "#") || t.Path == "" {
			return errors.New("보기 타일에 nodeId·instance·path가 없습니다")
		}
		if t.UID != "" {
			return errors.New("보기 타일에는 uid가 없습니다")
		}
		return nil
	}},
}

func noDeviceFields(t tileItem) error {
	if t.Instance != "" || t.Path != "" {
		return errors.New("장비 타일이 아닌데 instance·path가 있습니다")
	}
	return nil
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

// 새 타일 — 자리(parent·order)는 tileDoc.openTile이 정한다
// newTerminalTile: 서버만 터미널 타일을 만든다(WS5(터미널 타일 추가 주체))
func newTerminalTile(nodeID *int64, uid string, size tileSize) tileItem {
	return tileItem{ID: uuid.NewString(), Type: tileTerminal, Size: size, NodeID: nodeID, UID: uid}
}

// newViewerTile: 영상은 서버만 만든다(openVideoTile), 이미지는 프론트
func newViewerTile(nodeID int64, instance, path string, size tileSize) tileItem {
	return tileItem{ID: uuid.NewString(), Type: tileViewer, Size: size, NodeID: &nodeID, Instance: instance, Path: path}
}

func newFolderTile(nodeID *int64, uid string) tileItem {
	return tileItem{ID: uuid.NewString(), Type: tileFolder, Size: defaultTileSize, NodeID: nodeID, UID: uid}
}
