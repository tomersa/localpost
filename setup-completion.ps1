# setup-completion.ps1
# Automates autocompletion setup for localpost CLI on Windows

# Ensure we're in the correct directory
$scriptDir = Split-Path $PSCommandPath -Parent
Set-Location $scriptDir

# Generate completion script
$completionScript = "localpost_completion.ps1"
Write-Host "Generating PowerShell completion script..."
.\localpost.exe completion --shell powershell > $completionScript
if (-not (Test-Path $completionScript)) {
    Write-Host "Error: Failed to generate $completionScript" -ForegroundColor Red
    exit 1
}

# Create profile directory and file
$profileDir = Split-Path $PROFILE -Parent
New-Item -Path $profileDir -ItemType Directory -Force | Out-Null
New-Item -Path $PROFILE -ItemType File -Force | Out-Null
Write-Host "Created PowerShell profile at $PROFILE"

# Move completion script to profile directory
$destPath = Join-Path $profileDir $completionScript
Move-Item -Path $completionScript -Destination $destPath -Force
Write-Host "Moved completion script to $destPath"

# Add sourcing to profile
$sourceLine = ". '$destPath'"
$profileContent = Get-Content $PROFILE -ErrorAction SilentlyContinue
if ($profileContent -notcontains $sourceLine) {
    Add-Content -Path $PROFILE -Value $sourceLine
    Write-Host "Added completion script to PowerShell profile"
} else {
    Write-Host "Completion script already sourced in profile"
}

# Set execution policy if needed
$currentPolicy = Get-ExecutionPolicy -Scope CurrentUser
if ($currentPolicy -ne "RemoteSigned" -and $currentPolicy -ne "Unrestricted" -and $currentPolicy -ne "Bypass") {
    Write-Host "Setting execution policy to RemoteSigned..."
    Set-ExecutionPolicy -Scope CurrentUser -ExecutionPolicy RemoteSigned -Force
} else {
    Write-Host "Execution policy already allows scripts ($currentPolicy)"
}

# Reload profile
Write-Host "Reloading PowerShell profile..."
. $PROFILE

Write-Host "Autocompletion setup complete!" -ForegroundColor Green
Write-Host "Test it with: .\localpost.exe <Tab> or .\localpost.exe request <Tab>"