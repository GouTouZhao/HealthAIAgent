$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

$baseUrl = "http://localhost:8080"
$results = @()

function Add-Result {
    param(
        [string]$Step,
        [bool]$Ok,
        [double]$ElapsedMs,
        [string]$Message
    )
    $script:results += [PSCustomObject]@{
        step = $Step
        ok = $Ok
        elapsed_ms = [math]::Round($ElapsedMs, 1)
        message = $Message
    }
    $status = if ($Ok) { "OK" } else { "FAIL" }
    Write-Output ("[{0}] {1} ({2} ms) - {3}" -f $status, $Step, [math]::Round($ElapsedMs, 1), $Message)
}

function Invoke-Step {
    param(
        [string]$Step,
        [scriptblock]$Action
    )
    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    try {
        $value = & $Action
        $sw.Stop()
        Add-Result -Step $Step -Ok $true -ElapsedMs $sw.Elapsed.TotalMilliseconds -Message "success"
        return $value
    } catch {
        $sw.Stop()
        $msg = ""
        if ($_.Exception -and $_.Exception.Message) {
            $msg = $_.Exception.Message
        } else {
            $msg = "unknown error"
        }
        Add-Result -Step $Step -Ok $false -ElapsedMs $sw.Elapsed.TotalMilliseconds -Message $msg
        return $null
    }
}

function Invoke-JsonPost {
    param(
        [string]$Url,
        [object]$BodyObj,
        [int]$TimeoutSec = 90
    )
    $json = $BodyObj | ConvertTo-Json -Depth 40 -Compress
    return Invoke-RestMethod -Method Post -Uri $Url -ContentType 'application/json' -Body $json -TimeoutSec $TimeoutSec
}

function Invoke-JsonGet {
    param(
        [string]$Url,
        [int]$TimeoutSec = 30
    )
    return Invoke-RestMethod -Method Get -Uri $Url -TimeoutSec $TimeoutSec
}

function Invoke-JsonDelete {
    param(
        [string]$Url,
        [int]$TimeoutSec = 30
    )
    return Invoke-RestMethod -Method Delete -Uri $Url -TimeoutSec $TimeoutSec
}

$username = "user" + [DateTimeOffset]::Now.ToUnixTimeSeconds()
$password = "123456"
$chatId = "chat_" + [DateTimeOffset]::Now.ToUnixTimeSeconds()

$registerResp = Invoke-Step -Step "auth.register" -Action {
    Invoke-JsonPost -Url "$baseUrl/api/auth/register" -BodyObj @{ username = $username; password = $password }
}
if ($null -eq $registerResp) {
    Write-Output "[DONE] register failed, stop smoke test"
    exit 1
}
$userId = [int]$registerResp.user_id

Invoke-Step -Step "auth.login" -Action {
    Invoke-JsonPost -Url "$baseUrl/api/auth/login" -BodyObj @{ username = $username; password = $password }
} | Out-Null

Invoke-Step -Step "person.account.get" -Action {
    Invoke-JsonGet -Url "$baseUrl/api/person/account?user_id=$userId"
} | Out-Null

$renamedUsername = $username + "_n"
$pixel = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO8K0nQAAAAASUVORK5CYII="
Invoke-Step -Step "person.account.update" -Action {
    Invoke-JsonPost -Url "$baseUrl/api/person/account" -BodyObj @{ user_id = $userId; username = $renamedUsername; avatar_base64 = ("data:image/png;base64," + $pixel) }
} | Out-Null

$newPassword = "654321"
Invoke-Step -Step "person.password.reset" -Action {
    Invoke-JsonPost -Url "$baseUrl/api/person/reset-password" -BodyObj @{ user_id = $userId; old_password = $password; new_password = $newPassword }
} | Out-Null

Invoke-Step -Step "auth.login.after-reset" -Action {
    Invoke-JsonPost -Url "$baseUrl/api/auth/login" -BodyObj @{ username = $renamedUsername; password = $newPassword }
} | Out-Null

