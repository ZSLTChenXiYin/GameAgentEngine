param(
    [string]$RepoRoot = ".",
    [string]$EngineExePath,
    [string]$DevCliPath,
    [string]$WorkerExePath,
    [string]$FixtureFile = ".\docs\tests\callback_resume_fixture.json",
    [string]$DynamicDataInterfacesFile = ".\docs\tests\runtime_task_dynamic_interfaces.json",
    [string]$DynamicActionInterfacesFile = ".\docs\tests\callback_resume_dynamic_actions.json",
    [string]$ApiKey = "dev-key",
    [string]$CallbackToken = "dev-callback-token",
    [string]$RuntimeTaskToken = "dev-task-token",
    [int]$EnginePort = 18083,
    [string]$OutFile = ".\docs\tests\full_functional_callback_resume_result.json"
)

$ErrorActionPreference = "Stop"

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

function Invoke-CallbackWebRequest {
    param(
        [Parameter(Mandatory = $true)][string]$CallbackId,
        [Parameter(Mandatory = $true)][string]$Status,
        [object]$Result,
        [string]$CallbackRequestId = ""
    )

    $headers = @{ "Content-Type" = "application/json"; "X-Callback-Token" = $CallbackToken }
    if ($CallbackRequestId -ne "") {
        $headers["X-Callback-Request-Id"] = $CallbackRequestId
    }
    $body = @{ callback_id = $CallbackId; status = $Status; result = $Result } | ConvertTo-Json -Depth 20 -Compress
    return Invoke-WebRequest -Method POST -Uri "http://127.0.0.1:$EnginePort/api/v1/actions/callback" -Headers $headers -Body $body
}

function Invoke-RuntimeTaskJson {
    param(
        [Parameter(Mandatory = $true)][string]$Method,
        [Parameter(Mandatory = $true)][string]$Path,
        [object]$Body = $null
    )

    return Invoke-EngineJson -Method $Method -Path $Path -Body $Body -ExtraHeaders @{ "X-Runtime-Task-Token" = $RuntimeTaskToken }
}

