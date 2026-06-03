$ErrorActionPreference = "Stop"

function Invoke-JsonPost {
    param([string]$Url, [object]$BodyObj)
    $json = $BodyObj | ConvertTo-Json -Depth 20 -Compress
    Invoke-RestMethod -Method Post -Uri $Url -ContentType 'application/json' -Body $json
}

$username = "acc" + [DateTimeOffset]::Now.ToUnixTimeSeconds()
$registerResp = Invoke-JsonPost -Url 'http://localhost:8080/api/auth/register' -BodyObj @{ username = $username; password = '123456' }
$userId = [int]$registerResp.user_id

$accountBefore = Invoke-RestMethod -Method Get -Uri ("http://localhost:8080/api/person/account?user_id=" + $userId)

$pixel = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO8K0nQAAAAASUVORK5CYII='
$updateResp = Invoke-JsonPost -Url 'http://localhost:8080/api/person/account' -BodyObj @{ user_id = $userId; username = ($username + '_n'); avatar_base64 = ('data:image/png;base64,' + $pixel) }

$resetResp = Invoke-JsonPost -Url 'http://localhost:8080/api/person/reset-password' -BodyObj @{ user_id = $userId; old_password = '123456'; new_password = '654321' }
$loginResp = Invoke-JsonPost -Url 'http://localhost:8080/api/auth/login' -BodyObj @{ username = ($username + '_n'); password = '654321' }
$accountAfter = Invoke-RestMethod -Method Get -Uri ("http://localhost:8080/api/person/account?user_id=" + $userId)

Write-Output ("account_before=" + ($accountBefore | ConvertTo-Json -Compress))
Write-Output ("update_resp=" + ($updateResp | ConvertTo-Json -Compress))
Write-Output ("reset_resp=" + ($resetResp | ConvertTo-Json -Compress))
Write-Output ("login_resp=" + ($loginResp | ConvertTo-Json -Compress))
Write-Output ("account_after_has_avatar=" + [bool]($accountAfter.avatar_base64 -and $accountAfter.avatar_base64.Length -gt 0))
