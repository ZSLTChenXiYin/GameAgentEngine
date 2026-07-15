param(
    [string]$EngineBaseUrl = "http://127.0.0.1:8080",
    [string]$ApiKey = "dev-key",
    [string]$DevCliPath = ".\\dist\\GameAgentEngine-windows-amd64-v0.4.6\\GameAgentDevCli.exe",
    [string]$FixturePath = ".\\docs\\tests\\full_functional_base_data_world.yaml",
    [string]$WorldName = "FullFunctionalBaseDataWorld",
    [string]$StressWorldName = "FullFunctionalBaseDataRaceWorld",
    [string]$RunSuffix = "",
    [switch]$ResetBeforeImport,
    [string]$OutFile = ""
)

$ErrorActionPreference = "Stop"

function New-Headers {
    return @{
        "X-API-Key" = $ApiKey
        "Content-Type" = "application/json"
    }
}

function Invoke-EngineJson {
    param(
        [Parameter(Mandatory = $true)][string]$Method,
        [Parameter(Mandatory = $true)][string]$Path,
        [object]$Body = $null
    )

    $uri = "$EngineBaseUrl$Path"
    if ($null -eq $Body) {
        return Invoke-RestMethod -Method $Method -Uri $uri -Headers (New-Headers)
    }
    $json = $Body | ConvertTo-Json -Depth 16 -Compress
    return Invoke-RestMethod -Method $Method -Uri $uri -Headers (New-Headers) -Body $json
}