function Invoke-InvokeJson {
    param(
        [Parameter(Mandatory = $true)][object]$Body
    )

    return Invoke-EngineJson -Method POST -Path "/api/v1/invoke" -Body $Body
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

function Wait-LogCount {
    param(
        [Parameter(Mandatory = $true)][string]$WorldId,
        [Parameter(Mandatory = $true)][string]$EventName,
        [int]$Minimum = 1,
        [int]$TimeoutSeconds = 15
    )

    $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
    while ((Get-Date) -lt $deadline) {
        $logs = Invoke-EngineJson -Method GET -Path "/api/v1/logs?world_id=$WorldId&event_name=$EventName&limit=100"
        if (@($logs).Count -ge $Minimum) {
            return @($logs)
        }
        Start-Sleep -Milliseconds 500
    }
    return @(Invoke-EngineJson -Method GET -Path "/api/v1/logs?world_id=$WorldId&event_name=$EventName&limit=100")
}

if (-not (Test-Path $EngineExePath)) { throw "EngineExePath not found: $EngineExePath" }
if (-not (Test-Path $DevCliPath)) { throw "DevCliPath not found: $DevCliPath" }
if (-not (Test-Path $WorkerExePath)) { throw "WorkerExePath not found: $WorkerExePath" }
foreach ($path in @($FixtureFile, $DynamicDataInterfacesFile, $DynamicActionInterfacesFile)) {
    if (-not (Test-Path $path)) {
        throw "Required file not found: $path"
    }
}

$results = New-Object 'System.Collections.Generic.List[object]'
$tempRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("gae-s7-src-" + (Get-Date -Format "yyyyMMddHHmmss"))
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
  model: "fixture-s7"
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

  game_client_request_data_no_resume:
    category: "external_query"
    delivery_mode: "pull"
    primary_transport: "task_pull"
    consumer: "game_client"
    resume_policy: "none"

  spawn_item_record_only:
    category: "external_action"
    delivery_mode: "pull"
    primary_transport: "task_pull"
    consumer: "bridge"
    resume_policy: "none"
    callback_post_process: "record_only"

  spawn_item_write_memory:
    category: "external_action"
    delivery_mode: "pull"
    primary_transport: "task_pull"
    consumer: "bridge"
    resume_policy: "none"
    callback_post_process: "write_memory"
    callback_memory_level: "long_term"
    callback_memory_template: "spawn callback {status}: {result_json}"

  spawn_item_failure:
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

    $world = Run-DevCli -Args @("node", "create", "--type", "world", "--name", "FullFunctionalCallbackResumeWorld")
    $worldId = $world.id
    $npc = Run-DevCli -Args @("node", "create", "--world", $worldId, "--type", "npc", "--name", "Callback Broker")
    $npcId = $npc.id

    $sceneInvoke = Invoke-InvokeJson -Body @{
        world_id = $worldId
        node_id = $npcId
        task_type = "custom"
        messages = @(@{ role = "user"; content = "scene pause" })
        context = @{ dynamic_interfaces = [object[]](Get-Content -LiteralPath $DynamicDataInterfacesFile -Raw | ConvertFrom-Json) }
    }
    $sceneCallbackId = $sceneInvoke.action_calls[0].callback_id
    Assert-True (-not [string]::IsNullOrWhiteSpace($sceneCallbackId)) "scene callback_id missing"
    $sceneTask = Wait-TaskStatus -WorldId $worldId -CallbackId $sceneCallbackId -ExpectedStatus "pending"
    Assert-Equal $sceneTask.interface_name "game_client_request_data" "scene task interface mismatch"

    $sceneCallbackResponse = Invoke-CallbackWebRequest -CallbackId $sceneCallbackId -Status "success" -Result @{ scene = "starter_inn"; npc = "merchant" } -CallbackRequestId "s7-scene-1"
    Assert-Equal $sceneCallbackResponse.StatusCode 200 "scene callback status code mismatch"
    $sceneCallbackBody = $sceneCallbackResponse.Content | ConvertFrom-Json
    Assert-Equal $sceneCallbackBody.resumed.reply "scene-resumed-final" "scene resumed reply mismatch"
    Assert-True (-not [string]::IsNullOrWhiteSpace($sceneCallbackBody.resume_execution_id)) "scene resume_execution_id missing"
    $sceneSucceededTask = Wait-TaskStatus -WorldId $worldId -CallbackId $sceneCallbackId -ExpectedStatus "succeeded"
    $resumeLogs = Wait-LogCount -WorldId $worldId -EventName "resume_completed" -Minimum 1
    $reuseLogs = Wait-LogCount -WorldId $worldId -EventName "data_request_reused" -Minimum 1
    Assert-Equal @($resumeLogs).Count 1 "resume_completed count mismatch after first resume"
    Assert-Equal @($reuseLogs).Count 1 "data_request_reused count mismatch"

    $sceneReplayResponse = Invoke-CallbackWebRequest -CallbackId $sceneCallbackId -Status "success" -Result @{ scene = "starter_inn"; npc = "merchant" } -CallbackRequestId "s7-scene-1"
    Assert-Equal $sceneReplayResponse.StatusCode 200 "scene replay callback status code mismatch"
    Assert-Equal $sceneReplayResponse.Headers["X-Callback-Replayed"] "true" "scene replay header mismatch"
    $resumeLogsAfterReplay = Invoke-EngineJson -Method GET -Path "/api/v1/logs?world_id=$worldId&event_name=resume_completed&limit=100"
    Assert-Equal @($resumeLogsAfterReplay).Count 1 "scene replay should not duplicate resume_completed log"
    $sceneTasks = Get-WorldTasks -WorldId $worldId | Where-Object { $_.interface_name -eq "game_client_request_data" }
    Assert-Equal @($sceneTasks).Count 1 "scene resume should not create duplicate data_request runtime task"
    Add-CheckResult -Results $results -Area "callback" -Operation "success path and auto-resume" -Status "passed" -Evidence "callback_id=$sceneCallbackId task_id=$($sceneSucceededTask.task_id) reply=$($sceneCallbackBody.resumed.reply)"
    Add-CheckResult -Results $results -Area "callback" -Operation "replay protection" -Status "passed" -Evidence "callback_id=$sceneCallbackId replayed=$($sceneReplayResponse.Headers['X-Callback-Replayed'])"
    Add-CheckResult -Results $results -Area "callback" -Operation "duplicate data_request suppression" -Status "passed" -Evidence "callback_id=$sceneCallbackId reuse_logs=$(@($reuseLogs).Count) task_count=$(@($sceneTasks).Count)"

    $noneInvoke = Invoke-InvokeJson -Body @{
        world_id = $worldId
        node_id = $npcId
        task_type = "custom"
        messages = @(@{ role = "user"; content = "scene none" })
        context = @{ dynamic_interfaces = [object[]]@(
            [pscustomobject]@{
                id = "scene_query_none"
                kind = "data_request"
                external_interface = "game_client_request_data_no_resume"
                description = "Query scene without auto resume."
                query_types = @("node_detail")
                max_queries = 1
            }
        ) }
    }
    $noneCallbackId = $noneInvoke.action_calls[0].callback_id
    $null = Wait-TaskStatus -WorldId $worldId -CallbackId $noneCallbackId -ExpectedStatus "pending"
    $noneCallbackResponse = Invoke-CallbackWebRequest -CallbackId $noneCallbackId -Status "success" -Result @{ scene = "outer_gate" } -CallbackRequestId "s7-none-1"
    $noneCallbackBody = $noneCallbackResponse.Content | ConvertFrom-Json
    Assert-True ($null -eq $noneCallbackBody.resumed) "resume_policy=none should not return resumed payload"
    $noneTask = Wait-TaskStatus -WorldId $worldId -CallbackId $noneCallbackId -ExpectedStatus "succeeded"
    $resumeLogsAfterNone = Invoke-EngineJson -Method GET -Path "/api/v1/logs?world_id=$worldId&event_name=resume_completed&limit=100"
    Assert-Equal @($resumeLogsAfterNone).Count 1 "resume_policy=none should not add resume_completed log"
    Add-CheckResult -Results $results -Area "callback" -Operation "resume_policy none" -Status "passed" -Evidence "callback_id=$noneCallbackId task_id=$($noneTask.task_id)"

    $recordInvoke = Invoke-InvokeJson -Body @{
        world_id = $worldId
        node_id = $npcId
        task_type = "custom"
        messages = @(@{ role = "user"; content = "record only action" })
        context = @{ dynamic_interfaces = [object[]]@((Get-Content -LiteralPath $DynamicActionInterfacesFile -Raw | ConvertFrom-Json | Select-Object -First 1)) }
    }
    $recordCallbackId = $recordInvoke.action_calls[0].callback_id
    $recordTaskPending = Wait-TaskStatus -WorldId $worldId -CallbackId $recordCallbackId -ExpectedStatus "pending"
    Assert-Equal $recordTaskPending.interface_name "spawn_item_record_only" "record_only task interface mismatch"
    $null = & $WorkerExePath "pull-once" "--engine-base-url" "http://127.0.0.1:$EnginePort" "--runtime-task-token" $RuntimeTaskToken "--callback-token" $CallbackToken "--consumer" "bridge" "--lease-owner" "s7-record"
    if ($LASTEXITCODE -ne 0) { throw "record_only pull-once worker failed" }
    $recordTask = Wait-TaskStatus -WorldId $worldId -CallbackId $recordCallbackId -ExpectedStatus "succeeded"
    $recordInspect = Run-DevCliRaw -Args @("task", "inspect", $recordTask.task_id)
    Assert-True ($recordInspect.Contains("resume_execution=-")) "record_only inspect should not show resume execution"
    $recordMemories = Run-DevCli -Args @("memory", "list", "--node", $npcId)
    Assert-Equal @($recordMemories).Count 0 "record_only should not write memory"
    Add-CheckResult -Results $results -Area "callback" -Operation "post-process record_only" -Status "passed" -Evidence "callback_id=$recordCallbackId task_id=$($recordTask.task_id)"

    $writeInvoke = Invoke-InvokeJson -Body @{
        world_id = $worldId
        node_id = $npcId
        task_type = "custom"
        messages = @(@{ role = "user"; content = "write memory action" })
        context = @{ dynamic_interfaces = [object[]]@(
            [pscustomobject]@{
                id = "spawn_write_memory"
                kind = "action"
                external_interface = "spawn_item_write_memory"
                description = "Write callback memory."
                args_schema = @{ type = "object"; additionalProperties = $true }
                max_calls = 1
            }
        ) }
    }
    $writeCallbackId = $writeInvoke.action_calls[0].callback_id
    $writeTaskPending = Wait-TaskStatus -WorldId $worldId -CallbackId $writeCallbackId -ExpectedStatus "pending"
    Assert-Equal $writeTaskPending.interface_name "spawn_item_write_memory" "write_memory task interface mismatch"
    $null = & $WorkerExePath "pull-once" "--engine-base-url" "http://127.0.0.1:$EnginePort" "--runtime-task-token" $RuntimeTaskToken "--callback-token" $CallbackToken "--consumer" "bridge" "--lease-owner" "s7-write"
    if ($LASTEXITCODE -ne 0) { throw "write_memory pull-once worker failed" }
    $writeTask = Wait-TaskStatus -WorldId $worldId -CallbackId $writeCallbackId -ExpectedStatus "succeeded"
    $writeMemories = Run-DevCli -Args @("memory", "list", "--node", $npcId)
    Assert-Equal @($writeMemories).Count 1 "write_memory should create one memory"
    Assert-Equal $writeMemories[0].level "long_term" "write_memory level mismatch"
    Assert-True ($writeMemories[0].content.Contains("spawn callback success:")) "write_memory content missing success template"
    Assert-True ($writeMemories[0].content.Contains('"item_name":"potion"')) "write_memory content missing callback payload"
    Add-CheckResult -Results $results -Area "callback" -Operation "post-process write_memory" -Status "passed" -Evidence "callback_id=$writeCallbackId memory_id=$($writeMemories[0].id)"

    $failureInvoke = Invoke-InvokeJson -Body @{
        world_id = $worldId
        node_id = $npcId
        task_type = "custom"
        messages = @(@{ role = "user"; content = "failure action" })
        context = @{ dynamic_interfaces = [object[]]@(
            [pscustomobject]@{
                id = "spawn_failure"
                kind = "action"
                external_interface = "spawn_item_failure"
                description = "Fail callback action."
                args_schema = @{ type = "object"; additionalProperties = $true }
                max_calls = 1
            }
        ) }
    }
    $failureCallbackId = $failureInvoke.action_calls[0].callback_id
    $failureTaskPending = Wait-TaskStatus -WorldId $worldId -CallbackId $failureCallbackId -ExpectedStatus "pending"
    Assert-Equal $failureTaskPending.interface_name "spawn_item_failure" "failure task interface mismatch"
    $null = & $WorkerExePath "pull-once" "--engine-base-url" "http://127.0.0.1:$EnginePort" "--runtime-task-token" $RuntimeTaskToken "--callback-token" $CallbackToken "--consumer" "bridge" "--lease-owner" "s7-fail" "--fail-interface" "spawn_item_failure"
    if ($LASTEXITCODE -ne 0) { throw "failure pull-once worker failed" }
    $failureTask = Wait-TaskStatus -WorldId $worldId -CallbackId $failureCallbackId -ExpectedStatus "failed"
    Assert-True ($failureTask.error_message.Contains('"status":"failed"')) "failure task error_message should include callback payload"
    Add-CheckResult -Results $results -Area "callback" -Operation "failure path" -Status "passed" -Evidence "callback_id=$failureCallbackId task_id=$($failureTask.task_id)"

    $summary = [pscustomobject]@{
        engine_port = $EnginePort
        config_path = $configPath
        db_path = $dbPath
        world_id = $worldId
        scene_callback_id = $sceneCallbackId
        scene_task_id = $sceneSucceededTask.task_id
        none_callback_id = $noneCallbackId
        none_task_id = $noneTask.task_id
        record_only_callback_id = $recordCallbackId
        record_only_task_id = $recordTask.task_id
        write_memory_callback_id = $writeCallbackId
        write_memory_task_id = $writeTask.task_id
        failure_callback_id = $failureCallbackId
        failure_task_id = $failureTask.task_id
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
