# Install/Register Clipboard Server to Windows Startup

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
Set-Location $scriptDir

Write-Host "Building cp-server-bg.exe (headless without console window)..." -ForegroundColor Cyan
$env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")
if (-not (Test-Path "$scriptDir\rsrc.syso")) {
    go run github.com/akavel/rsrc@latest -ico "$scriptDir\icon.ico" -o "$scriptDir\rsrc.syso"
}
go build -ldflags "-H=windowsgui" -o "$scriptDir\cp-server-bg.exe" "$scriptDir\main.go"

if (-not (Test-Path "$scriptDir\cp-server-bg.exe")) {
    Write-Error "Failed to build cp-server-bg.exe!"
    exit 1
}

$startupFolder = [System.Environment]::GetFolderPath('Startup')
$shortcutPath = Join-Path $startupFolder "cp-server.lnk"

Write-Host "Creating startup shortcut at: $shortcutPath" -ForegroundColor Cyan
$wshShell = New-Object -ComObject WScript.Shell
$shortcut = $wshShell.CreateShortcut($shortcutPath)
$shortcut.TargetPath = "$scriptDir\cp-server-bg.exe"
$shortcut.WorkingDirectory = $scriptDir
$shortcut.Description = "Clipboard Server for iOS sync"
$shortcut.Save()

Write-Host "`nSuccessfully installed to Startup!" -ForegroundColor Green
Write-Host "cp-server will now automatically run silently in the background whenever you log into Windows."
Write-Host "To start it now in background: Start-Process '$scriptDir\cp-server-bg.exe'"
