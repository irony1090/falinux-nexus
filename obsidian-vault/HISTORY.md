# HISTORY

> CURRENT/MEMORY가 비대해지는 것을 막기 위한 상세 기록 저장소.
> 작업 상태가 바뀌면 기존 CURRENT 내용을 여기로 정리해 내린다.

## 과거 작업 기록

### 2026-06-17 - Nexus 프로젝트 분리 & vault 신규 구성
- 이전 프로젝트(PortBridge / test-jig)의 방향성이 "테스트 장비 데스크탑 UI"에서
  "2-tier 에이전트 관리 플랫폼"으로 전환됨에 따라 신규 프로젝트 **Nexus**로 분리
- 새 경로: `/home/jh-bae/irony/nexus`
- vault 구성: PortBridge 전용 자산(데스크탑 메타포 UI, 아이콘/그룹 계층, Dashboard,
  제안서/와이어프레임)은 폐기. agent↔server 통신 아키텍처·PTY 실행 추상화·구독 모델·
  SQLite/sqlc 학습만 MEMORY로 계승
- PortBridge의 상세 HISTORY는 가져오지 않음 (test-jig 폴더에 원본 유지)
