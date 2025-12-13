Set-Location "submodules/warframe-public-export"

Get-ChildItem -Recurse -Filter *.json | ForEach-Object {
    $fullPath = $_.FullName
    $rel = Resolve-Path -Relative $_.FullName
    $rel = $rel -replace '^[.\\/]+', ''
    $dest = Join-Path "..\..\assets" $rel
    $destDir = Split-Path $dest
    New-Item -ItemType Directory -Force -Path $destDir | Out-Null
    Copy-Item $fullPath $dest -Force
}
