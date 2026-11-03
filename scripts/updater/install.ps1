# updater 설치(Windows) — 관리자 PowerShell: powershell -ExecutionPolicy Bypass -File install.ps1 <updater.exe> <updater.json> [-Dir C:\Nexus\updater]
# 다시 실행 = exe 교체 + 재시작, updater.json은 이미 있으면 그대로(Linux install.sh와 같은 계약). REF-update-p4.md 단위-4(설치)
# 단계 스크립트가 Start-Process로 띄운 프로세스는 단계가 정상으로 끝나면 남는다(updater 작업을 재시작해도 남음, VM 확인 2026-11-03). 부팅 뒤에도 켜져야 하는 것은 자기 예약 작업(Start-ScheduledTask)으로 켤 것
param(
    [Parameter(Mandatory = $true, Position = 0)][string]$Exe,
    [Parameter(Mandatory = $true, Position = 1)][string]$Config,
    [string]$Dir = 'C:\Nexus\updater',
    [string]$Task = 'NexusUpdater'
)
$ErrorActionPreference = 'Stop'

$me = [Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()
if (-not $me.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) { throw '관리자 PowerShell에서 실행하세요' }
$Exe = (Resolve-Path $Exe).Path
$Config = (Resolve-Path $Config).Path

# 보완-2(폴더 권한): 상속을 끊고 Administrators(S-1-5-32-544)·SYSTEM(S-1-5-18)만 — 이름은 언어마다 달라 SID로
New-Item -ItemType Directory -Force -Path $Dir | Out-Null
& icacls $Dir /inheritance:r /grant:r '*S-1-5-32-544:(OI)(CI)F' '*S-1-5-18:(OI)(CI)F' | Out-Null
if ($LASTEXITCODE -ne 0) { throw "icacls 실패($LASTEXITCODE)" }

# 켜져 있으면 끄고 프로세스가 끝날 때까지 기다림(실행 중 exe는 덮어쓸 수 없음)
$exePath = Join-Path $Dir 'updater.exe'
if (Get-ScheduledTask -TaskName $Task -ErrorAction SilentlyContinue) { Stop-ScheduledTask -TaskName $Task }
for ($i = 0; $i -lt 20; $i++) {
    $p = Get-Process updater -ErrorAction SilentlyContinue | Where-Object { $_.Path -eq $exePath }
    if (-not $p) { break }
    if ($i -eq 10) { $p | Stop-Process -Force }
    Start-Sleep -Milliseconds 500
}

Copy-Item $Exe (Join-Path $Dir 'updater.new') -Force
Move-Item (Join-Path $Dir 'updater.new') $exePath -Force
$cfgPath = Join-Path $Dir 'updater.json'
if (-not (Test-Path $cfgPath)) { Copy-Item $Config $cfgPath } else { Write-Host "updater.json 있음 — 그대로 둠" }

# 예약 작업: 부팅 때 SYSTEM으로, 시간 제한 없음(기본 3일이면 꺼짐), 멈추면 1분 뒤 다시
$action = New-ScheduledTaskAction -Execute $exePath -Argument '-log updater.log' -WorkingDirectory $Dir
$trigger = New-ScheduledTaskTrigger -AtStartup
$principal = New-ScheduledTaskPrincipal -UserId 'S-1-5-18' -LogonType ServiceAccount -RunLevel Highest
$settings = New-ScheduledTaskSettingsSet -ExecutionTimeLimit ([TimeSpan]::Zero) -RestartCount 999 -RestartInterval (New-TimeSpan -Minutes 1) `
    -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -StartWhenAvailable -MultipleInstances IgnoreNew
Register-ScheduledTask -TaskName $Task -Action $action -Trigger $trigger -Principal $principal -Settings $settings -Force | Out-Null
Start-ScheduledTask -TaskName $Task
Start-Sleep -Seconds 3

$info = Get-ScheduledTaskInfo -TaskName $Task
Write-Host "작업 $Task 상태 = $((Get-ScheduledTask -TaskName $Task).State), 마지막 결과 = $($info.LastTaskResult)"
$log = Join-Path $Dir 'updater.log'
if (Test-Path $log) { Get-Content $log -Tail 5 -Encoding UTF8 }
