param(
    [string]$RepoRoot = ".",
    [string]$EngineExePath,
    [string]$DevCliPath,
    [string]$WorkerExePath,
    [string]$FixtureFile = "",
    [string]$TradeDynamicInterfacesFile = "",
    [string]$WorldTimeSettingsPath = "",
    [string]$WorldStatePath = "",
    [string]$TickPolicyPath = "",
    [string]$ApiKey = "dev-key",
    [string]$RuntimeTaskToken = "dev-task-token",
    [int]$EnginePort = 18084,
    [string]$OutFile = ""
)

$ErrorActionPreference = "Stop"
$ScriptDir = if ($PSScriptRoot) { $PSScriptRoot } else { Split-Path -Parent $MyInvocation.MyCommand.Path }
$SdkToolingSmokePath = Join-Path $ScriptDir "sdk_tooling_smoke.go"

if ([string]::IsNullOrWhiteSpace($FixtureFile)) { $FixtureFile = Join-Path $ScriptDir "tooling_smoke_fixture.json" }
if ([string]::IsNullOrWhiteSpace($TradeDynamicInterfacesFile)) { $TradeDynamicInterfacesFile = Join-Path $ScriptDir "runtime_task_dynamic_action_trade.json" }
if ([string]::IsNullOrWhiteSpace($WorldTimeSettingsPath)) { $WorldTimeSettingsPath = Join-Path $ScriptDir "world_time_settings_flexible.json" }
if ([string]::IsNullOrWhiteSpace($WorldStatePath)) { $WorldStatePath = Join-Path $ScriptDir "state_world_state.json" }
if ([string]::IsNullOrWhiteSpace($TickPolicyPath)) { $TickPolicyPath = Join-Path $ScriptDir "state_tick_policy.json" }
if ([string]::IsNullOrWhiteSpace($OutFile)) { $OutFile = Join-Path $ScriptDir "full_functional_tooling_smoke_result.json" }

function Assert-True {
    param(
        [Parameter(Mandatory = $true)][bool]$Condition,
        [Parameter(Mandatory = $true)][string]$Message
    )

    if (-not $Condition) {
        throw $Message
    }
}

function Assert-Equal {
    param(
        [Parameter(Mandatory = $true)]$Actual,
        [Parameter(Mandatory = $true)]$Expected,
        [Parameter(Mandatory = $true)][string]$Message
    )

    if ($Actual -ne $Expected) {
        throw "$Message. expected=[$Expected] actual=[$Actual]"
    }
}

