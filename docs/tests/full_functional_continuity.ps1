param(
    [string]$EngineBaseUrl = "http://127.0.0.1:18080",
    [string]$ApiKey = "dev-key",
    [string]$DevCliPath,
    [string]$BaseDataResultPath = ".\\docs\\tests\\full_functional_base_data_result.json",
    [string]$WorldTimeSettingsPath = ".\\docs\\tests\\world_time_settings_flexible.json",
    [string]$WorldStatePath = ".\\docs\\tests\\state_world_state.json",
    [string]$StoryStatePath = ".\\docs\\tests\\state_story_state.json",
    [string]$StoryHistoryPath = ".\\docs\\tests\\state_story_history.json",
    [string]$TickPolicyPath = ".\\docs\\tests\\state_tick_policy.json",
    [string]$OutFile = ""
)

$ErrorActionPreference = "Stop"

function Invoke-EngineJson {
    param(
        [Parameter(Mandatory = $true)][string]$Method,
        [Parameter(Mandatory = $true)][string]$Path,
        [object]$Body = $null
    )

    $headers = @{ "X-API-Key" = $ApiKey; "Content-Type" = "application/json" }
    $uri = "$EngineBaseUrl$Path"
    if ($null -eq $Body) {
        return Invoke-RestMethod -Method $Method -Uri $uri -Headers $headers
    }
    $json = $Body | ConvertTo-Json -Depth 20 -Compress
    return Invoke-RestMethod -Method $Method -Uri $uri -Headers $headers -Body $json
}

function Run-DevCli {
    param([Parameter(Mandatory = $true)][string[]]$Args)

    $allArgs = @("--server", $EngineBaseUrl, "--key", $ApiKey) + $Args
    $output = & $DevCliPath @allArgs
    if ($LASTEXITCODE -ne 0) {
        throw "DevCli failed: $($allArgs -join ' ')"
    }
    if ([string]::IsNullOrWhiteSpace(($output -join "`n"))) {
        return $null
    }
    return ($output -join "`n") | ConvertFrom-Json
}

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

function Add-CheckResult {
    param(
        [Parameter(Mandatory = $true)][AllowEmptyCollection()][System.Collections.Generic.List[object]]$Results,
        [Parameter(Mandatory = $true)][string]$Area,
        [Parameter(Mandatory = $true)][string]$Operation,
        [Parameter(Mandatory = $true)][string]$Channel,
        [Parameter(Mandatory = $true)][string]$Status,
        [Parameter(Mandatory = $true)][string]$Evidence
    )

    $Results.Add([pscustomobject]@{
        area = $Area
        operation = $Operation
        channel = $Channel
        status = $Status
        evidence = $Evidence
    }) | Out-Null
}

if (-not $DevCliPath) {
    throw "DevCliPath is required"
}

foreach ($path in @($BaseDataResultPath, $WorldTimeSettingsPath, $WorldStatePath, $StoryStatePath, $StoryHistoryPath, $TickPolicyPath)) {
    if (-not (Test-Path $path)) {
        throw "Required file not found: $path"
    }
}

$baseData = Get-Content -LiteralPath $BaseDataResultPath -Raw | ConvertFrom-Json
$worldId = $baseData.world_id
Assert-True (-not [string]::IsNullOrWhiteSpace($worldId)) "world_id missing from base data result"

$results = New-Object 'System.Collections.Generic.List[object]'

$settings = Run-DevCli -Args @("world", "settings", "set", $worldId, "--world-time-settings-file", $WorldTimeSettingsPath, "--pipeline-mode", "full")
Assert-Equal $settings.pipeline_mode "full" "world settings pipeline mode mismatch"
Assert-Equal $settings.world_time_settings.tick_scale_mode "flexible" "world time settings tick_scale_mode mismatch"
Add-CheckResult -Results $results -Area "world_time_settings" -Operation "set" -Channel "devcli" -Status "passed" -Evidence "tick_scale_mode=$($settings.world_time_settings.tick_scale_mode)"

$stateWorld = Run-DevCli -Args @("state", "set", $worldId, "world_state", "--file", $WorldStatePath)
$stateStory = Run-DevCli -Args @("state", "set", $worldId, "story_state", "--file", $StoryStatePath)
$stateHistory = Run-DevCli -Args @("state", "set", $worldId, "story_history", "--file", $StoryHistoryPath)
$statePolicy = Run-DevCli -Args @("state", "set", $worldId, "tick_policy", "--file", $TickPolicyPath)
Assert-Equal $stateWorld.state_component.component_type "world_state" "world_state set failed"
Assert-Equal $stateStory.state_component.component_type "story_state" "story_state set failed"
Assert-Equal $stateHistory.state_component.component_type "story_history" "story_history set failed"
Assert-Equal $statePolicy.state_component.component_type "tick_policy" "tick_policy set failed"
Add-CheckResult -Results $results -Area "state" -Operation "seed continuity components" -Channel "devcli" -Status "passed" -Evidence "world_state/story_state/story_history/tick_policy"

