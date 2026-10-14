$dir = "c:\Proyectos\USACH\TESIS\TeamEdu\front-TeamEdu\content\compatibility"
$files = Get-ChildItem "$dir\*.js" | Where-Object { $_.Name -ne "createCompatibility.js" -and $_.Name -ne "index.js" }

foreach ($f in $files) {
    $content = [System.IO.File]::ReadAllText($f.FullName, [System.Text.Encoding]::UTF8)
    
    # Skip if already using the factory
    if ($content -match 'createCompatibility') {
        Write-Host "Already updated: $($f.Name)"
        continue
    }
    
    # Strip "export default " prefix and trailing newline
    $prefix = "export default "
    $body = $content.TrimStart()
    if ($body.StartsWith($prefix)) {
        $body = $body.Substring($prefix.Length).TrimEnd()
    }
    
    $header = "import { createCompatibility } from './createCompatibility.js'" + "`r`n" + "`r`n"
    $newContent = $header + "export default createCompatibility(" + $body + ")`r`n"
    
    [System.IO.File]::WriteAllText($f.FullName, $newContent, [System.Text.Encoding]::UTF8)
    Write-Host "Updated: $($f.Name)"
}
Write-Host "Done."
