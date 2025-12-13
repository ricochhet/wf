param (
    [string]$SubmodulePath
)

if (-not $SubmodulePath) {
    Write-Host "Usage: .\script.ps1 <path/to/submodule>"
    exit 1
}

git submodule deinit -f $SubmodulePath
Remove-Item -Recurse -Force ".git/modules/$SubmodulePath"
git config --remove-section submodule.$SubmodulePath
git rm -f $SubmodulePath
