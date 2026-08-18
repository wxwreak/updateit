$binPath = "$env:USERPROFILE\.local\bin"
$logPath = "$env:USERPROFILE\.local\share\updateit"
if (!(Test-Path $binPath)) { New-Item -ItemType Directory -Force -Path $binPath }
if (!(Test-Path $logPath)) { New-Item -ItemType Directory -Force -Path $logPath }
$url = "https://github.com/wxwreak/updateit/releases/latest/download/updateit-windows-amd64.exe"
$dest = "$binPath\updateit.exe"
Write-Host "Downloading updateit to $dest..."
Invoke-WebRequest -Uri $url -OutFile $dest
New-Item -Path "$logPath\latest.log" -ItemType File -Force | Out-Null
Write-Host "Updateit successfully installed to $binPath"
