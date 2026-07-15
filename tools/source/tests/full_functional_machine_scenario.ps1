param(
    [string]$RepoRoot = ".",
    [string]$EngineExePath,
    [string]$DevCliPath,
    [string]$WorkerExePath,
    [string]$FixtureFile = "",
    [string]$DynamicInterfacesFile = "",
    [string]$WorldTimeSettingsPath = "",
    [string]$WorldStatePath = "",
    [string]$StoryStatePath = "",
    [string]$StoryHistoryPath = "",
    [string]$TickPolicyPath = "",
    [string]$ApiKey = "dev-key",
    [string]$RuntimeTaskToken = "dev-task-token",
    [string]$CallbackToken = "dev-callback-token",
    [int]$EnginePort = 18085,
    [string]$OutFile = ""
)

$ErrorActionPreference = "Stop"
$ScriptDir = if ($PSScriptRoot) { $PSScriptRoot } else { Split-Path -Parent $MyInvocation.MyCommand.Path }

if ([string]::IsNullOrWhiteSpace($FixtureFile)) { $FixtureFile = Join-Path $ScriptDir "machine_scenario_fixture.json" }
if ([string]::IsNullOrWhiteSpace($DynamicInterfacesFile)) { $DynamicInterfacesFile = Join-Path $ScriptDir "runtime_task_dynamic_interfaces.json" }
if ([string]::IsNullOrWhiteSpace($WorldTimeSettingsPath)) { $WorldTimeSettingsPath = Join-Path $ScriptDir "world_time_settings_flexible.json" }
if ([string]::IsNullOrWhiteSpace($WorldStatePath)) { $WorldStatePath = Join-Path $ScriptDir "state_world_state.json" }
if ([string]::IsNullOrWhiteSpace($StoryStatePath)) { $StoryStatePath = Join-Path $ScriptDir "state_story_state.json" }
if ([string]::IsNullOrWhiteSpace($StoryHistoryPath)) { $StoryHistoryPath = Join-Path $ScriptDir "state_story_history.json" }
if ([string]::IsNullOrWhiteSpace($TickPolicyPath)) { $TickPolicyPath = Join-Path $ScriptDir "state_tick_policy.json" }
if ([string]::IsNullOrWhiteSpace($OutFile)) { $OutFile = Join-Path $ScriptDir "full_functional_machine_scenario_result.json" }

function Assert-True {
    param(
        [Parameter(Mandatory = $true)][bool]$Condition,
        [Parameter(Mandatory = $true)][string]$Message
    )
    if (-not $Condition) { throw $Message }
}

function Assert-Equal {
    param(
        [Parameter(Mandatory = $true)]$Actual,
        [Parameter(Mandatory = $true)]$Expected,
        [Parameter(Mandatory = $true)][string]$Message
    )
    if ($Actual -ne $Expected) { throw "$Message. expected=[$Expected] actual=[$Actual]" }
}

function Invoke-EngineJson {
    param(
        [Parameter(Mandatory = $true)][string]$Method,
        [Parameter(Mandatory = $true)][string]$Path,
        [object]$Body = $null,
        [hashtable]$ExtraHeaders = @{}
    )
    $headers = @{ "X-API-Key" = $ApiKey; "Content-Type" = "application/json" }
    foreach ($key in $ExtraHeaders.Keys) { $headers[$key] = $ExtraHeaders[$key] }
    $uri = "http://127.0.0.1:$EnginePort$Path"
    if ($null -eq $Body) { return Invoke-RestMethod -Method $Method -Uri $uri -Headers $headers }
    $json = $Body | ConvertTo-Json -Depth 20 -Compress
    return Invoke-RestMethod -Method $Method -Uri $uri -Headers $headers -Body $json
}

function Invoke-RuntimeTaskJson {
    param(
        [Parameter(Mandatory = $true)][string]$Method,
        [Parameter(Mandatory = $true)][string]$Path,
        [object]$Body = $null
    )
    return Invoke-EngineJson -Method $Method -Path $Path -Body $Body -ExtraHeaders @{ "X-Runtime-Task-Token" = $RuntimeTaskToken }
}

