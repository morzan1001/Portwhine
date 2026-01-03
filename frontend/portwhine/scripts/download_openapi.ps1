# Downloads the OpenAPI spec from the running API server.
# Usage: .\scripts\download_openapi.ps1
#
# After downloading, run: .\fix_generated_models.ps1 -Regenerate

param(
    [string]$ApiUrl = "https://api.portwhine.local/openapi.json"
)

$OutputFile = Join-Path $PSScriptRoot "..\lib\api\swagger\portwhine.swagger"

Write-Host "Downloading OpenAPI spec from $ApiUrl..." -ForegroundColor Cyan
Invoke-WebRequest -Uri $ApiUrl -OutFile $OutputFile -SkipCertificateCheck
Write-Host "Saved to: $OutputFile" -ForegroundColor Green
Write-Host ""
Write-Host "Now run: .\fix_generated_models.ps1 -Regenerate" -ForegroundColor Yellow
