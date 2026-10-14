[CmdletBinding()]
param(
  [string[]]$ConfigRoots = @('deploy'),
  [switch]$Strict
)

$ErrorActionPreference = 'Stop'
$repo = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$findings = [System.Collections.Generic.List[object]]::new()

function Add-Finding([string]$Kind, [string]$Path, [string]$Detail, [bool]$Blocking) {
  $findings.Add([pscustomobject]@{
      kind = $Kind
      path = $Path
      detail = $Detail
      blocking = $Blocking
    })
}

# A Node executable belonging to Codex/Adobe is not this application's runtime.
# Only report a process when its command line points into this repository or
# explicitly names one of the retired application services.
try {
  Get-CimInstance Win32_Process -ErrorAction Stop |
    Where-Object { $_.Name -match '^(node|npm|npx)(\.exe)?$' } |
    ForEach-Object {
      $command = [string]$_.CommandLine
      $applicationProcess = ($command -match [regex]::Escape($repo)) -and
        ($command -match '(?i)server\.js|local-executor|services[\\/]121-browser-worker')
      if ($applicationProcess) {
        Add-Finding 'process' $_.ProcessId $command $true
      }
    }
} catch {
  Add-Finding 'process-scan' 'Win32_Process' $_.Exception.Message $false
}

foreach ($root in $ConfigRoots) {
  $resolved = Join-Path $repo $root
  if (-not (Test-Path $resolved)) { continue }
  Get-ChildItem -LiteralPath $resolved -Recurse -File -Include '*.yml','*.yaml','*.conf' |
    ForEach-Object {
      $text = Get-Content -LiteralPath $_.FullName -Raw
      $relative = $_.FullName.Substring($repo.Length).TrimStart('\','/')
      if ($_.Extension -in @('.yml','.yaml')) {
        $nodeLines = ($text -split "`r?`n") | Where-Object {
          $_ -match '(?i)^\s*(image|command|entrypoint|build|profiles|volumes|v88-node:|browser-worker:|shuihuo-compat:)' -and
            $_ -match '(?i)node|npm|playwright'
        }
        foreach ($line in $nodeLines) {
          # Profiles marked node-legacy are retained rollback evidence, not a
          # default runtime. They remain reported but non-blocking.
          $legacy = $text -match 'profiles:\s*\[\s*["'']node-legacy["'']\s*\]' -or
            $relative -match '(?i)(^|[\\/])v78-public([\\/]|$)'
          Add-Finding 'compose-node-reference' $relative $line.Trim() (-not $legacy)
        }
      }
      if ($_.Extension -eq '.conf') {
        foreach ($line in (($text -split "`r?`n") | Where-Object { $_ -match '(?i)proxy_pass\s+http://(?!go-api)' })) {
          $legacy = $relative -match '(?i)(^|[\\/])v78-public([\\/]|$)'
          Add-Finding 'nginx-node-upstream' $relative $line.Trim() (-not $legacy)
        }
      }
    }
}

$blocking = @($findings | Where-Object { $_.blocking })
$report = [pscustomobject]@{
  runtime = 'go'
  repository = $repo
  blockingCount = $blocking.Count
  findings = @($findings)
}
$report | ConvertTo-Json -Depth 6
if ($Strict -and $blocking.Count -gt 0) { exit 1 }
exit 0