$tick = Run-DevCli -Args @("world", "tick", $worldId, "--type", "manual", "--time", "day-12 hour-9", "--requested-ticks", "2", "--autonomous-limit", "0")
Assert-Equal $tick.advanced_ticks 2 "tick advanced_ticks mismatch"
Assert-True ($tick.world_time_state.current_time_label.Length -gt 0) "tick world_time_state label missing"
Assert-True ($tick.world_time_state.current_time_label -ne "day-12 hour-9") "tick world_time_state label did not advance"
Assert-Equal $tick.world_time_state.last_advanced_ticks 2 "tick last_advanced_ticks mismatch"
$requestId = $tick.invoke.request_id
Assert-True (-not [string]::IsNullOrWhiteSpace($requestId)) "tick request_id missing"
Add-CheckResult -Results $results -Area "tick" -Operation "advance world tick" -Channel "devcli" -Status "passed" -Evidence "request_id=$requestId advanced_ticks=$($tick.advanced_ticks)"

$timelineLatest = Run-DevCli -Args @("timeline", "latest", $worldId, "--json")
$timelineList = Run-DevCli -Args @("timeline", "list", $worldId, "--limit", "5", "--json")
Assert-True ($timelineLatest.timeline.tick_number -ge 1) "latest timeline tick_number should be at least 1"
Assert-Equal $timelineLatest.timeline.advanced_ticks 2 "latest timeline advanced_ticks mismatch"
Assert-Equal $timelineLatest.timeline.data.world_time_state.current_time_label $tick.world_time_state.current_time_label "latest timeline world time mismatch"
Assert-True (($timelineList.timelines | Measure-Object | Select-Object -ExpandProperty Count) -ge 1) "timeline list should contain at least one entry"
Assert-Equal $timelineList.timelines[0].tick_number $timelineLatest.timeline.tick_number "timeline latest/list head mismatch"
Add-CheckResult -Results $results -Area "timeline" -Operation "latest/list" -Channel "devcli" -Status "passed" -Evidence "latest_tick=$($timelineLatest.timeline.tick_number)"

$stateList = Run-DevCli -Args @("state", "list", $worldId, "--json")
$worldState = Run-DevCli -Args @("state", "get", $worldId, "world_state")
$storyState = Run-DevCli -Args @("state", "get", $worldId, "story_state")
$storyHistory = Run-DevCli -Args @("state", "get", $worldId, "story_history")
$worldTimeState = Run-DevCli -Args @("state", "get", $worldId, "world_time_state")
Assert-True (($stateList.components | Measure-Object | Select-Object -ExpandProperty Count) -ge 5) "state list should contain at least five continuity components"
Assert-True ($worldState.state_component.data.summary.Length -gt 0) "world_state summary missing"
Assert-True ($storyState.state_component.data.current_situation.Length -gt 0) "story_state current_situation missing"
Assert-True (($storyHistory.state_component.data.entries | Measure-Object | Select-Object -ExpandProperty Count) -ge 2) "story_history should include the new tick entry"
Assert-Equal $worldTimeState.state_component.data.current_time_label $tick.world_time_state.current_time_label "world_time_state current_time_label mismatch"
Add-CheckResult -Results $results -Area "state" -Operation "list/get" -Channel "devcli" -Status "passed" -Evidence "world_time_label=$($worldTimeState.state_component.data.current_time_label)"

$continuity = Run-DevCli -Args @("debug", "continuity", $worldId, "--request-id", $requestId, "--log-limit", "20", "--trace-limit", "10", "--json")
Assert-True ($null -ne $continuity.latest_timeline) "continuity latest_timeline missing"
Assert-True (($continuity.state_components | Measure-Object | Select-Object -ExpandProperty Count) -ge 5) "continuity state components missing"
Assert-True (($continuity.logs | Measure-Object | Select-Object -ExpandProperty Count) -ge 1) "continuity logs missing"
Assert-True (($continuity.traces | Measure-Object | Select-Object -ExpandProperty Count) -ge 1) "continuity traces missing"
Add-CheckResult -Results $results -Area "continuity" -Operation "debug continuity" -Channel "devcli" -Status "passed" -Evidence "logs=$($continuity.logs.Count) traces=$($continuity.traces.Count)"

$logs = Run-DevCli -Args @("logs", "--world", $worldId, "--task-type", "world_tick", "--request-id", $requestId, "--limit", "20", "--json")
$traces = Run-DevCli -Args @("debug", "traces", "--world", $worldId, "--limit", "10", "--json")
$requestLogs = @($logs | Where-Object { $_.request_id -eq $requestId })
$requestTraces = @($traces.traces | Where-Object { $_.request_id -eq $requestId })
Assert-True ($requestLogs.Count -ge 1) "request-scoped logs missing"
Assert-True ($requestTraces.Count -ge 1) "request-scoped traces missing"
Add-CheckResult -Results $results -Area "observability" -Operation "logs/traces correlation" -Channel "devcli" -Status "passed" -Evidence "request_id=$requestId logs=$($requestLogs.Count) traces=$($requestTraces.Count)"

$summary = [pscustomobject]@{
    world_id = $worldId
    request_id = $requestId
    latest_tick_number = $timelineLatest.timeline.tick_number
    latest_time_label = $worldTimeState.state_component.data.current_time_label
    checks = $results
}

$summaryJson = $summary | ConvertTo-Json -Depth 16
if ($OutFile -ne "") {
    $dir = Split-Path -Parent $OutFile
    if ($dir -and -not (Test-Path $dir)) {
        New-Item -ItemType Directory -Force -Path $dir | Out-Null
    }
    Set-Content -LiteralPath $OutFile -Value $summaryJson
}

$summaryJson
