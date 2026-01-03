# Dart API Client Build Script
# 
# The swagger_dart_code_generator has a bug where it generates "ListType" instead of "List<Type>"
# This script handles the workaround.
#
# Usage:
#   .\fix_generated_models.ps1              - Normal build (swagger disabled, just json_serializable)
#   .\fix_generated_models.ps1 -Regenerate  - Full regeneration from OpenAPI spec (enables swagger temporarily)

param(
    [switch]$Regenerate
)

$modelsFile = "lib/api/generated/portwhine.models.swagger.dart"
$buildYaml = "build.yaml"

function Fix-GeneratedModels {
    if (Test-Path $modelsFile) {
        $content = Get-Content $modelsFile -Raw
        $originalContent = $content
        
        # Fix all occurrences of "ListSomething" to "List<Something>"
        $content = $content -replace '\bList([A-Z][a-zA-Z0-9]*)\b(?!\s*<)', 'List<$1>'
        
        if ($content -ne $originalContent) {
            Set-Content $modelsFile -Value $content -NoNewline
            Write-Host "  Fixed List types in generated models!"
            return $true
        } else {
            Write-Host "  No fixes needed."
            return $false
        }
    }
    return $false
}

function Enable-SwaggerGenerator {
    $content = Get-Content $buildYaml -Raw
    $content = $content -replace 'enabled: false', 'enabled: true'
    Set-Content $buildYaml -Value $content -NoNewline
}

function Disable-SwaggerGenerator {
    $content = Get-Content $buildYaml -Raw
    $content = $content -replace 'enabled: true', 'enabled: false'
    Set-Content $buildYaml -Value $content -NoNewline
}

Write-Host "=== Dart API Client Build ==="
Write-Host ""

if ($Regenerate) {
    Write-Host "Regenerating API client from OpenAPI spec..."
    Write-Host ""
    
    # Clean everything
    Write-Host "Step 1: Cleaning..."
    dart run build_runner clean 2>&1 | Out-Null
    Remove-Item -Path ".dart_tool/build" -Recurse -Force -ErrorAction SilentlyContinue
    Remove-Item -Path "lib/api/generated/*.dart" -Force -ErrorAction SilentlyContinue
    
    # Enable swagger generator temporarily
    Write-Host "Step 2: Enabling swagger generator..."
    Enable-SwaggerGenerator
    
    # Generate swagger files (will fail at json_serializable due to ListType bug)
    Write-Host "Step 3: Generating swagger models..."
    dart run build_runner build --delete-conflicting-outputs 2>&1 | Out-Null
    
    # Apply fixes
    Write-Host "Step 4: Applying List<Type> fixes..."
    Fix-GeneratedModels
    
    # Disable swagger generator again
    Write-Host "Step 5: Disabling swagger generator..."
    Disable-SwaggerGenerator
    
    # Clean build cache and run again
    Write-Host "Step 6: Cleaning build cache..."
    Remove-Item -Path ".dart_tool/build" -Recurse -Force -ErrorAction SilentlyContinue
    
    Write-Host "Step 7: Running json_serializable..."
    dart run build_runner build --delete-conflicting-outputs
    
    Write-Host ""
    Write-Host "Done! API client regenerated."
    
} else {
    # Normal build - swagger is disabled, just run build_runner
    Write-Host "Running build_runner..."
    dart run build_runner build --delete-conflicting-outputs
}