Invoke-Step -Step "person.profile.save" -Action {
    Invoke-JsonPost -Url "$baseUrl/api/person/profile" -BodyObj @{
        user_id = $userId
        height = 175
        weight = 78
        weight_unit = "kg"
        birth_date = "1997-06-15"
        gender = "male"
        core_goals = @("fat_loss", "fitness")
        detailed_goal = "reduce body fat and improve endurance"
        injury_history = ""
        favorite_sports = "running,cycling"
        auto_add_to_plan = $true
        sports_description = "3 sessions per week, 40 minutes each, medium intensity"
        target_intensity = "normal"
    }
} | Out-Null

Invoke-Step -Step "person.profile.get" -Action {
    Invoke-JsonGet -Url "$baseUrl/api/person/profile?user_id=$userId"
} | Out-Null

$saveWeightResp = Invoke-Step -Step "person.weight-record.save" -Action {
    Invoke-JsonPost -Url "$baseUrl/api/person/weight-record" -BodyObj @{
        user_id = $userId
        weight = 77.6
        weight_unit = "kg"
        state = "fasting"
    }
}

Invoke-Step -Step "person.weight-record.get" -Action {
    Invoke-JsonGet -Url "$baseUrl/api/person/weight-record?user_id=$userId&limit=10"
} | Out-Null

Invoke-Step -Step "person.weight-dashboard.get" -Action {
    Invoke-JsonGet -Url "$baseUrl/api/person/weight-dashboard?user_id=$userId&view=day"
} | Out-Null

if ($null -ne $saveWeightResp -and $saveWeightResp.record_id) {
    Invoke-Step -Step "person.weight-record.delete" -Action {
        Invoke-JsonDelete -Url "$baseUrl/api/person/weight-record?user_id=$userId&record_id=$($saveWeightResp.record_id)"
    } | Out-Null
}

$askOtherResp = Invoke-Step -Step "agent.ask.mode4" -Action {
    Invoke-JsonPost -Url "$baseUrl/api/agent/ask" -TimeoutSec 180 -BodyObj @{
        user_id = $userId
        chat_id = $chatId
        mode = "4_OtherQuestion"
        user_input = "I have tailbone discomfort. What core exercise can replace crunches?"
    }
}

Invoke-Step -Step "agent.ask-stream.mode4" -Action {
    $payload = @{ user_id = $userId; chat_id = $chatId; mode = "4_OtherQuestion"; user_input = "Give me one warm-up suggestion before training." } | ConvertTo-Json -Depth 10 -Compress
    Invoke-WebRequest -Method Post -Uri "$baseUrl/api/agent/ask-stream" -ContentType 'application/json' -Body $payload -TimeoutSec 180 | Out-Null
} | Out-Null

Invoke-Step -Step "agent.ask.mode1" -Action {
    Invoke-JsonPost -Url "$baseUrl/api/agent/ask" -TimeoutSec 180 -BodyObj @{
        user_id = $userId
        chat_id = $chatId
        mode = "1_FoodRecognition"
        user_input = "apple"
    }
} | Out-Null

$makePlanResp = Invoke-Step -Step "agent.ask.mode2" -Action {
    Invoke-JsonPost -Url "$baseUrl/api/agent/ask" -TimeoutSec 240 -BodyObj @{
        user_id = $userId
        chat_id = $chatId
        mode = "2_MakePlan"
        user_input = "Please make a detailed 3-month fat loss plan from Monday to Sunday."
    }
}

if ($null -ne $makePlanResp) {
    Invoke-Step -Step "agent.ask.mode3" -Action {
        Invoke-JsonPost -Url "$baseUrl/api/agent/ask" -TimeoutSec 240 -BodyObj @{
            user_id = $userId
            chat_id = $chatId
            mode = "3_ChangePlan"
            user_input = "Replace Wednesday running with swimming, keep others unchanged."
        }
    } | Out-Null
}

