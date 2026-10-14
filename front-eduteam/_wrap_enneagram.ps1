$dir = "c:\Proyectos\USACH\TESIS\TeamEdu\front-TeamEdu\content\enneagram"
$files = Get-ChildItem "$dir\type-*.js"

foreach ($f in $files) {
    $content = [System.IO.File]::ReadAllText($f.FullName, [System.Text.Encoding]::UTF8)
    
    # Skip if already using the factory
    if ($content -match 'createEnneagramType') {
        Write-Host "Already updated: $($f.Name)"
        continue
    }
    
    # Strip "export default " prefix and trailing newline
    $prefix = "export default "
    $body = $content.TrimStart()
    if ($body.StartsWith($prefix)) {
        $body = $body.Substring($prefix.Length).TrimEnd()
    }
    
    $header = "import { createEnneagramType } from './createEnneagramType.js'" + "`r`n" + "`r`n"
    $newContent = $header + "export default createEnneagramType(" + $body + ")`r`n"
    
    [System.IO.File]::WriteAllText($f.FullName, $newContent, [System.Text.Encoding]::UTF8)
    Write-Host "Updated: $($f.Name)"
}
Write-Host "Done."
