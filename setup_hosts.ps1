# Self-elevate the script to run as Administrator if not already elevated
$myCurrentUser = [Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()
if (-not $myCurrentUser.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
    Write-Host "Requesting Administrator privileges to modify the hosts file..." -ForegroundColor Yellow
    Start-Process powershell.exe -ArgumentList "-NoProfile -ExecutionPolicy Bypass -File `"$PSCommandPath`"" -Verb RunAs
    exit
}

$hostsPath = "C:\Windows\System32\drivers\etc\hosts"
$entry = "127.0.0.1 birdnet-go.local"

Write-Host "Checking hosts file for birdnet-go.local..." -ForegroundColor Cyan

if ((Get-Content $hostsPath) -notcontains $entry) {
    # Ensure there is a newline before appending
    Add-Content -Path $hostsPath -Value "`n$entry" -Force
    Write-Host "Successfully added birdnet-go.local to your hosts file!" -ForegroundColor Green
} else {
    Write-Host "The entry birdnet-go.local already exists in your hosts file." -ForegroundColor Yellow
}

Write-Host "Press any key to close..." -ForegroundColor Gray
$null = [System.Console]::ReadKey()
