# Build script for S1Monitor on Windows
$ErrorActionPreference = "Stop"

# Define variables
$BinaryName = "s1monitor.exe"
$MainPath = ".\cmd\s1monitor"
$BinDir = ".\bin"

# Function to display usage
function Show-Usage {
    Write-Host "Build script for S1Monitor"
    Write-Host ""
    Write-Host "Usage:"
    Write-Host "  .\build.ps1 [command]"
    Write-Host ""
    Write-Host "Commands:"
    Write-Host "  build      Build the application"
    Write-Host "  build-all  Build for all platforms"
    Write-Host "  clean      Clean build artifacts"
    Write-Host "  run        Build and run the application"
    Write-Host "  daemon     Build and run in daemon mode"
    Write-Host "  tidy       Run go mod tidy"
}

# Function to build the application
function Build {
    Write-Host "Building S1Monitor..."
    go build -v -o $BinaryName $MainPath
    Write-Host "Build complete: $BinaryName"
}

# Function to build for all platforms
function Build-All {
    Write-Host "Building for all platforms..."
    
    # Create bin directory if it doesn't exist
    if (-not (Test-Path $BinDir)) {
        New-Item -ItemType Directory -Path $BinDir | Out-Null
    }
    
    # Windows amd64
    Write-Host "Building for Windows amd64..."
    $env:GOOS = "windows"; $env:GOARCH = "amd64"
    go build -v -o "$BinDir\s1monitor_windows_amd64.exe" $MainPath
    
    # Linux amd64
    Write-Host "Building for Linux amd64..."
    $env:GOOS = "linux"; $env:GOARCH = "amd64"
    go build -v -o "$BinDir\s1monitor_linux_amd64" $MainPath
    
    # macOS amd64
    Write-Host "Building for macOS amd64..."
    $env:GOOS = "darwin"; $env:GOARCH = "amd64"
    go build -v -o "$BinDir\s1monitor_darwin_amd64" $MainPath
    
    # Reset environment variables
    $env:GOOS = ""
    $env:GOARCH = ""
    
    Write-Host "Build complete. Binaries available in $BinDir"
}

# Function to clean build artifacts
function Clean {
    Write-Host "Cleaning build artifacts..."
    if (Test-Path $BinaryName) { Remove-Item $BinaryName -Force }
    if (Test-Path $BinDir) { Remove-Item $BinDir -Recurse -Force }
    Write-Host "Clean complete."
}

# Function to run the application
function Run {
    Build
    Write-Host "Running S1Monitor..."
    & ".\$BinaryName"
}

# Function to run in daemon mode
function Run-Daemon {
    Build
    Write-Host "Running S1Monitor in daemon mode..."
    & ".\$BinaryName" -d
}

# Function to run go mod tidy
function Tidy {
    Write-Host "Running go mod tidy..."
    go mod tidy
    Write-Host "Done."
}

# Main script execution
$command = $args[0]
if ($null -eq $command) {
    Show-Usage
    exit 0
}

switch ($command) {
    "build" { Build }
    "build-all" { Build-All }
    "clean" { Clean }
    "run" { Run }
    "daemon" { Run-Daemon }
    "tidy" { Tidy }
    default {
        Write-Host "Unknown command: $command"
        Show-Usage
        exit 1
    }
}
