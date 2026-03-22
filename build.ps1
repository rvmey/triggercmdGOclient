$oses = @("windows", "darwin", "linux")
$arches = @("amd64", "386", "arm", "arm7")
foreach ($os in $oses) {
    foreach ($arch in $arches) {
        $goarch = $arch
        $goarm = $null
        if ($arch -eq "arm7") {
            $goarch = "arm"
            $goarm = "7"
        } elseif ($arch -eq "arm") {
            $goarm = "6"
        }

        $isArmVariant = ($arch -eq "arm" -or $arch -eq "arm7")
        if (($isArmVariant -and $os -eq "linux") -or (-not $isArmVariant)) {
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
            $env:GOARCH = $goarch
            if ($goarm) { $env:GOARM = $goarm } else { Remove-Item Env:GOARM -ErrorAction SilentlyContinue }
            go build -v -o binaries\tcmd-$os-$arch

            # Rename to .exe if it's windows
            if ($os -eq "windows") {
                mv binaries\tcmd-$os-$arch binaries\tcmd-$os-$arch.exe
            }
        }        
    }
}
Remove-Item Env:GOARM -ErrorAction SilentlyContinue
$signtool = "C:\Program Files (x86)\Windows Kits\10\bin\10.0.26100.0\x64\signtool.exe"
$signArgs = @("/sha1", "4514d960dbb0b24571b642c38cfa439b67e7c738", "/tr", "http://timestamp.sectigo.com", "/td", "sha256", "/fd", "sha256", "/n", "VanderMey Consulting LLC")
$windowsExes = Get-ChildItem -Path .\binaries\tcmd-windows-*.exe
foreach ($exe in $windowsExes) {
    & $signtool sign @signArgs $exe.FullName
}

Copy-Item .\binaries\tcmd-windows-amd64.exe .\binaries\tcmd.exe
Copy-Item .\binaries\tcmd-windows-386.exe .\binaries\tcmd_32bit.exe