function Run-DevCli {
    param(
        [Parameter(Mandatory = $true)][string[]]$Args
    )

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

function Get-StatusCode {
    param([Parameter(Mandatory = $true)]$ErrorRecord)

    if ($null -ne $ErrorRecord.Exception.Response -and $null -ne $ErrorRecord.Exception.Response.StatusCode) {
        return [int]$ErrorRecord.Exception.Response.StatusCode
    }
    if ($null -ne $ErrorRecord.Exception.StatusCode) {
        return [int]$ErrorRecord.Exception.StatusCode
    }
    return -1
}

function Assert-NotFound {
    param(
        [Parameter(Mandatory = $true)][scriptblock]$Action,
        [Parameter(Mandatory = $true)][string]$Message
    )

    try {
        & $Action
    } catch {
        if ((Get-StatusCode -ErrorRecord $_) -eq 404) {
            return
        }
        throw
    }
    throw $Message
}

function Find-WorldByName {
    param([Parameter(Mandatory = $true)][string]$Name)

    $worlds = Invoke-EngineJson -Method GET -Path "/api/v1/worlds"
    return $worlds | Where-Object { $_.name -eq $Name } | Select-Object -First 1
}

function Find-NodeByName {
    param(
        [Parameter(Mandatory = $true)][string]$WorldId,
        [Parameter(Mandatory = $true)][string]$Name
    )

    $nodes = Invoke-EngineJson -Method GET -Path "/api/v1/nodes?world_id=$WorldId"
    return $nodes | Where-Object { $_.name -eq $Name } | Select-Object -First 1
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

if (-not (Test-Path $DevCliPath)) {
    throw "DevCli not found: $DevCliPath"
}
if (-not (Test-Path $FixturePath)) {
    throw "Fixture not found: $FixturePath"
}

$results = New-Object 'System.Collections.Generic.List[object]'

if ([string]::IsNullOrWhiteSpace($RunSuffix)) {
    $RunSuffix = Get-Date -Format "yyyyMMddHHmmss"
}

$resolvedWorldName = "$WorldName-$RunSuffix"
$resolvedStressWorldName = "$StressWorldName-$RunSuffix"
$rawFixture = Get-Content -LiteralPath $FixturePath -Raw
$generatedFixture = $rawFixture -replace '(?m)^  name: .*$', "  name: $resolvedWorldName"
$generatedFixturePath = Join-Path ([System.IO.Path]::GetTempPath()) ("full-functional-base-data-" + $RunSuffix + ".yaml")
Set-Content -LiteralPath $generatedFixturePath -Value $generatedFixture -Encoding ASCII

$importArgs = @("import", $generatedFixturePath)
if ($ResetBeforeImport) {
    $importArgs += "--reset"
}
$importResult = Run-DevCli -Args $importArgs
Assert-Equal $importResult.world_name $resolvedWorldName "fixture import world name mismatch"
Add-CheckResult -Results $results -Area "world" -Operation "import fixture" -Channel "devcli" -Status "passed" -Evidence "world_id=$($importResult.world_id)"

$world = Find-WorldByName -Name $resolvedWorldName
Assert-True ($null -ne $world) "imported world not found via HTTP"
$worldId = $world.id
Add-CheckResult -Results $results -Area "world" -Operation "locate imported world" -Channel "http" -Status "passed" -Evidence "world_id=$worldId"

$baseNpc = Find-NodeByName -WorldId $worldId -Name "Quartermaster Lin"
Assert-True ($null -ne $baseNpc) "base npc not found"

$nodeHttp = Invoke-EngineJson -Method POST -Path "/api/v1/nodes" -Body @{
    world_id = $worldId
    name = "Watch Captain Rhea"
    node_type = "npc"
    parent_id = $baseNpc.id
}
$nodeId = $nodeHttp.id
Assert-Equal $nodeHttp.name "Watch Captain Rhea" "http node create returned wrong name"
Add-CheckResult -Results $results -Area "node" -Operation "create" -Channel "http" -Status "passed" -Evidence "node_id=$nodeId"

$nodeCli = Run-DevCli -Args @("node", "get", $nodeId)
Assert-Equal $nodeCli.node.name "Watch Captain Rhea" "devcli node get mismatch after http create"
Add-CheckResult -Results $results -Area "node" -Operation "get after create" -Channel "devcli" -Status "passed" -Evidence "node_id=$nodeId"

$updatedNode = Run-DevCli -Args @("node", "update", $nodeId, "--name", "Watch Captain Rhea II", "--type", "npc")
Assert-Equal $updatedNode.name "Watch Captain Rhea II" "devcli node update returned wrong name"
$nodeHttpAfterUpdate = Invoke-EngineJson -Method GET -Path "/api/v1/nodes/$nodeId"
Assert-Equal $nodeHttpAfterUpdate.node.name "Watch Captain Rhea II" "http node get mismatch after devcli update"
Add-CheckResult -Results $results -Area "node" -Operation "update" -Channel "devcli->http" -Status "passed" -Evidence "node_id=$nodeId"

$legacyNodes = Run-DevCli -Args @("nodes", "--world", $worldId)
$httpNodes = Invoke-EngineJson -Method GET -Path "/api/v1/nodes?world_id=$worldId"
Assert-Equal ($legacyNodes | Measure-Object | Select-Object -ExpandProperty Count) ($httpNodes | Measure-Object | Select-Object -ExpandProperty Count) "legacy nodes count mismatch"
Add-CheckResult -Results $results -Area "node" -Operation "legacy list parity" -Channel "devcli/http" -Status "passed" -Evidence "count=$($httpNodes.Count)"

$componentHttp = Invoke-EngineJson -Method POST -Path "/api/v1/components" -Body @{
    node_id = $nodeId
    component_type = "rule"
    data = "watch=north"
}
$componentId = $componentHttp.id
Assert-Equal $componentHttp.component_type "rule" "http component create returned wrong type"
Add-CheckResult -Results $results -Area "component" -Operation "create" -Channel "http" -Status "passed" -Evidence "component_id=$componentId"

$componentCli = Run-DevCli -Args @("component", "get", $componentId)
Assert-Equal $componentCli.id $componentId "devcli component get mismatch after http create"
$null = Run-DevCli -Args @("component", "update", $componentId, "--data", "watch=west")
$componentHttpAfterUpdate = Invoke-EngineJson -Method GET -Path "/api/v1/components/$componentId"
Assert-True ($componentHttpAfterUpdate.data -like '*west*') "http component get mismatch after devcli update"
Add-CheckResult -Results $results -Area "component" -Operation "update" -Channel "devcli->http" -Status "passed" -Evidence "component_id=$componentId"

$memoryCli = Run-DevCli -Args @("memory", "create", "--node", $nodeId, "--content", "Northern watch changed patrol route.", "--level", "short_term", "--tags", "patrol,watch")
$memoryId = $memoryCli.id
Assert-Equal $memoryCli.node_id $nodeId "devcli memory create returned wrong node"
Add-CheckResult -Results $results -Area "memory" -Operation "create" -Channel "devcli" -Status "passed" -Evidence "memory_id=$memoryId"

$memoryHttp = Invoke-EngineJson -Method GET -Path "/api/v1/memories/$memoryId"
Assert-Equal $memoryHttp.content "Northern watch changed patrol route." "http memory get mismatch after devcli create"
$updatedMemory = Invoke-EngineJson -Method PUT -Path "/api/v1/memories/$memoryId" -Body @{
    content = "Northern watch rerouted through the western stairs."
    tags = "patrol,west"
}
Assert-True ($updatedMemory.content -like '*western stairs*') "http memory update returned wrong content"
$memoryCliAfterUpdate = Run-DevCli -Args @("memory", "get", $memoryId)
Assert-True ($memoryCliAfterUpdate.content -like '*western stairs*') "devcli memory get mismatch after http update"
Add-CheckResult -Results $results -Area "memory" -Operation "update" -Channel "http->devcli" -Status "passed" -Evidence "memory_id=$memoryId"

$relationHttp = Invoke-EngineJson -Method POST -Path "/api/v1/relations" -Body @{
    world_id = $worldId
    source_id = $nodeId
    target_id = $baseNpc.id
    relation_type = "subordinate"
    weight = 5
    properties = '{"duty":"watch_command"}'
}
$relationId = $relationHttp.id
Assert-Equal $relationHttp.relation_type "subordinate" "http relation create returned wrong type"
Add-CheckResult -Results $results -Area "relation" -Operation "create" -Channel "http" -Status "passed" -Evidence "relation_id=$relationId"

$relationCli = Run-DevCli -Args @("relation", "get", $relationId)
Assert-Equal $relationCli.id $relationId "devcli relation get mismatch after http create"
$null = Run-DevCli -Args @("relation", "update", $relationId, "--weight", "7", "--props", '{"duty":"west_watch"}')
$relationHttpAfterUpdate = Invoke-EngineJson -Method GET -Path "/api/v1/relations/$relationId"
Assert-Equal $relationHttpAfterUpdate.weight 7 "http relation get mismatch after devcli update"
Add-CheckResult -Results $results -Area "relation" -Operation "update" -Channel "devcli->http" -Status "passed" -Evidence "relation_id=$relationId"

$worldSettingsCli = Run-DevCli -Args @(
    "world", "settings", "set", $worldId,
    "--memory-limit", "24",
    "--analysis-rounds", "3",
    "--context-depth", "4",
    "--auto-apply=false",
    "--review-above", "high",
    "--propagation-max-depth", "2",
    "--enable-propagation-machine=true",
    "--sub-task-max-retries", "5",
    "--sub-task-timeout-secs", "90",
    "--pipeline-mode", "polling"
)
Assert-Equal $worldSettingsCli.pipeline_mode "polling" "devcli world settings set returned wrong pipeline mode"
$worldSettingsHttp = Invoke-EngineJson -Method GET -Path "/api/v1/worlds/$worldId/settings"
Assert-Equal $worldSettingsHttp.pipeline_mode "polling" "http world settings get mismatch after devcli set"
Assert-Equal $worldSettingsHttp.memory_limit 24 "http world settings get mismatch for memory_limit"
Add-CheckResult -Results $results -Area "world_settings" -Operation "set/get" -Channel "devcli->http" -Status "passed" -Evidence "pipeline_mode=$($worldSettingsHttp.pipeline_mode)"

$worldPolicyHttp = Invoke-EngineJson -Method PUT -Path "/api/v1/worlds/$worldId/policy" -Body @{
    blocked_actions = @("spawn_item")
    safe_actions = @("inspect_map", "request_backup")
}
Assert-True (($worldPolicyHttp.blocked_actions | Measure-Object).Count -eq 1) "http world policy set mismatch"
$worldPolicyCli = Run-DevCli -Args @("world", "policy", "get", $worldId)
Assert-Equal (($worldPolicyCli.safe_actions -join ",")) "inspect_map,request_backup" "devcli world policy get mismatch after http set"
Add-CheckResult -Results $results -Area "world_policy" -Operation "set/get" -Channel "http->devcli" -Status "passed" -Evidence "blocked=$($worldPolicyCli.blocked_actions -join ',')"

$stressWorld = Run-DevCli -Args @("node", "create", "--type", "world", "--name", $resolvedStressWorldName)
$stressWorldId = $stressWorld.id
$stressNode = Run-DevCli -Args @("node", "create", "--world", $stressWorldId, "--type", "npc", "--name", "Stress Harness Node")
$stressNodeId = $stressNode.id

$jobs = @()
for ($i = 0; $i -lt 6; $i++) {
    $jobs += Start-Job -ScriptBlock {
        param($BaseUrl, $Key, $NodeId, $Index)
        $headers = @{ "X-API-Key" = $Key; "Content-Type" = "application/json" }
        $body = @{ node_id = $NodeId; component_type = "rule"; data = ('{"slot":' + $Index + '}') } | ConvertTo-Json -Compress
        Invoke-RestMethod -Method POST -Uri ($BaseUrl + "/api/v1/components") -Headers $headers -Body $body | Out-Null
    } -ArgumentList $EngineBaseUrl, $ApiKey, $stressNodeId, $i
}
Wait-Job $jobs | Out-Null
$jobErrors = @($jobs | Where-Object { $_.State -ne "Completed" })
if ($jobErrors.Count -gt 0) {
    $jobErrors | Receive-Job -Keep | Out-Null
    throw "component/world_settings stress jobs did not all complete successfully"
}
$null = $jobs | Receive-Job
$jobs | Remove-Job | Out-Null

$stressComponents = Invoke-EngineJson -Method GET -Path "/api/v1/components?node_id=$stressNodeId"
Assert-Equal ($stressComponents | Measure-Object | Select-Object -ExpandProperty Count) 6 "concurrent component creation count mismatch"
Add-CheckResult -Results $results -Area "component" -Operation "concurrent create on fresh world" -Channel "http" -Status "passed" -Evidence "world_id=$stressWorldId count=6"

$deleteComponent = Run-DevCli -Args @("component", "delete", $componentId)
Assert-Equal $deleteComponent.status "deleted" "devcli component delete failed"
Assert-NotFound -Action { Invoke-EngineJson -Method GET -Path "/api/v1/components/$componentId" | Out-Null } -Message "component still exists after delete"
Add-CheckResult -Results $results -Area "component" -Operation "delete" -Channel "devcli->http" -Status "passed" -Evidence "component_id=$componentId"

$deleteMemory = Run-DevCli -Args @("memory", "delete", $memoryId)
Assert-Equal $deleteMemory.status "deleted" "devcli memory delete failed"
Assert-NotFound -Action { Invoke-EngineJson -Method GET -Path "/api/v1/memories/$memoryId" | Out-Null } -Message "memory still exists after delete"
Add-CheckResult -Results $results -Area "memory" -Operation "delete" -Channel "devcli->http" -Status "passed" -Evidence "memory_id=$memoryId"

$deleteRelation = Invoke-EngineJson -Method DELETE -Path "/api/v1/relations/$relationId"
Add-CheckResult -Results $results -Area "relation" -Operation "delete" -Channel "http" -Status "passed" -Evidence "relation_id=$relationId"
Assert-NotFound -Action { Invoke-EngineJson -Method GET -Path "/api/v1/relations/$relationId" | Out-Null } -Message "relation still exists after delete"

$deleteNode = Run-DevCli -Args @("node", "delete", $nodeId)
Assert-Equal $deleteNode.status "deleted" "devcli node delete failed"
Assert-NotFound -Action { Invoke-EngineJson -Method GET -Path "/api/v1/nodes/$nodeId" | Out-Null } -Message "node still exists after delete"
Add-CheckResult -Results $results -Area "node" -Operation "delete" -Channel "devcli->http" -Status "passed" -Evidence "node_id=$nodeId"

$summary = [pscustomobject]@{
    run_suffix = $RunSuffix
    world_name = $resolvedWorldName
    world_id = $worldId
    stress_world_name = $resolvedStressWorldName
    stress_world_id = $stressWorldId
    checks = $results
}

$summaryJson = $summary | ConvertTo-Json -Depth 12
if ($OutFile -ne "") {
    $dir = Split-Path -Parent $OutFile
    if ($dir -and -not (Test-Path $dir)) {
        New-Item -ItemType Directory -Force -Path $dir | Out-Null
    }
    Set-Content -LiteralPath $OutFile -Value $summaryJson
}

$summaryJson
