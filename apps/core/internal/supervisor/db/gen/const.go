package superdb

const (
	NodeKindFolder = "FOLDER"
	NodeKindScript = "SCRIPT"
)

// 공유 단계(node_permission/process_permission 반환값). 비교는 항상 >= → REF-process-share-db.md
const (
	PermBlock   int16 = 0 // 차단 = 접근 없음
	PermView    int16 = 10
	PermOperate int16 = 20
	PermEdit    int16 = 30
	PermOwner   int16 = 40 // 함수 반환값에만 있음(공유 행에 저장 안 함)
)
