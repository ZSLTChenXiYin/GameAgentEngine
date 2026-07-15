param(
    [string]$RepoRoot = ".",
    [string]$EngineExePath,
    [string]$DevCliPath,
    [string]$WorkerExePath,
    [string]$FixtureFile = "",
    [string]$TradeDynamicInterfacesFile = "",
    [string]$ApiKey = "dev-key",
    [string]$CallbackToken = "dev-callback-token",
    [string]$RuntimeTaskToken = "dev-task-token",
    [string]$GameHTTPBearerToken = "local-test-token",
    [int]$EnginePort = 18082,
    [int]$PushPort = 19000,
    [string]$OutFile = ""
)

$ErrorActionPreference = "Stop"
$ScriptDir = if ($PSScriptRoot) { $PSScriptRoot } else { Split-Path -Parent $MyInvocation.MyCommand.Path }

if ([string]::IsNullOrWhiteSpace($FixtureFile)) { $FixtureFile = Join-Path $ScriptDir "runtime_task_delivery_fixture.json" }
if ([string]::IsNullOrWhiteSpace($TradeDynamicInterfacesFile)) { $TradeDynamicInterfacesFile = Join-Path $ScriptDir "runtime_task_dynamic_action_trade.json" }
if ([string]::IsNullOrWhiteSpace($OutFile)) { $OutFile = Join-Path $ScriptDir "full_functional_runtime_tasks_result.json" }

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

