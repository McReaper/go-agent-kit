param(
    [string]$Password
)

$files = gci -Path $PSScriptRoot
Write-Host "Installing $($files.Count) files"
Invoke-Expression "Copy-Item $PSScriptRoot\notes.exe $env:LOCALAPPDATA\notes"
