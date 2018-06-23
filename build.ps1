$oses = @("windows", "darwin", "linux")
$arches = @("amd64", "386", "arm")
foreach ($os in $oses) {
    foreach ($arch in $arches) {
        if (($arch -eq "arm" -and $os -eq "linux") -or ($arch -ne "arm")) {
            # Delete any existing binary
            if ($os -eq "windows") {
                if (Test-Path binaries\tcmd-$os-$arch.exe -PathType Leaf) {
                    rm binaries\tcmd-$os-$arch.exe
                }
            } else {
                if (Test-Path binaries\tcmd-$os-$arch -PathType Leaf) {
                    rm binaries\tcmd-$os-$arch
                }
            }

            # Build the new binary
            $env:GOOS = $os
            $env:GOARCH = $arch
            go build -v -o binaries\tcmd-$os-$arch

            # Rename to .exe if it's windows
            if ($os -eq "windows") {
                mv binaries\tcmd-$os-$arch binaries\tcmd-$os-$arch.exe
            }
        }        
    }
}
