$startupFolder = [System.Environment]::GetFolderPath('Startup')
$shortcutPath = Join-Path $startupFolder "cp-server.lnk"

if (Test-Path $shortcutPath) {
    Remove-Item $shortcutPath -Force
    Write-Host "Removed Startup shortcut: $shortcutPath" -ForegroundColor Yellow
}

# Stop any running instances
Stop-Process -Name "cp-server", "cp-server-bg" -ErrorAction SilentlyContinue
Write-Host "Stopped any running cp-server processes." -ForegroundColor Green
