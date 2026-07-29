# momo install script
# Usage: irm https://raw.githubusercontent.com/Bhagattji/momo/main/scripts/install.ps1 | iex

 = "Bhagattji/momo"
 = "momo-windows-amd64.exe"
 = "momo-windows-amd64.exe.sig"
 = "https://raw.githubusercontent.com/Bhagattji/momo/main/pubkey.asc"
 = "C:\Users\xbhag\.momo\bin"
 = "\momo.exe"

Write-Host "Installing momo from /latest ..." -ForegroundColor Cyan

# Create install dir
if (-not (Test-Path )) { New-Item -ItemType Directory -Path  -Force | Out-Null }

# Get latest release
 = Invoke-RestMethod "https://api.github.com/repos//releases/latest"
 = .assets | Where-Object { .name -eq  } | Select-Object -ExpandProperty browser_download_url
 = .assets | Where-Object { .name -eq  } | Select-Object -ExpandProperty browser_download_url

if (-not ) {
    Write-Error "Asset  not found in latest release"
    exit 1
}

Write-Host "Downloading  ..." -ForegroundColor Yellow
Invoke-WebRequest -Uri  -OutFile 

# Verify signature if sig exists
if () {
    Write-Host "Downloading signature ..." -ForegroundColor Yellow
     = ".sig"
    Invoke-WebRequest -Uri  -OutFile 
    
    # Download public key
    Write-Host "Downloading public key ..." -ForegroundColor Yellow
     = "C:\Users\xbhag\AppData\Local\Temp\momo-pubkey.asc"
    Invoke-WebRequest -Uri  -OutFile 
    
    # Import & verify
    Write-Host "Verifying signature ..." -ForegroundColor Yellow
     = "C:\Users\xbhag\AppData\Local\Temp\momo-gnupg"
    if (Test-Path ) { Remove-Item  -Recurse -Force }
    New-Item -ItemType Directory -Path  | Out-Null
     = 
    
    gpg --batch --import  2>
     = gpg --batch --verify   2>&1
    if (0 -eq 0) {
        Write-Host "✅ Signature verified!" -ForegroundColor Green
    } else {
        Write-Warning "⚠️ Signature verification failed: "
        Write-Warning "Proceeding anyway (insecure)."
    }
    Remove-Item  -Recurse -Force -ErrorAction SilentlyContinue
}

# Add to PATH (current session)
c:\Users\xbhag\AppData\Roaming\Code\User\globalStorage\github.copilot-chat\debugCommand;c:\Users\xbhag\AppData\Roaming\Code\User\globalStorage\github.copilot-chat\copilotCli;C:\WINDOWS\system32;C:\WINDOWS;C:\WINDOWS\System32\Wbem;C:\WINDOWS\System32\WindowsPowerShell\v1.0\;C:\WINDOWS\System32\OpenSSH\;C:\Program Files\nodejs\;C:\Program Files\Go\bin;C:\Program Files (x86)\Gpg4win\..\GnuPG\bin;C:\Program Files\Git\cmd;c:\Users\xbhag\AppData\Roaming\Code\User\globalStorage\github.copilot-chat\debugCommand;c:\Users\xbhag\AppData\Roaming\Code\User\globalStorage\github.copilot-chat\copilotCli;C:\WINDOWS\system32;C:\WINDOWS;C:\WINDOWS\System32\Wbem;C:\WINDOWS\System32\WindowsPowerShell\v1.0\;C:\WINDOWS\System32\OpenSSH\;C:\Program Files\nodejs\;C:\Program Files\Go\bin;c:\Users\xbhag\AppData\Roaming\Code\User\globalStorage\github.copilot-chat\debugCommand;c:\Users\xbhag\AppData\Roaming\Code\User\globalStorage\github.copilot-chat\copilotCli;C:\WINDOWS\system32;C:\WINDOWS;C:\WINDOWS\System32\Wbem;C:\WINDOWS\System32\WindowsPowerShell\v1.0\;C:\WINDOWS\System32\OpenSSH\;C:\Program Files\nodejs\;C:\Program Files\Go\bin;C:\Users\xbhag\.kimi-code\bin;C:\Users\xbhag\AppData\Local\Programs\Python\Python314\Scripts\;C:\Users\xbhag\AppData\Local\Programs\Python\Python314\;C:\Users\xbhag\AppData\Local\Programs\Python\Launcher\;C:\Users\xbhag\.cargo\bin;c:\Users\xbhag\AppData\Roaming\Code\User\globalStorage\github.copilot-chat\debugCommand;c:\Users\xbhag\AppData\Roaming\Code\User\globalStorage\github.copilot-chat\copilotCli;C:\WINDOWS\system32;C:\WINDOWS;C:\WINDOWS\System32\Wbem;C:\WINDOWS\System32\WindowsPowerShell\v1.0\;C:\WINDOWS\System32\OpenSSH\;C:\Program Files\nodejs\;C:\Program Files\Git\cmd;C:\WINDOWS\system32;C:\WINDOWS;C:\WINDOWS\System32\Wbem;C:\WINDOWS\System32\WindowsPowerShell\v1.0\;C:\WINDOWS\System32\OpenSSH\;C:\Program Files\nodejs\;C:\Program Files\Git\cmd;D:\orca\bin;c:\Users\xbhag\.vscode\extensions\ms-python.debugpy-2026.6.0-win32-x64\bundled\scripts\noConfigScripts;C:\Users\xbhag\.local\bin;;C:\Users\xbhag\.boxcode-cli\bin;C:\Users\xbhag\AppData\Local\Programs\Microsoft VS Code\bin;C:\Users\xbhag\AppData\Roaming\npm;C:\Users\xbhag\go\bin;C:\Users\xbhag\bin;C:\Users\xbhag\AppData\Local\momo += ";"

# Persist PATH (user scope)
 = [Environment]::GetEnvironmentVariable("Path", "User")
if ( -notlike "**") {
    [Environment]::SetEnvironmentVariable("Path", ";", "User")
    Write-Host "Added  to user PATH" -ForegroundColor Green
}

Write-Host "
momo installed to " -ForegroundColor Cyan
Write-Host "Run 'momo' or restart terminal to use." -ForegroundColor Cyan