$todayResp = Invoke-Step -Step "agent.today-plan.get" -Action {
    Invoke-JsonGet -Url "$baseUrl/api/agent/today-plan?user_id=$userId"
}

if ($null -ne $todayResp -and $todayResp.tasks -and $todayResp.tasks.Count -gt 0) {
    $firstTask = $todayResp.tasks[0]
    Invoke-Step -Step "agent.toggle-task.post" -Action {
        Invoke-JsonPost -Url "$baseUrl/api/agent/toggle-task" -BodyObj @{
            user_id = $userId
            task_date = $todayResp.task_date
            activity = $firstTask.activity
            completed = $true
        }
    } | Out-Null
}

$chatHistoryResp = Invoke-Step -Step "agent.chat-history.get" -Action {
    Invoke-JsonGet -Url "$baseUrl/api/agent/chat-history?user_id=$userId&chat_id=$chatId&pair_limit=5"
}

Invoke-Step -Step "agent.plan-history.get" -Action {
    Invoke-JsonGet -Url "$baseUrl/api/agent/plan-history?user_id=$userId"
} | Out-Null

$planHistoryResp = Invoke-Step -Step "agent.plan-history.get.for-mutation" -Action {
    Invoke-JsonGet -Url "$baseUrl/api/agent/plan-history?user_id=$userId"
}

$historyItems = @()
if ($null -ne $planHistoryResp -and $null -ne $planHistoryResp.items) {
    $historyItems = @($planHistoryResp.items)
}

if ($historyItems.Count -gt 0) {
    $historyRecordId = $historyItems[0].id
    Invoke-Step -Step "agent.plan-history.replace" -Action {
        Invoke-JsonPost -Url "$baseUrl/api/agent/plan-history/replace" -BodyObj @{ user_id = $userId; record_id = $historyRecordId }
    } | Out-Null

    $planHistoryAfterReplace = Invoke-Step -Step "agent.plan-history.get.after-replace" -Action {
        Invoke-JsonGet -Url "$baseUrl/api/agent/plan-history?user_id=$userId"
    }

    $postReplaceItems = @()
    if ($null -ne $planHistoryAfterReplace -and $null -ne $planHistoryAfterReplace.items) {
        $postReplaceItems = @($planHistoryAfterReplace.items)
    }
    if ($postReplaceItems.Count -gt 0) {
        $deleteRecordId = $postReplaceItems[0].id
        Invoke-Step -Step "agent.plan-history.delete" -Action {
            Invoke-JsonDelete -Url "$baseUrl/api/agent/plan-history?user_id=$userId&record_id=$deleteRecordId"
        } | Out-Null
    }
} else {
    Add-Result -Step "agent.plan-history.replace" -Ok $true -ElapsedMs 0 -Message "skipped(no history item)"
}

$memoryResp = Invoke-Step -Step "person.memory.get" -Action {
    Invoke-JsonGet -Url "$baseUrl/api/person/memory?user_id=$userId"
}

if ($null -ne $memoryResp -and $memoryResp.items -and $memoryResp.items.Count -gt 0) {
    $firstMemoryId = $memoryResp.items[0].id
    Invoke-Step -Step "person.memory.delete" -Action {
        Invoke-JsonDelete -Url "$baseUrl/api/person/memory?user_id=$userId&memory_id=$firstMemoryId"
    } | Out-Null
}

if ($null -ne $chatHistoryResp) {
    Invoke-Step -Step "agent.chat-history.delete" -Action {
        Invoke-JsonDelete -Url "$baseUrl/api/agent/chat-history?user_id=$userId&chat_id=$chatId"
    } | Out-Null
}

Write-Output ""
Write-Output "================ Smoke Test Summary ================"
$results | Format-Table -AutoSize
$passedCount = ($results | Where-Object { $_.ok }).Count
$failedCount = ($results | Where-Object { -not $_.ok }).Count
Write-Output ("total=" + $results.Count + ", passed=" + $passedCount + ", failed=" + $failedCount)
Write-Output "===================================================="
