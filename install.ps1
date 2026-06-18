# Rune installer for Windows
# Usage: irm https://raw.githubusercontent.com/rune-context/rune/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

$repo = "rune-context/rune"
$binary = "rune"

function Main {
    $os = "windows"
    $arch = Get-Arch

    Write-Host "Rune installer"
    Write-Host "  OS:   $os"
    Write-Host "  Arch: $arch"
    Write-Host ""

    $url = "https://github.com/$repo/releases/latest/download/$binary-$os-$arch.zip"
    Write-Host "Downloading: $url"

    $tmp = New-TemporaryFile | ForEach-Object { Remove-Item $_; New-Item -ItemType Directory -Path $_ }
    $archive = Join-Path $tmp "rune.zip"

    try {
        Invoke-WebRequest -Uri $url -OutFile $archive -UseBasicParsing
        Expand-Archive -Path $archive -DestinationPath $tmp -Force

        $installDir = Join-Path $env:USERPROFILE "bin"
        if (-not (Test-Path $installDir)) {
            New-Item -ItemType Directory -Path $installDir -Force | Out-Null
        }

        $src = Join-Path $tmp "$binary.exe"
        $dst = Join-Path $installDir "$binary.exe"
        Copy-Item -Path $src -Destination $dst -Force

        # Add to PATH if not already there
        $userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
        if ($userPath -notlike "*$installDir*") {
            [Environment]::SetEnvironmentVariable("PATH", "$userPath;$installDir", "User")
            Write-Host "Added $installDir to PATH"
        }

        Write-Host ""
        Write-Host "Installed to: $dst"
        Write-Host ""
        Write-Host "Get started:"
        Write-Host "  cd your-project"
        Write-Host "  rune init"
        Write-Host "  rune index"
    }
    finally {
        Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
    }
}

function Get-Arch {
    $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
    switch ($arch) {
        "X64"   { return "amd64" }
        "Arm64" { return "arm64" }
        default { throw "Unsupported architecture: $arch" }
    }
}

Main
