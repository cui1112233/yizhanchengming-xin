param(
  [string]$RepoRoot = '',
  [switch]$Apply
)

$ErrorActionPreference = 'Stop'
if ([string]::IsNullOrWhiteSpace($RepoRoot)) {
  $RepoRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
}
$root = [System.IO.Path]::GetFullPath($RepoRoot).TrimEnd('\')
if (-not (Test-Path -LiteralPath $root -PathType Container)) {
  throw "项目目录不存在：$root"
}

function Get-ArtifactBytes {
  param([string]$Path)
  if (Test-Path -LiteralPath $Path -PathType Leaf) {
    return (Get-Item -LiteralPath $Path).Length
  }
  $sum = (Get-ChildItem -LiteralPath $Path -Recurse -Force -File -ErrorAction SilentlyContinue | Measure-Object Length -Sum).Sum
  if ($null -eq $sum) { return 0 }
  return [int64]$sum
}

function Add-Candidate {
  param([System.Collections.Generic.List[object]]$List, [string]$Path, [string]$Reason)
  if (-not (Test-Path -LiteralPath $Path)) { return }
  $full = [System.IO.Path]::GetFullPath($Path)
  if (-not $full.StartsWith($root + '\', [StringComparison]::OrdinalIgnoreCase)) {
    throw "拒绝处理项目目录外路径：$full"
  }
  $List.Add([pscustomobject]@{
    path = $full
    bytes = Get-ArtifactBytes -Path $full
    reason = $Reason
    action = if ($Apply) { 'removed' } else { 'would_remove' }
  })
}

$candidates = [System.Collections.Generic.List[object]]::new()
Add-Candidate -List $candidates -Path (Join-Path $root 'executors\doubao-relay\target') -Reason 'Rust build cache'

$backend = Join-Path $root 'backend'
if (Test-Path -LiteralPath $backend -PathType Container) {
  $keep = @('qiantie-local.exe', 'qiantie-local.previous.exe', 'video-management-connector.exe', 'giant-material-plugin.exe')
  Get-ChildItem -LiteralPath $backend -File -Filter 'qiantie*.exe' |
    Where-Object { $_.Name -notin $keep } |
    ForEach-Object { Add-Candidate -List $candidates -Path $_.FullName -Reason 'superseded Go build' }
}

if ($Apply) {
  foreach ($candidate in $candidates) {
    Remove-Item -LiteralPath $candidate.path -Recurse -Force
  }
}

@($candidates) | ConvertTo-Json -Compress