function Invoke-EngineJson {
    param(
        [Parameter(Mandatory = $true)][string]$Method,
        [Parameter(Mandatory = $true)][string]$Path,
        [object]$Body = $null,
        [hashtable]$ExtraHeaders = @{}
    )

    $headers = @{ "X-API-Key" = $ApiKey; "Content-Type" = "application/json" }
    foreach ($key in $ExtraHeaders.Keys) {
        $headers[$key] = $ExtraHeaders[$key]
    }
    $uri = "http://127.0.0.1:$EnginePort$Path"
    if ($null -eq $Body) {
        return Invoke-RestMethod -Method $Method -Uri $uri -Headers $headers
    }
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
    if ($LASTEXITCODE -ne 0) {
        throw "DevCli failed: $($allArgs -join ' ')"
    }
    if ([string]::IsNullOrWhiteSpace(($output -join "`n"))) {
        return $null
    }
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

    if ($null -eq $Process) {
        return
    }
    try {
        if (-not $Process.HasExited) {
            Stop-Process -Id $Process.Id -Force
        }
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

    $Results.Add([pscustomobject]@{
        area = $Area
        operation = $Operation
        status = $Status
        evidence = $Evidence
    }) | Out-Null
}

if (-not (Test-Path $EngineExePath)) { throw "EngineExePath not found: $EngineExePath" }
if (-not (Test-Path $DevCliPath)) { throw "DevCliPath not found: $DevCliPath" }
if (-not (Test-Path $WorkerExePath)) { throw "WorkerExePath not found: $WorkerExePath" }
foreach ($path in @($FixtureFile, $TradeDynamicInterfacesFile, $WorldTimeSettingsPath, $WorldStatePath, $TickPolicyPath, $SdkToolingSmokePath)) {
    if (-not (Test-Path $path)) {
        throw "Required file not found: $path"
    }
}

$results = New-Object 'System.Collections.Generic.List[object]'
$tempRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("gae-s8-src-" + (Get-Date -Format "yyyyMMddHHmmss"))
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
  callback_token: "dev-callback-token"
  runtime_task_token: "$RuntimeTaskToken"

llm:
  provider: "fixture"
  model: "fixture-s8"
  api_key: ""
  base_url: ""
  fixture_file: "$((Resolve-Path $FixtureFile).Path -replace '\\', '\\')"

engine:
  execution_mode: "debug"
  world_lock_enabled: true
  autonomous_scheduler_enabled: false
  runtime_task_governance_interval_seconds: 0

external_interfaces:
  npc_trade_action:
    category: "external_action"
    delivery_mode: "pull"
    primary_transport: "task_pull"
    consumer: "bridge"
    resume_policy: "none"
"@
Set-Content -LiteralPath $configPath -Value $configText -Encoding ASCII

$engineProc = $null

try {
    $engineProc = Start-HiddenProcess -FilePath $EngineExePath -ArgumentList @("serve", "--config", $configPath) -StdOutPath $engineStdOut -StdErrPath $engineStdErr
    Wait-EngineHealthy

    $world = Run-DevCli -Args @("node", "create", "--type", "world", "--name", "FullFunctionalToolingSmokeWorld")
    $worldId = $world.id
    $npc = Run-DevCli -Args @("node", "create", "--world", $worldId, "--type", "npc", "--name", "Tooling Merchant")
    $npcId = $npc.id

    $settings = Run-DevCli -Args @("world", "settings", "set", $worldId, "--world-time-settings-file", $WorldTimeSettingsPath, "--pipeline-mode", "full")
    Assert-Equal $settings.pipeline_mode "full" "tooling smoke pipeline mode mismatch"
    $null = Run-DevCli -Args @("state", "set", $worldId, "world_state", "--file", $WorldStatePath)
    $null = Run-DevCli -Args @("state", "set", $worldId, "tick_policy", "--file", $TickPolicyPath)

    $taskInvoke = Run-DevCli -Args @("invoke", $worldId, $npcId, "--task-type", "custom", "--dynamic-interfaces-file", $TradeDynamicInterfacesFile, "--message", "tooling smoke task")
    $taskCallbackId = $taskInvoke.action_calls[0].callback_id
    Assert-True (-not [string]::IsNullOrWhiteSpace($taskCallbackId)) "tooling smoke callback_id missing"

    $pendingTasks = Invoke-RuntimeTaskJson -Method GET -Path "/api/v1/runtime/tasks/pending?consumer=bridge&limit=20"
    Assert-True (@($pendingTasks.tasks).Count -ge 1) "tooling smoke expected pending tasks"
    Add-CheckResult -Results $results -Area "sdk" -Operation "seed runtime task" -Status "passed" -Evidence "callback_id=$taskCallbackId pending_count=$(@($pendingTasks.tasks).Count)"

    $tick = Run-DevCli -Args @("world", "tick", $worldId, "--type", "manual", "--time", "day-12 hour-8", "--requested-ticks", "2", "--autonomous-limit", "0")
    Assert-Equal $tick.advanced_ticks 2 "tooling smoke tick advanced_ticks mismatch"
    Add-CheckResult -Results $results -Area "continuity" -Operation "seed world tick" -Status "passed" -Evidence "request_id=$($tick.invoke.request_id)"

    $sdkSmoke = (& go run $SdkToolingSmokePath --server "http://127.0.0.1:$EnginePort" --key $ApiKey --world $worldId --node $npcId) -join "`n"
    if ($LASTEXITCODE -ne 0) {
        throw "sdk tooling smoke failed"
    }
    $sdkResult = $sdkSmoke | ConvertFrom-Json
    Assert-Equal $sdkResult.pending_task_iface "npc_trade_action" "sdk pending task interface mismatch"
    Assert-True ($sdkResult.trace_count -ge 1) "sdk trace_count should be >= 1"
    Assert-True ($sdkResult.continuity_trace_count -ge 1) "sdk continuity_trace_count should be >= 1"
    Add-CheckResult -Results $results -Area "sdk" -Operation "runtime task helper smoke" -Status "passed" -Evidence "pending_task_id=$($sdkResult.pending_task_id) latest_tick=$($sdkResult.latest_tick_number)"

    $nodeList = Run-DevCli -Args @("node", "list", "--world", $worldId)
    $legacyNodes = Run-DevCli -Args @("nodes", "--world", $worldId)
    Assert-Equal @($nodeList).Count @($legacyNodes).Count "DevCli node list and legacy nodes count mismatch"

    $taskList = Invoke-RuntimeTaskJson -Method GET -Path "/api/v1/runtime/tasks?world_id=$worldId&limit=20"
    $taskInspect = Run-DevCli -Args @("task", "get", $sdkResult.pending_task_id, "--json")
    Assert-Equal $taskInspect.task_id $sdkResult.pending_task_id "DevCli task get id mismatch"
    Assert-Equal $taskInspect.interface_name "npc_trade_action" "DevCli task get interface mismatch"

    $continuity = Run-DevCli -Args @("debug", "continuity", $worldId, "--json")
    Assert-True ($null -ne $continuity.latest_timeline) "DevCli continuity latest_timeline missing"
    Assert-True (@($continuity.traces).Count -ge 1) "DevCli continuity traces missing"

    $traces = Run-DevCli -Args @("debug", "traces", "--world", $worldId, "--json")
    Assert-True ($traces.count -ge 1) "DevCli traces count should be >= 1"
    Add-CheckResult -Results $results -Area "devcli" -Operation "node/task compatibility smoke" -Status "passed" -Evidence "node_count=$(@($nodeList).Count) task_id=$($taskInspect.task_id)"

    $creatorTasksApis = @(
        "/api/v1/runtime/tasks?limit=100&world_id=$worldId",
        "/api/v1/runtime/tasks/stats"
    )
    foreach ($path in $creatorTasksApis) {
        $resp = Invoke-EngineJson -Method GET -Path $path
        Assert-True ($null -ne $resp) "Creator Tasks API returned null for $path"
    }
    Assert-True (@($taskList.tasks).Count -ge 1) "Creator Tasks source task list should contain at least one task"
    Assert-Equal $taskList.tasks[0].task_id $sdkResult.pending_task_id "Creator Tasks first task id should match SDK pending task"
    Add-CheckResult -Results $results -Area "creator" -Operation "Tasks page smoke" -Status "passed" -Evidence "task_id=$($taskList.tasks[0].task_id) stats_total=$($sdkResult.runtime_task_total)"

    $latestTimeline = Invoke-EngineJson -Method GET -Path "/api/v1/worlds/$worldId/timelines/latest"
    $timelines = Invoke-EngineJson -Method GET -Path "/api/v1/worlds/$worldId/timelines?limit=6"
    $stateComponents = Invoke-EngineJson -Method GET -Path "/api/v1/worlds/$worldId/state-components"
    $logs = Invoke-EngineJson -Method GET -Path "/api/v1/logs?world_id=$worldId&task_type=world_tick&limit=60"
    $continuityTraces = Invoke-EngineJson -Method GET -Path "/debug/traces?world_id=$worldId&limit=30"
    Assert-True ($null -ne $latestTimeline.timeline) "Creator Continuity latest timeline missing"
    Assert-True (@($timelines.timelines).Count -ge 1) "Creator Continuity timelines missing"
    Assert-True (@($stateComponents.components).Count -ge 2) "Creator Continuity state components missing"
    Assert-True (@($logs).Count -ge 1) "Creator Continuity logs missing"
    Assert-True ($continuityTraces.count -ge 1) "Creator Continuity traces missing"
    Add-CheckResult -Results $results -Area "creator" -Operation "Continuity page smoke" -Status "passed" -Evidence "timeline_tick=$($latestTimeline.timeline.tick_number) state_components=$(@($stateComponents.components).Count)"

    $tracePage = Invoke-EngineJson -Method GET -Path "/debug/traces?world_id=$worldId&limit=30"
    Assert-True ($tracePage.count -ge 1) "Creator Traces page source missing traces"
    Add-CheckResult -Results $results -Area "creator" -Operation "Traces page smoke" -Status "passed" -Evidence "trace_count=$($tracePage.count)"

    $summary = [pscustomobject]@{
        engine_port = $EnginePort
        config_path = $configPath
        db_path = $dbPath
        world_id = $worldId
        node_id = $npcId
        pending_task_id = $sdkResult.pending_task_id
        latest_tick_number = $sdkResult.latest_tick_number
        checks = $results
    }

    $summaryJson = $summary | ConvertTo-Json -Depth 16
    $dir = Split-Path -Parent $OutFile
    if ($dir -and -not (Test-Path $dir)) {
        New-Item -ItemType Directory -Force -Path $dir | Out-Null
    }
    Set-Content -LiteralPath $OutFile -Value $summaryJson
    $summaryJson
} finally {
    Stop-ManagedProcess -Process $engineProc
}