function Run-DevCli {
    param([Parameter(Mandatory = $true)][string[]]$Args)
    $allArgs = @("--server", "http://127.0.0.1:$EnginePort", "--key", $ApiKey) + $Args
    $output = & $DevCliPath @allArgs
    if ($LASTEXITCODE -ne 0) { throw "DevCli failed: $($allArgs -join ' ')" }
    if ([string]::IsNullOrWhiteSpace(($output -join "`n"))) { return $null }
    return ($output -join "`n") | ConvertFrom-Json
}

function Wait-EngineHealthy {
    param([int]$TimeoutSeconds = 30)
    $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
    while ((Get-Date) -lt $deadline) {
        try {
            Invoke-RestMethod -Method GET -Uri "http://127.0.0.1:$EnginePort/health" | Out-Null
            return
        } catch {
            Start-Sleep -Milliseconds 500
        }
    }
    throw "engine did not become healthy within ${TimeoutSeconds}s"
}

function Start-HiddenProcess {
    param(
        [Parameter(Mandatory = $true)][string]$FilePath,
        [Parameter(Mandatory = $true)][string[]]$ArgumentList,
        [Parameter(Mandatory = $true)][string]$StdOutPath,
        [Parameter(Mandatory = $true)][string]$StdErrPath,
        [string]$WorkingDirectory = $RepoRoot
    )
    return Start-Process -FilePath $FilePath -ArgumentList $ArgumentList -WorkingDirectory $WorkingDirectory -RedirectStandardOutput $StdOutPath -RedirectStandardError $StdErrPath -WindowStyle Hidden -PassThru
}

function Stop-ManagedProcess {
    param($Process)
    if ($null -eq $Process) { return }
    try {
        if (-not $Process.HasExited) { Stop-Process -Id $Process.Id -Force }
    } catch {
    }
}

function Add-CheckResult {
    param(
        [Parameter(Mandatory = $true)][AllowEmptyCollection()][System.Collections.Generic.List[object]]$Results,
        [Parameter(Mandatory = $true)][string]$Area,
        [Parameter(Mandatory = $true)][string]$Operation,
        [Parameter(Mandatory = $true)][string]$Status,
        [Parameter(Mandatory = $true)][string]$Evidence
    )
    $Results.Add([pscustomobject]@{ area = $Area; operation = $Operation; status = $Status; evidence = $Evidence }) | Out-Null
}

function Get-TaskByCallbackId {
    param(
        [Parameter(Mandatory = $true)][string]$WorldId,
        [Parameter(Mandatory = $true)][string]$CallbackId
    )
    $resp = Invoke-RuntimeTaskJson -Method GET -Path "/api/v1/runtime/tasks?world_id=$WorldId&limit=50"
    return @($resp.tasks) | Where-Object { $_.callback_id -eq $CallbackId } | Select-Object -First 1
}

function Wait-TaskStatus {
    param(
        [Parameter(Mandatory = $true)][string]$WorldId,
        [Parameter(Mandatory = $true)][string]$CallbackId,
        [Parameter(Mandatory = $true)][string]$ExpectedStatus,
        [int]$TimeoutSeconds = 20
    )
    $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
    while ((Get-Date) -lt $deadline) {
        $task = Get-TaskByCallbackId -WorldId $WorldId -CallbackId $CallbackId
        if ($null -ne $task -and $task.status -eq $ExpectedStatus) { return $task }
        Start-Sleep -Milliseconds 500
    }
    $last = Get-TaskByCallbackId -WorldId $WorldId -CallbackId $CallbackId
    throw "task for callback_id=$CallbackId did not reach status=$ExpectedStatus; last=$($last | ConvertTo-Json -Depth 8 -Compress)"
}

if (-not (Test-Path $EngineExePath)) { throw "EngineExePath not found: $EngineExePath" }
if (-not (Test-Path $DevCliPath)) { throw "DevCliPath not found: $DevCliPath" }
if (-not (Test-Path $WorkerExePath)) { throw "WorkerExePath not found: $WorkerExePath" }
foreach ($path in @($FixtureFile, $DynamicInterfacesFile, $WorldTimeSettingsPath, $WorldStatePath, $StoryStatePath, $StoryHistoryPath, $TickPolicyPath)) {
    if (-not (Test-Path $path)) { throw "Required file not found: $path" }
}

