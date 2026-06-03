$ErrorActionPreference = "Stop"

function Invoke-JsonPost {
    param(
        [string]$Url,
        [object]$BodyObj
    )
    $json = $BodyObj | ConvertTo-Json -Depth 20 -Compress
    return Invoke-RestMethod -Method Post -Uri $Url -ContentType 'application/json' -Body $json
}

$username = "u" + [DateTimeOffset]::Now.ToUnixTimeSeconds()
$registerResp = Invoke-JsonPost -Url 'http://localhost:8080/api/auth/register' -BodyObj @{ username = $username; password = '123456' }
$userId = [int]$registerResp.user_id

$saveResp = Invoke-JsonPost -Url 'http://localhost:8080/api/person/profile' -BodyObj @{
    user_id = $userId
    height = 172
    weight = 66
    birth_date = '1998-05-01'
    gender = 'male'
    core_goals = @('减脂')
    target_intensity = 'normal'
    auto_add_to_plan = $false
}

$getResp = Invoke-RestMethod -Method Get -Uri ("http://localhost:8080/api/person/profile?user_id=" + $userId)

Write-Output ("save_resp=" + ($saveResp | ConvertTo-Json -Compress))
Write-Output ("get_resp=" + ($getResp | ConvertTo-Json -Depth 8 -Compress))