function Run-DevCliRaw {
    param([Parameter(Mandatory = $true)][string[]]$Args)

    $allArgs = @("--server", "http://127.0.0.1:$EnginePort", "--key", $ApiKey) + $Args
    $output = & $DevCliPath @allArgs 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "DevCli failed: $($allArgs -join ' ')`n$($output -join "`n")"
    }
    return ($output -join "`n")
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

function Get-WorldTasks {
    param([Parameter(Mandatory = $true)][string]$WorldId)

    $resp = Invoke-RuntimeTaskJson -Method GET -Path "/api/v1/runtime/tasks?world_id=$WorldId&limit=200"
    return @($resp.tasks)
}

function Get-TaskByCallbackId {
    param(
        [Parameter(Mandatory = $true)][string]$WorldId,
        [Parameter(Mandatory = $true)][string]$CallbackId
    )

    $tasks = Get-WorldTasks -WorldId $WorldId
    return $tasks | Where-Object { $_.callback_id -eq $CallbackId } | Select-Object -First 1
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
        if ($null -ne $task -and $task.status -eq $ExpectedStatus) {
            return $task
        }
        Start-Sleep -Milliseconds 500
    }
    $last = Get-TaskByCallbackId -WorldId $WorldId -CallbackId $CallbackId
    throw "task for callback_id=$CallbackId did not reach status=$ExpectedStatus; last=$($last | ConvertTo-Json -Depth 8 -Compress)"
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
foreach ($path in @($FixtureFile, $TradeDynamicInterfacesFile)) {
    if (-not (Test-Path $path)) {
        throw "Required file not found: $path"
    }
}

$results = New-Object 'System.Collections.Generic.List[object]'
$tempRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("gae-s6-src-" + (Get-Date -Format "yyyyMMddHHmmss"))
New-Item -ItemType Directory -Force -Path $tempRoot | Out-Null
$dbPath = Join-Path $tempRoot "gameagentengine.db"
$configPath = Join-Path $tempRoot "gameagentengine.conf.yaml"
$engineStdOut = Join-Path $tempRoot "engine.stdout.log"
$engineStdErr = Join-Path $tempRoot "engine.stderr.log"
$workerStdOut = Join-Path $tempRoot "worker.stdout.log"
$workerStdErr = Join-Path $tempRoot "worker.stderr.log"

$configText = @"
server:
  host: "127.0.0.1"
  port: $EnginePort

database:
  driver: "sqlite"
  dsn: "$($dbPath -replace '\\', '\\')"
  migrations_enabled: true
  write_retry_enabled: true
  write_retry_max_attempts: 3
  write_retry_base_delay_ms: 40
  write_retry_max_delay_ms: 250
  log_batch_enabled: true
  log_batch_size: 32
  log_batch_flush_ms: 750
  log_batch_queue_size: 1024

auth:
  api_key: "$ApiKey"
  callback_token: "$CallbackToken"
  runtime_task_token: "$RuntimeTaskToken"
  callback_require_request_id: true

llm:
  provider: "fixture"
  model: "fixture-s6"
  api_key: ""
  base_url: ""
  fixture_file: "$((Resolve-Path $FixtureFile).Path -replace '\\', '\\')"

engine:
  execution_mode: "debug"
  world_lock_enabled: true
  autonomous_scheduler_enabled: false
  autonomous_scheduler_interval_seconds: 300
  autonomous_scheduler_max_nodes_per_world: 10
  runtime_task_governance_interval_seconds: 0
  runtime_task_heartbeat_timeout_seconds: 2
  runtime_task_auto_requeue_enabled: false
  runtime_task_auto_requeue_limit: 100
  runtime_task_auto_requeue_delay_ms: 50

external_integrations:
  game_http:
    type: "http_adapter"
    base_url: "http://127.0.0.1:$PushPort"
    path: "/api/v1/runtime/dispatch"
    timeout_ms: 1000
    retry_max_attempts: 1
    retry_backoff_ms: 50
    idempotency_header: "Idempotency-Key"
    auth:
      mode: "bearer"
      token: "$GameHTTPBearerToken"

external_interfaces:
  game_client_request_data:
    category: "external_query"
    delivery_mode: "push"
    primary_transport: "game_http"
    consumer: "game_client"
    resume_policy: "resume_paused_execution"

  spawn_item:
    category: "external_action"
    delivery_mode: "hybrid"
    primary_transport: "game_http"
    fallback_transport: "task_pull"
    consumer: "bridge"
    max_attempts: 3
    heartbeat_timeout_auto_requeue: true
    heartbeat_timeout_requeue_delay_ms: 500
    heartbeat_timeout_reason: "spawn_item timeout auto requeue"
    resume_policy: "none"
    callback_post_process: "write_memory"
    callback_memory_level: "long_term"
    callback_memory_template: "spawn callback {status}: {result_json}"

  npc_trade_action:
    category: "external_action"
    delivery_mode: "pull"
    primary_transport: "task_pull"
    consumer: "bridge"
    max_attempts: 3
    resume_policy: "none"
"@
Set-Content -LiteralPath $configPath -Value $configText -Encoding ASCII

$engineProc = $null
$pushWorkerProc = $null

try {
    $engineProc = Start-HiddenProcess -FilePath $EngineExePath -ArgumentList @("serve", "--config", $configPath) -StdOutPath $engineStdOut -StdErrPath $engineStdErr
    Wait-EngineHealthy

    $world = Run-DevCli -Args @("node", "create", "--type", "world", "--name", "FullFunctionalRuntimeTasksWorld")
    $worldId = $world.id
    $npc = Run-DevCli -Args @("node", "create", "--world", $worldId, "--type", "npc", "--name", "Broker Toma")
    $npcId = $npc.id

    $pushWorkerProc = Start-HiddenProcess -FilePath $WorkerExePath -ArgumentList @(
        "push-receiver",
        "--engine-base-url", "http://127.0.0.1:$EnginePort",
        "--runtime-task-token", $RuntimeTaskToken,
        "--callback-token", $CallbackToken,
        "--game-http-bearer-token", $GameHTTPBearerToken,
        "--push-port", "$PushPort",
        "--callback-delay", "250ms"
    ) -StdOutPath $workerStdOut -StdErrPath $workerStdErr
    Start-Sleep -Seconds 1

    $pushInvoke = Run-DevCli -Args @("invoke", $worldId, $npcId, "--task-type", "custom", "--message", "push runtime task")
    $pushCallbackId = $pushInvoke.action_calls[0].callback_id
    Assert-True (-not [string]::IsNullOrWhiteSpace($pushCallbackId)) "push callback_id missing"
    $pushTask = Wait-TaskStatus -WorldId $worldId -CallbackId $pushCallbackId -ExpectedStatus "succeeded"
    Assert-Equal $pushTask.delivery_mode "push" "push task delivery_mode mismatch"
    Assert-Equal $pushTask.transport "game_http" "push task transport mismatch"
    Add-CheckResult -Results $results -Area "push" -Operation "delivery success" -Status "passed" -Evidence "task_id=$($pushTask.task_id) dispatch_attempts=$($pushTask.dispatch_attempts)"

    $pullInvoke = Run-DevCli -Args @("invoke", $worldId, $npcId, "--task-type", "custom", "--dynamic-interfaces-file", $TradeDynamicInterfacesFile, "--message", "pull runtime task")
    $pullCallbackId = $pullInvoke.action_calls[0].callback_id
    $pullTaskPending = Get-TaskByCallbackId -WorldId $worldId -CallbackId $pullCallbackId
    Assert-Equal $pullTaskPending.status "pending" "pull task should start pending"
    Assert-Equal $pullTaskPending.interface_name "npc_trade_action" "pull task interface mismatch"
    $null = & $WorkerExePath "pull-once" "--engine-base-url" "http://127.0.0.1:$EnginePort" "--runtime-task-token" $RuntimeTaskToken "--callback-token" $CallbackToken "--consumer" "bridge" "--lease-owner" "s6-pull-once"
    if ($LASTEXITCODE -ne 0) { throw "pull-once worker failed" }
    $pullTask = Wait-TaskStatus -WorldId $worldId -CallbackId $pullCallbackId -ExpectedStatus "succeeded"
    Add-CheckResult -Results $results -Area "pull" -Operation "delivery success" -Status "passed" -Evidence "task_id=$($pullTask.task_id) completed_at=$($pullTask.completed_at)"

    Stop-ManagedProcess -Process $pushWorkerProc
    $pushWorkerProc = $null
    Start-Sleep -Seconds 1

    $hybridInvoke = Run-DevCli -Args @("invoke", $worldId, $npcId, "--task-type", "custom", "--message", "hybrid runtime task")
    $hybridCallbackId = $hybridInvoke.action_calls[0].callback_id
    $hybridTask = Wait-TaskStatus -WorldId $worldId -CallbackId $hybridCallbackId -ExpectedStatus "released"
    Assert-Equal $hybridTask.delivery_mode "hybrid" "hybrid task delivery_mode mismatch"
    Assert-Equal $hybridTask.transport "task_pull" "hybrid fallback transport mismatch"
    Assert-Equal $hybridTask.last_dispatch_decision "fallback_to_pull" "hybrid dispatch decision mismatch"
    Add-CheckResult -Results $results -Area "hybrid" -Operation "fallback transition" -Status "passed" -Evidence "task_id=$($hybridTask.task_id) failure_class=$($hybridTask.last_dispatch_failure_class)"

    $manualInvoke = Run-DevCli -Args @("invoke", $worldId, $npcId, "--task-type", "custom", "--dynamic-interfaces-file", $TradeDynamicInterfacesFile, "--message", "manual claim task")
    $manualCallbackId = $manualInvoke.action_calls[0].callback_id
    $manualTaskPending = Get-TaskByCallbackId -WorldId $worldId -CallbackId $manualCallbackId
    $claimed = Invoke-RuntimeTaskJson -Method POST -Path "/api/v1/runtime/tasks/claim" -Body @{ task_id = $manualTaskPending.task_id; consumer = "bridge"; lease_owner = "s6-manual" }
    Assert-Equal $claimed.task.status "claimed" "manual task claim status mismatch"
    $started = Invoke-RuntimeTaskJson -Method POST -Path "/api/v1/runtime/tasks/start" -Body @{ task_id = $manualTaskPending.task_id; lease_token = $claimed.task.lease_token }
    Assert-Equal $started.task.status "running" "manual task start status mismatch"
    $heartbeat = Invoke-RuntimeTaskJson -Method POST -Path "/api/v1/runtime/tasks/heartbeat" -Body @{ task_id = $manualTaskPending.task_id; lease_token = $claimed.task.lease_token }
    Assert-Equal $heartbeat.task.status "running" "manual task heartbeat status mismatch"
    $released = Invoke-RuntimeTaskJson -Method POST -Path "/api/v1/runtime/tasks/release" -Body @{ task_id = $manualTaskPending.task_id; lease_token = $claimed.task.lease_token; retry_delay_ms = 0; error_message = "manual release" }
    Assert-Equal $released.task.status "released" "manual task release status mismatch"
    Add-CheckResult -Results $results -Area "pull" -Operation "claim/start/heartbeat/release" -Status "passed" -Evidence "task_id=$($manualTaskPending.task_id)"

    $timeoutInvoke = Run-DevCli -Args @("invoke", $worldId, $npcId, "--task-type", "custom", "--dynamic-interfaces-file", $TradeDynamicInterfacesFile, "--message", "manual timeout task")
    $timeoutCallbackId = $timeoutInvoke.action_calls[0].callback_id
    $timeoutTaskPending = Get-TaskByCallbackId -WorldId $worldId -CallbackId $timeoutCallbackId
    $timeoutClaimed = Invoke-RuntimeTaskJson -Method POST -Path "/api/v1/runtime/tasks/claim" -Body @{ task_id = $timeoutTaskPending.task_id; consumer = "bridge"; lease_owner = "s6-timeout" }
    $timeoutStarted = Invoke-RuntimeTaskJson -Method POST -Path "/api/v1/runtime/tasks/start" -Body @{ task_id = $timeoutTaskPending.task_id; lease_token = $timeoutClaimed.task.lease_token }
    Assert-Equal $timeoutStarted.task.status "running" "timeout task start status mismatch"
    Start-Sleep -Seconds 3
    $swept = Invoke-RuntimeTaskJson -Method POST -Path "/api/v1/runtime/tasks/heartbeat-timeout/sweep" -Body @{ timeout_seconds = 1 }
    Assert-True ($swept.affected -ge 1) "heartbeat timeout sweep affected no tasks"
    $timeoutTask = Wait-TaskStatus -WorldId $worldId -CallbackId $timeoutCallbackId -ExpectedStatus "heartbeat_timeout"
    $requeued = Invoke-RuntimeTaskJson -Method POST -Path "/api/v1/runtime/tasks/requeue" -Body @{ task_id = $timeoutTask.task_id; retry_delay_ms = 0; error_message = "manual requeue" }
    Assert-Equal $requeued.task.status "released" "timeout task requeue status mismatch"
    Add-CheckResult -Results $results -Area "pull" -Operation "heartbeat-timeout and requeue" -Status "passed" -Evidence "task_id=$($timeoutTask.task_id) timeout_count=$($timeoutTask.heartbeat_timeout_count)"

    $worldTasks = Get-WorldTasks -WorldId $worldId
    Assert-True ($worldTasks.Count -ge 5) "world runtime task count should be at least 5"
    $statsOutput = Run-DevCliRaw -Args @("task", "stats")
    Assert-True ($statsOutput.Contains("Runtime Task Stats")) "task stats output missing header"
    Assert-True ($statsOutput.Contains("fallback_to_pull")) "task stats output missing fallback_to_pull"

    $inspectOutput = Run-DevCliRaw -Args @("task", "inspect", $hybridTask.task_id)
    Assert-True ($inspectOutput.Contains("dispatch_decision=fallback_to_pull")) "task inspect missing dispatch decision"
    Assert-True ($inspectOutput.Contains("payload=")) "task inspect missing payload"
    Add-CheckResult -Results $results -Area "diagnostics" -Operation "list/stats/inspect" -Status "passed" -Evidence "inspect_task_id=$($hybridTask.task_id) world_task_count=$($worldTasks.Count)"

    $summary = [pscustomobject]@{
        engine_port = $EnginePort
        push_port = $PushPort
        config_path = $configPath
        db_path = $dbPath
        world_id = $worldId
        push_task_id = $pushTask.task_id
        pull_task_id = $pullTask.task_id
        hybrid_task_id = $hybridTask.task_id
        timeout_task_id = $timeoutTask.task_id
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
    Stop-ManagedProcess -Process $pushWorkerProc
    Stop-ManagedProcess -Process $engineProc
}