$results = New-Object 'System.Collections.Generic.List[object]'
$tempRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("gae-s9-src-" + (Get-Date -Format "yyyyMMddHHmmss"))
New-Item -ItemType Directory -Force -Path $tempRoot | Out-Null
$dbPath = Join-Path $tempRoot "gameagentengine.db"
$configPath = Join-Path $tempRoot "gameagentengine.conf.yaml"
$engineStdOut = Join-Path $tempRoot "engine.stdout.log"
$engineStdErr = Join-Path $tempRoot "engine.stderr.log"

$configText = @"
server:
  host: "127.0.0.1"
  port: $EnginePort

database:
  driver: "sqlite"
  dsn: "$($dbPath -replace '\\', '\\')"
  migrations_enabled: true

auth:
  api_key: "$ApiKey"
  callback_token: "$CallbackToken"
  runtime_task_token: "$RuntimeTaskToken"
  callback_require_request_id: true

llm:
  provider: "fixture"
  model: "fixture-s9"
  api_key: ""
  base_url: ""
  fixture_file: "$((Resolve-Path $FixtureFile).Path -replace '\\', '\\')"

engine:
  execution_mode: "debug"
  world_lock_enabled: true
  autonomous_scheduler_enabled: false
  runtime_task_governance_interval_seconds: 0

external_interfaces:
  game_client_request_data:
    category: "external_query"
    delivery_mode: "pull"
    primary_transport: "task_pull"
    consumer: "game_client"
    resume_policy: "resume_paused_execution"
"@
Set-Content -LiteralPath $configPath -Value $configText -Encoding ASCII

$engineProc = $null

