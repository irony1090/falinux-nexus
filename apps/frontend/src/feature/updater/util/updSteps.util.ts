import type { Updater } from '../api/updater.api'

// steps.json 빈 틀 — steps.json 타일과 Files 타일 갈아치우기 토글(steps.json이 없을 때) 공용
export const stepsTemplate = (ext: 'sh' | 'ps1') => `{
  "uses": [],
  "deploy": [
    { "id": "stop", "script": "00-stop.${ext}" },
    { "id": "backup", "builtin": "backup" },
    { "id": "copy", "builtin": "copy" },
    { "id": "start", "script": "03-start.${ext}" }
  ],
  "rollback": [
    { "id": "stop", "script": "00-stop.${ext}" },
    { "id": "restore", "builtin": "restore" },
    { "id": "start", "script": "03-start.${ext}" }
  ]
}
`

// 틀 확장자 = 같은 메인키 설치본 os. windows만 = .ps1, 섞임·없음 = .sh
export const osSetOf = (instances: Pick<Updater, 'os'>[] = []) => new Set(instances.map(i => i.os).filter(Boolean))
export const stepsExt = (os: Set<string>): 'sh' | 'ps1' => os.size === 1 && os.has('windows') ? 'ps1' : 'sh'
