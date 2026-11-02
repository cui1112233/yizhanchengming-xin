param(
  [string]$ConfiguredProxy = '',
  [string]$FallbackProxy = 'http://127.0.0.1:7890'
)

$configured = $ConfiguredProxy.Trim()
if ($configured) {
  return $configured
}

$client = $null
try {
  $proxyUri = [Uri]$FallbackProxy
  if (-not $proxyUri.Host -or $proxyUri.Port -le 0) {
    return ''
  }
  $client = [Net.Sockets.TcpClient]::new()
  $connect = $client.ConnectAsync($proxyUri.Host, $proxyUri.Port)
  if ($connect.Wait(300) -and $client.Connected) {
    return $FallbackProxy
  }
} catch {
  return ''
} finally {
  if ($client) { $client.Dispose() }
}

return ''
