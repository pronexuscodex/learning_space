# Install the Systems & AI Academy on Windows (no administrator rights needed).
#
#   irm https://raw.githubusercontent.com/pronexuscodex/learning_space/main/install.ps1 | iex
#
# It downloads the latest release (or $env:ACADEMY_VERSION = 'vX.Y.Z'),
# checks it against the release's SHA256SUMS, and installs:
#
#   %LocalAppData%\Programs\Academy\academy.exe   the program
#   Start Menu > Systems & AI Academy             a shortcut with the icon
#   your user PATH                                so 'academy' works in any terminal
#   Settings > Apps > Installed apps              where it can be uninstalled
#
# Your progress is kept in %LocalAppData%\academy and your PDFs in
# Documents\Academy Library, so reinstalling or upgrading never touches them.
#
# Upgrade: run the same command again.
# Remove:  Settings > Apps, or
#          & ([scriptblock]::Create((irm https://raw.githubusercontent.com/pronexuscodex/learning_space/main/install.ps1))) -Uninstall

param(
    [switch]$Uninstall,
    [string]$Version = $env:ACADEMY_VERSION
)

$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue' # Invoke-WebRequest is far faster without the progress bar
[Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12

$Repo = 'pronexuscodex/learning_space'
$AppName = 'Systems & AI Academy'
$InstallDir = Join-Path $env:LOCALAPPDATA 'Programs\Academy'
$Exe = Join-Path $InstallDir 'academy.exe'
$Shortcut = Join-Path ([Environment]::GetFolderPath('Programs')) "$AppName.lnk"
$UninstallKey = 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\SystemsAIAcademy'

# The uninstaller is written next to the program, so Settings > Apps can run it.
$UninstallScript = @'
$ErrorActionPreference = 'Continue'
$InstallDir = '__DIR__'
$Exe = Join-Path $InstallDir 'academy.exe'
$where = $null
if (Test-Path $Exe) { $where = & $Exe -where 2>$null | Where-Object { $_ -notlike 'Program:*' } }
Remove-Item -LiteralPath '__LNK__' -Force -ErrorAction SilentlyContinue
Remove-Item -LiteralPath 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\SystemsAIAcademy' -Recurse -Force -ErrorAction SilentlyContinue
$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if ($userPath) {
    $kept = ($userPath -split ';' | Where-Object { $_ -and ($_.TrimEnd('\') -ne $InstallDir.TrimEnd('\')) }) -join ';'
    if ($kept -ne $userPath) { [Environment]::SetEnvironmentVariable('Path', $kept, 'User') }
}
# Only the installer's own files: an older academy kept progress, backups
# and PDFs in this folder, and those must survive an uninstall.
foreach ($f in 'academy.exe', 'academy.exe.old', 'README.md', 'CHANGELOG.md', 'LICENSE', 'uninstall.ps1') {
    Remove-Item -LiteralPath (Join-Path $InstallDir $f) -Force -ErrorAction SilentlyContinue
}
if (-not (Get-ChildItem -LiteralPath $InstallDir -Force -ErrorAction SilentlyContinue)) {
    Remove-Item -LiteralPath $InstallDir -Force -ErrorAction SilentlyContinue
} else {
    Write-Host "Kept your files in $InstallDir"
}
Write-Host 'Removed Systems & AI Academy.'
Write-Host 'Your progress and PDFs were kept; delete these folders yourself if you no longer want them:'
if ($where) { $where | ForEach-Object { Write-Host $_ } }
else { Write-Host "  $env:LOCALAPPDATA\academy and Documents\Academy Library" }
'@
$UninstallScript = $UninstallScript.Replace('__DIR__', $InstallDir.Replace("'", "''")).Replace('__LNK__', $Shortcut.Replace("'", "''"))

if ($Uninstall) {
    $u = Join-Path $InstallDir 'uninstall.ps1'
    if (Test-Path $u) {
        & $u
    } else {
        # Nothing registered: run the same steps from memory.
        & ([scriptblock]::Create($UninstallScript))
    }
    return
}

# The native processor, even when this PowerShell runs under emulation.
$native = (Get-ItemProperty 'HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment').PROCESSOR_ARCHITECTURE
switch ($native) {
    'AMD64' { $arch = 'amd64' }
    'ARM64' { $arch = 'arm64' }
    default { throw "academy install: unsupported processor '$native' (64-bit Intel/AMD or ARM is needed)" }
}

function Get-LatestTag {
    # The latest-release page redirects to its tag; no API token or rate limit.
    try {
        $req = [Net.WebRequest]::Create("https://github.com/$Repo/releases/latest")
        $req.Method = 'HEAD'
        $resp = $req.GetResponse()
        $tag = $resp.ResponseUri.AbsoluteUri.Split('/')[-1]
        $resp.Close()
        return $tag
    } catch {
        return ''
    }
}

function Test-ReleaseFiles([string]$Tag) {
    # The release's SHA256SUMS is attached.
    try {
        $req = [Net.WebRequest]::Create("https://github.com/$Repo/releases/download/$Tag/SHA256SUMS")
        $req.Method = 'HEAD'
        $resp = $req.GetResponse()
        $resp.Close()
        return $true
    } catch {
        return $false
    }
}

function Get-NewestGoodTag {
    # The newest release with a proper vX.Y.Z tag and its files attached.
    try {
        $releases = Invoke-RestMethod -UseBasicParsing "https://api.github.com/repos/$Repo/releases?per_page=20"
    } catch {
        return ''
    }
    foreach ($r in $releases) {
        if ($r.draft -or $r.prerelease -or $r.tag_name -notmatch '^v\d+\.\d+\.\d+$') { continue }
        if ($r.assets | Where-Object { $_.name -eq 'SHA256SUMS' }) { return $r.tag_name }
    }
    return ''
}

if ($Version) {
    if ($Version -notmatch '^v\d+\.\d+\.\d+$') {
        throw "academy install: '$Version' is not a release version; set `$env:ACADEMY_VERSION = 'vX.Y.Z', or leave it unset for the latest"
    }
} else {
    $Version = Get-LatestTag
    # The newest release may still be building, or may have been published
    # with a mistyped tag: then install the newest one that is complete.
    if ($Version -notmatch '^v\d+\.\d+\.\d+$' -or -not (Test-ReleaseFiles $Version)) {
        $good = Get-NewestGoodTag
        if (-not $good) {
            throw 'academy install: could not find a release to install; check your connection and try again in a few minutes, or set $env:ACADEMY_VERSION'
        }
        if ($Version -and $good -ne $Version) {
            Write-Host "The newest release ($Version) is not ready to install; installing $good instead."
        }
        $Version = $good
    }
}

$name = "academy-$Version-windows-$arch"
$base = "https://github.com/$Repo/releases/download/$Version"
$tmp = Join-Path ([IO.Path]::GetTempPath()) ("academy-install-" + [Guid]::NewGuid())
New-Item -ItemType Directory -Path $tmp | Out-Null
try {
    Write-Host "Downloading $AppName $Version for windows/$arch..."
    $zip = Join-Path $tmp "$name.zip"
    $sums = Join-Path $tmp 'SHA256SUMS'
    # A new release gets its files a few minutes after it is published.
    try {
        Invoke-WebRequest -UseBasicParsing "$base/SHA256SUMS" -OutFile $sums
    } catch {
        if ($env:ACADEMY_VERSION) { throw "academy install: there is no release $Version (see https://github.com/$Repo/releases)" }
        throw "academy install: $Version has no files yet (it may have just been published); try again in a few minutes"
    }
    Invoke-WebRequest -UseBasicParsing "$base/$name.zip" -OutFile $zip

    $want = $null
    foreach ($line in Get-Content $sums) {
        $f = $line -split '\s+'
        if ($f.Count -ge 2 -and $f[1] -eq "$name.zip") { $want = $f[0].ToLower() }
    }
    if (-not $want) { throw "academy install: $name.zip is not listed in SHA256SUMS" }
    $got = (Get-FileHash -Algorithm SHA256 $zip).Hash.ToLower()
    if ($got -ne $want) { throw "academy install: checksum mismatch for $name.zip (expected $want, got $got); nothing was installed" }
    Write-Host 'Checksum verified.'

    Expand-Archive -LiteralPath $zip -DestinationPath $tmp -Force
    $newExe = Join-Path $tmp "$name\academy.exe"
    if (-not (Test-Path $newExe)) { throw 'academy install: the archive does not contain academy.exe' }

    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    # A running academy.exe cannot be overwritten, but it can be renamed.
    $old = "$Exe.old"
    Remove-Item -LiteralPath $old -Force -ErrorAction SilentlyContinue
    if (Test-Path $Exe) { Move-Item -LiteralPath $Exe -Destination $old -Force }
    Copy-Item -LiteralPath $newExe -Destination $Exe -Force
    Remove-Item -LiteralPath $old -Force -ErrorAction SilentlyContinue
    foreach ($doc in 'README.md', 'CHANGELOG.md', 'LICENSE') {
        $src = Join-Path $tmp "$name\$doc"
        if (Test-Path $src) { Copy-Item -LiteralPath $src -Destination $InstallDir -Force }
    }
    Get-ChildItem -LiteralPath $InstallDir | Unblock-File
    Write-Host "Installed $Exe"
} finally {
    Remove-Item -LiteralPath $tmp -Recurse -Force -ErrorAction SilentlyContinue
}

# Start Menu shortcut, with the icon built into academy.exe. It opens in
# your default terminal (Windows Terminal or the console).
$shell = New-Object -ComObject WScript.Shell
$lnk = $shell.CreateShortcut($Shortcut)
$lnk.TargetPath = $Exe
$lnk.WorkingDirectory = $env:USERPROFILE
$lnk.IconLocation = "$Exe,0"
$lnk.Description = 'Computer science, AI and security, stage by stage'
$lnk.Save()
Write-Host "Added $AppName to the Start Menu"

# User PATH, so 'academy' works in every new terminal.
$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
$parts = @()
if ($userPath) { $parts = $userPath -split ';' | Where-Object { $_ } }
if (-not ($parts | Where-Object { $_.TrimEnd('\') -eq $InstallDir.TrimEnd('\') })) {
    [Environment]::SetEnvironmentVariable('Path', (($parts + $InstallDir) -join ';'), 'User')
    Write-Host "Added $InstallDir to your PATH (new terminals will see it)"
}
if (-not (($env:Path -split ';') -contains $InstallDir)) { $env:Path = "$env:Path;$InstallDir" }

# Settings > Apps entry.
$uninstallPs1 = Join-Path $InstallDir 'uninstall.ps1'
Set-Content -LiteralPath $uninstallPs1 -Value $UninstallScript -Encoding UTF8
$sizeKB = [int]((Get-ChildItem -LiteralPath $InstallDir | Measure-Object -Property Length -Sum).Sum / 1KB)
$command = "powershell.exe -NoProfile -ExecutionPolicy Bypass -File `"$uninstallPs1`""
New-Item -Path $UninstallKey -Force | Out-Null
$values = @{
    DisplayName          = $AppName
    DisplayVersion       = $Version.TrimStart('v')
    Publisher            = 'pronexuscodex'
    DisplayIcon          = "$Exe,0"
    InstallLocation      = $InstallDir
    UninstallString      = $command
    QuietUninstallString = $command
    URLInfoAbout         = "https://github.com/$Repo"
    HelpLink             = "https://github.com/$Repo#readme"
}
foreach ($k in $values.Keys) { New-ItemProperty -Path $UninstallKey -Name $k -Value $values[$k] -PropertyType String -Force | Out-Null }
foreach ($k in 'NoModify', 'NoRepair') { New-ItemProperty -Path $UninstallKey -Name $k -Value 1 -PropertyType DWord -Force | Out-Null }
New-ItemProperty -Path $UninstallKey -Name EstimatedSize -Value $sizeKB -PropertyType DWord -Force | Out-Null

Write-Host ''
Write-Host "Done. Open $AppName from the Start Menu, or type:  academy"
Write-Host 'See where your progress and PDFs are kept:  academy -where'
Write-Host 'Upgrade later by running the same install command again.'