try {
    $engineProc = Start-HiddenProcess -FilePath $EngineExePath -ArgumentList @("serve", "--config", $configPath) -StdOutPath $engineStdOut -StdErrPath $engineStdErr
    Wait-EngineHealthy
    Add-CheckResult -Results $results -Area "runtime" -Operation "isolated Engine runtime started" -Status "passed" -Evidence "port=$EnginePort config=$configPath"

    $world = Run-DevCli -Args @("node", "create", "--type", "world", "--name", "FullFunctionalMachineScenarioWorld")
    $worldId = $world.id
    $npc = Run-DevCli -Args @("node", "create", "--world", $worldId, "--type", "npc", "--name", "Scene Broker")
    $npcId = $npc.id

    $null = Run-DevCli -Args @("world", "settings", "set", $worldId, "--world-time-settings-file", $WorldTimeSettingsPath, "--pipeline-mode", "full")
    $null = Run-DevCli -Args @("state", "set", $worldId, "world_state", "--file", $WorldStatePath)
    $null = Run-DevCli -Args @("state", "set", $worldId, "story_state", "--file", $StoryStatePath)
    $null = Run-DevCli -Args @("state", "set", $worldId, "story_history", "--file", $StoryHistoryPath)
    $null = Run-DevCli -Args @("state", "set", $worldId, "tick_policy", "--file", $TickPolicyPath)
    $stateWorldTime = Invoke-EngineJson -Method PUT -Path "/api/v1/worlds/$worldId/state-components/world_time_state" -Body @{ current_time_label = "Cycle day 12 hour 8"; total_ticks = 2; last_tick_number = 1; last_tick_type = "manual"; last_advanced_ticks = 2 }
    Assert-Equal $stateWorldTime.state_component.component_type "world_time_state" "machine scenario world_time_state seed mismatch"

    $invoke = Invoke-EngineJson -Method POST -Path "/api/v1/invoke" -Body @{
        world_id = $worldId
        node_id = $npcId
        task_type = "npc_dialogue"
        messages = @(@{ role = "user"; content = "Before answering, query the nearby scene and then respond." })
        context = @{ pipeline_mode = "full"; dynamic_interfaces = [object[]](Get-Content -LiteralPath $DynamicInterfacesFile -Raw | ConvertFrom-Json) }
    }
    $callbackId = $invoke.action_calls[0].callback_id
    $requestId = $invoke.request_id
    Assert-True (-not [string]::IsNullOrWhiteSpace($callbackId)) "machine scenario callback_id missing"
    Assert-True (-not [string]::IsNullOrWhiteSpace($requestId)) "machine scenario request_id missing"
    Add-CheckResult -Results $results -Area "invoke" -Operation "NPC dialogue invoke with request-scoped dynamic interfaces" -Status "passed" -Evidence "request_id=$requestId callback_id=$callbackId"

    $task = Wait-TaskStatus -WorldId $worldId -CallbackId $callbackId -ExpectedStatus "pending"
    Assert-Equal $task.interface_name "game_client_request_data" "machine scenario runtime task interface mismatch"
    Add-CheckResult -Results $results -Area "runtime" -Operation "runtime task created" -Status "passed" -Evidence "task_id=$($task.task_id) interface=$($task.interface_name)"

    $workerOutput = & $WorkerExePath "pull-once" "--engine-base-url" "http://127.0.0.1:$EnginePort" "--runtime-task-token" $RuntimeTaskToken "--callback-token" $CallbackToken "--consumer" "game_client" "--lease-owner" "s9-worker"
    if ($LASTEXITCODE -ne 0) { throw "machine scenario pull-once worker failed" }
    Add-CheckResult -Results $results -Area "worker" -Operation "test worker started" -Status "passed" -Evidence "consumer=game_client lease_owner=s9-worker"

    $completedTask = Wait-TaskStatus -WorldId $worldId -CallbackId $callbackId -ExpectedStatus "succeeded"
    Add-CheckResult -Results $results -Area "worker" -Operation "callback completed by worker" -Status "passed" -Evidence "task_id=$($completedTask.task_id) callback_id=$callbackId"

    $logs = Invoke-EngineJson -Method GET -Path "/api/v1/logs?world_id=$worldId&request_id=$requestId&limit=100"
    $resumeLogs = @($logs | Where-Object { $_.event_name -eq "resume_completed" })
    $pausedLogs = @($logs | Where-Object { $_.event_name -eq "data_request_paused_for_client" })
    Assert-Equal $resumeLogs.Count 1 "machine scenario resume_completed log count mismatch"
    Assert-Equal $pausedLogs.Count 1 "machine scenario paused log count mismatch"
    Add-CheckResult -Results $results -Area "resume" -Operation "paused execution resumed" -Status "passed" -Evidence "resume_logs=$($resumeLogs.Count) paused_logs=$($pausedLogs.Count)"

    $tickLogs = Invoke-EngineJson -Method GET -Path "/api/v1/logs?world_id=$worldId&limit=100"
    $traces = Invoke-EngineJson -Method GET -Path "/debug/traces?world_id=$worldId&limit=20"
    Assert-True ($traces.count -ge 1) "machine scenario traces missing"
    $continuity = Run-DevCli -Args @("debug", "continuity", $worldId, "--request-id", $requestId, "--log-limit", "20", "--trace-limit", "10", "--json")
    $latestTimelinePresent = $null -ne $continuity.latest_timeline
    Assert-True (@($continuity.logs).Count -ge 1) "machine scenario continuity logs missing"
    Assert-True (@($continuity.traces).Count -ge 1) "machine scenario continuity traces missing"
    Assert-True (@($continuity.state_components).Count -ge 4) "machine scenario continuity state components missing"
    Add-CheckResult -Results $results -Area "observability" -Operation "logs / traces / continuity confirmed" -Status "passed" -Evidence "request_id=$requestId continuity_logs=$(@($continuity.logs).Count) continuity_traces=$(@($continuity.traces).Count) continuity_state_components=$(@($continuity.state_components).Count) latest_timeline_present=$latestTimelinePresent total_logs=$(@($tickLogs).Count)"

    $summary = [pscustomobject]@{
        engine_port = $EnginePort
        config_path = $configPath
        db_path = $dbPath
        world_id = $worldId
        node_id = $npcId
        request_id = $requestId
        callback_id = $callbackId
        task_id = $completedTask.task_id
        latest_timeline_present = $latestTimelinePresent
        checks = $results
    }

    $summaryJson = $summary | ConvertTo-Json -Depth 16
    $dir = Split-Path -Parent $OutFile
    if ($dir -and -not (Test-Path $dir)) { New-Item -ItemType Directory -Force -Path $dir | Out-Null }
    Set-Content -LiteralPath $OutFile -Value $summaryJson
    $summaryJson
} finally {
    Stop-ManagedProcess -Process $engineProc
}
