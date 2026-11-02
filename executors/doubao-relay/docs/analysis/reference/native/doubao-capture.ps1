param([Parameter(Mandatory=$true)][string]$OutputDirectory,[string]$GenerationInputFile)
$ErrorActionPreference='Stop'
[Console]::OutputEncoding=New-Object System.Text.UTF8Encoding($false)
$OutputEncoding=[Console]::OutputEncoding
[Net.ServicePointManager]::SecurityProtocol=[Net.SecurityProtocolType]::Tls12
Add-Type -AssemblyName System.Drawing
Add-Type -TypeDefinition @'
using System;
using System.IO;
using System.Text;
using System.Net.WebSockets;
using System.Threading;
using System.Threading.Tasks;
using System.Collections.Concurrent;
public sealed class DoubaoCaptureSocket : IDisposable {
 public readonly ConcurrentQueue<string> Messages = new ConcurrentQueue<string>();
 private readonly ClientWebSocket socket = new ClientWebSocket();
 private readonly CancellationTokenSource cancel = new CancellationTokenSource();
 public void Connect(string url) {
  using (var timeout=new CancellationTokenSource(5000)) socket.ConnectAsync(new Uri(url),timeout.Token).GetAwaiter().GetResult();
  Task.Run((Func<Task>)Read);
 }
 public void Send(string json) {
  var data=Encoding.UTF8.GetBytes(json);
  using (var timeout=new CancellationTokenSource(5000)) socket.SendAsync(new ArraySegment<byte>(data),WebSocketMessageType.Text,true,timeout.Token).GetAwaiter().GetResult();
 }
 private async Task Read() {
  try {
   byte[] buffer=new byte[16384];
   while (!cancel.IsCancellationRequested) {
    using (var stream=new MemoryStream()) {
     WebSocketReceiveResult result;
     do {
      result=await socket.ReceiveAsync(new ArraySegment<byte>(buffer),cancel.Token);
      if(result.MessageType==WebSocketMessageType.Close)return;
      stream.Write(buffer,0,result.Count);
      if(stream.Length>16*1024*1024)throw new Exception("响应过大");
     } while(!result.EndOfMessage);
     if(Messages.Count>5000)throw new Exception("事件队列过大");
     Messages.Enqueue(Encoding.UTF8.GetString(stream.ToArray()));
    }
   }
  } catch(Exception) { Messages.Enqueue("{\"captureSocketClosed\":true}"); }
 }
 public void Dispose(){cancel.Cancel();socket.Dispose();cancel.Dispose();}
}
'@
$report=[ordered]@{schemaVersion=1;status='started';sessions=0;readySessions=0;generationRequests=0;responseBodies=0;originalImages=0;files=@();watermarkVerified=$false;historyRequests=0;chatConnections=0;pageReconnects=0;pageDiscoveryFailures=0;lastReadReason='';historyPolls=0;historyReadFailures=0;resultSource='';message=''}
$sessions=New-Object System.Collections.ArrayList
$script:originals=@{}
$script:videoPrimaryURL=$null
$videoNetworkCandidates=@{}
$nativeShadowEnabled=Test-Path -LiteralPath (Join-Path $PSScriptRoot '../native-shadow-bin/enabled')
$isVideo=$false
$generationRefused=$false
$lastUncertainReply=$null
$unverifiedRenderingSince=$null
if($GenerationInputFile){$captureInput=Get-Content -LiteralPath $GenerationInputFile -Raw -Encoding UTF8|ConvertFrom-Json;$isVideo=$captureInput.type -eq 'video'}
$parallelTracking=$captureInput.parallelTracking -and $captureInput.submissionMode -ne 'work'
$trackingKey=if($parallelTracking -and $captureInput.trackingKey -match '^shenlangDoubaoTask_[a-zA-Z0-9_]+$'){[string]$captureInput.trackingKey}else{$null}
$boundConversationId=$null;$boundSubmissionId=$null;$boundSubmissionCreatedAt=$null;$followupSubmissionIds=@();$boundRequest=$null;$backgroundTracking=$false;$handoffAnnounced=$false;$activeConversationId=$null
function Find-ConversationId($value){
 if($value -is [string]){if($value.TrimStart().StartsWith('{')){try{Find-ConversationId (ConvertFrom-Json $value)}catch{}};return}
 if($value -is [System.Collections.IEnumerable] -and $value -isnot [pscustomobject]){foreach($item in $value){Find-ConversationId $item};return}
 if($value -is [pscustomobject]){foreach($prop in $value.PSObject.Properties){if($prop.Name -eq 'conversation_id'){[string]$prop.Value}else{Find-ConversationId $prop.Value}}}
}
function Request-ConversationId($request){try{$ids=@(Find-ConversationId (ConvertFrom-Json $request.postData)|Select-Object -Unique);if($ids.Count -eq 1){return [string]$ids[0]}}catch{};return ''}
function Save-Report {$report | ConvertTo-Json -Depth 8 | Set-Content -Encoding UTF8 (Join-Path $OutputDirectory 'doubao-result-capture-report.json')}
function Parse-Originals($value,[int]$depth=0) {
 if ($isVideo) {return}
 if ($null -eq $value -or $depth -gt 50) {return}
 if ($value -is [string]) {
  if ($value.TrimStart().StartsWith('{') -or $value.TrimStart().StartsWith('[')) {
   try {$parsed=ConvertFrom-Json -InputObject $value;Parse-Originals $parsed ($depth+1)} catch {}
  };return
 }
 if ($value -is [System.Collections.IEnumerable] -and $value -isnot [pscustomobject]) {foreach($child in $value){Parse-Originals $child ($depth+1)};return}
 foreach($prop in $value.PSObject.Properties) {
  if ($prop.Name -eq 'image_ori_raw' -and $prop.Value.url) {
   $url=[Net.WebUtility]::HtmlDecode([string]$prop.Value.url)
   $uri=$null
   if ([uri]::TryCreate($url,[UriKind]::Absolute,[ref]$uri) -and $uri.Scheme -eq 'https' -and $uri.Host -match '(^|\.)(byteimg\.com|volces\.com)$') {$script:originals[$url]=$true}
  }
  Parse-Originals $prop.Value ($depth+1)
 }
}
function Parse-Body([string]$body) {
 if (!$body -or $body.Length -gt 16777216) {return}
 try {$data=$body | ConvertFrom-Json;Parse-Originals $data;return} catch {}
 foreach($line in ($body -split "`n")) {
  if ($line -match '^data:\s*(.+)') {try {$data=$Matches[1] | ConvertFrom-Json;Parse-Originals $data} catch {}}
 }
}
function Command($session,[string]$method,$params) {
 if($trackingKey -and $method -eq 'Runtime.evaluate' -and $params.expression){$params.expression=$params.expression.Replace('shenlangDoubaoGeneration',$trackingKey)}
 $session.nextId++
 $session.socket.Send((@{id=$session.nextId;method=$method;params=$params}|ConvertTo-Json -Depth 8 -Compress))
 return $session.nextId
}
try {
 $listeners=@(Get-NetTCPConnection -LocalPort 9707 -State Listen -ErrorAction SilentlyContinue)
 if (!$listeners.Count) {throw '本地接口未开启，请先运行豆包本地连接测试并保持豆包打开。'}
 foreach($l in $listeners) {
  if ($l.LocalAddress -notin @('127.0.0.1','::1') -or (Get-Process -Id $l.OwningProcess).ProcessName -ine 'Doubao') {throw '接口不属于本机豆包，已停止测试。'}
 }
 $targets=Invoke-RestMethod 'http://127.0.0.1:9707/json/list' -TimeoutSec 3
 foreach($target in $targets) {
  $u=[uri]$target.url
  if ($target.type -ne 'page' -or $u.Host -notin @('doubao-chat','doubao-background')) {continue}
  $ws=[uri]$target.webSocketDebuggerUrl
  if ($ws.Scheme -ne 'ws' -or $ws.Host -notin @('127.0.0.1','localhost','::1') -or $ws.Port -ne 9707) {throw '页面地址不符合本机连接范围。'}
  $socket=New-Object DoubaoCaptureSocket
  $session=@{connectedAt=(Get-Date);targetId=[string]$target.id;page=$u.Host;socket=$socket;nextId=0;requests=@{};videoRequests=@{};pending=@{};historyPending=@{};historyRequest=$null;readerPending=@{};ready=$false;closed=$false}
  [void]$sessions.Add($session)
  $socket.Connect($ws.AbsoluteUri)
  $session.enableId=Command $session 'Network.enable' @{maxTotalBufferSize=16777216;maxResourceBufferSize=8388608}
 }
 $report.sessions=$sessions.Count
 if (!$sessions.Count) {throw '未找到已经定位过的豆包聊天页面。'}
 $readyDeadline=(Get-Date).AddSeconds(8)
 do {
  foreach($s in $sessions) {
   $message=$null
   while($s.socket.Messages.TryDequeue([ref]$message)) {
    $event=$message | ConvertFrom-Json
    if ($event.id -eq $s.enableId -and !$event.error) {$s.ready=$true}
   }
  }
  Start-Sleep -Milliseconds 100
 } while (@($sessions | Where-Object {$_.ready}).Count -lt $sessions.Count -and (Get-Date) -lt $readyDeadline)
 $report.readySessions=@($sessions | Where-Object {$_.ready}).Count
 if (!$report.readySessions) {throw '页面连接可用，但未能开启生成结果观察。'}
 $report.status='observing';Save-Report
 [Console]::Out.WriteLine('DOUBAO_READY')
 Write-Host '捕获已准备好。现在请在桌面豆包新建对话，生成一张测试图片。'
 Write-Host '建议提示词：一只蓝色小鸟站在树枝上，柔和阳光，写实摄影。'
 Write-Host '请不要切换账号或浏览其他历史对话。图片接收最多等待 15 分钟。'
 # Generation may replace the chat target while preparing a model/new conversation.
 # Synchronize subscriptions after preparation and before the actual send.
 $captureCommand=Get-Command Command -CommandType Function
 $captureSessions=$sessions;$captureReport=$report
 function Sync-CaptureBeforeSend {
  $live=Invoke-RestMethod 'http://127.0.0.1:9707/json/list' -TimeoutSec 3
  $wanted=@($live|Where-Object {try{$_.type -eq 'page' -and ([uri]$_.url).Host -in @('doubao-chat','doubao-background')}catch{$false}})
  foreach($target in $wanted){
   if(@($captureSessions|Where-Object {!$_.closed -and $_.targetId -eq [string]$target.id}).Count){continue}
   $ws=[uri]$target.webSocketDebuggerUrl
   if($ws.Scheme -ne 'ws' -or $ws.Host -notin @('127.0.0.1','localhost','::1') -or $ws.Port -ne 9707){throw '提交前页面地址不符合本机连接范围。'}
   $newSocket=New-Object DoubaoCaptureSocket;$newSocket.Connect($ws.AbsoluteUri)
   $newSession=@{connectedAt=(Get-Date);targetId=[string]$target.id;page=([uri]$target.url).Host;socket=$newSocket;nextId=0;requests=@{};videoRequests=@{};pending=@{};historyPending=@{};historyRequest=$null;readerPending=@{};ready=$false;closed=$false}
   $newSession.enableId=& $captureCommand $newSession 'Network.enable' @{maxTotalBufferSize=16777216;maxResourceBufferSize=8388608;maxPostDataSize=16777216}
   [void]$captureSessions.Add($newSession)
  }
  $limit=(Get-Date).AddSeconds(8)
  do{
   foreach($s in @($captureSessions|Where-Object {!$_.ready -and !$_.closed})){
    $raw=$null
    while($s.socket.Messages.TryDequeue([ref]$raw)){
     $e=$raw|ConvertFrom-Json
     if($e.captureSocketClosed){$s.closed=$true}
     if($e.id -eq $s.enableId -and !$e.error){$s.ready=$true}
    }
   }
   $missing=@($wanted|Where-Object {$id=[string]$_.id;!@($captureSessions|Where-Object {$_.targetId -eq $id -and $_.ready -and !$_.closed}).Count})
   if(!$missing.Count){break};Start-Sleep -Milliseconds 25
  }while((Get-Date) -lt $limit)
  if(!$wanted.Count -or $missing.Count){throw '提交前结果观察未就绪，任务未发送。'}
  $captureReport['preSendObservationReady']=$true
 }
 if($GenerationInputFile){& (Join-Path $PSScriptRoot 'doubao-generate.ps1') -InputFile $GenerationInputFile -ReportFile (Join-Path $OutputDirectory 'submission.json')}
 $sendSince=0
 if($GenerationInputFile){
  $submission=ConvertFrom-Json ([IO.File]::ReadAllText((Join-Path $OutputDirectory 'submission.json'),[Text.Encoding]::UTF8))
  if($submission.submitted -and $submission.sendStartedAt){$sendSince=[long]$submission.sendStartedAt}
  $generationRequest=ConvertFrom-Json ([IO.File]::ReadAllText($GenerationInputFile,[Text.Encoding]::UTF8))
  if($generationRequest.submissionMode -eq 'work' -and $submission.submitted){
   $report.status='submitted_manual'
   $report.message='已提交至豆包，请自行核实生成结果。'
   Save-Report
   return
  }
 }
 $generationSource=[IO.File]::ReadAllText((Join-Path $PSScriptRoot 'doubao-generation-page.js'),[Text.Encoding]::UTF8)
 $feedbackBootstrap='if(typeof shenlangDoubaoGeneration === "undefined"){'+$generationSource+'};'
 $readerSource=[IO.File]::ReadAllText((Join-Path $PSScriptRoot 'doubao-result-reader.js'),[Text.Encoding]::UTF8)
 $nextHistoryPoll=Get-Date
 $nextPageCheck=Get-Date
 $nextFeedbackPoll=Get-Date
 $feedbackPending=@{}
 $identityPending=@{}
 $streamBindingPending=@{}
 $streamBindingBacklog=New-Object System.Collections.Generic.Queue[string]
 $streamBindingBytes=0
 $earlyBinding=$null
 $fallbackPending=@{}
 $fallbackAttempts=@{}
 $decoderSource=[IO.File]::ReadAllText((Join-Path $PSScriptRoot 'doubao-video-decoder.js'),[Text.Encoding]::UTF8)
 $downloadPending=@{}
 $downloadButtonTriggered=$false
 $nextDownloadAttempt=Get-Date
 $resultSeenAt=$null
 $deliveryWarningShown=$false
 $confirmPath=Join-Path $OutputDirectory 'confirm-video.json'
 $feedbackRequest=$null
 if($GenerationInputFile){$feedbackRequest=Get-Content -LiteralPath $GenerationInputFile -Raw -Encoding UTF8|ConvertFrom-Json;$feedbackRequest.prompt=[string]$feedbackRequest.prompt;$feedbackRequest|Add-Member -NotePropertyName submittedAt -NotePropertyValue $sendSince -Force}
 $historyTemplate=$null
 function Refresh-ChatConnection {
  try{$liveTargets=Invoke-RestMethod 'http://127.0.0.1:9707/json/list' -TimeoutSec 3}catch{$report.pageDiscoveryFailures++;return}
  $chatTargets=@($liveTargets|Where-Object {try{$t=[uri]$_.url;$_.type -eq 'page' -and $t.Host -eq 'doubao-chat'}catch{$false}})
  if($chatTargets.Count -ne 1){$report.lastReadReason='chat_page_not_unique';return}
  $target=$chatTargets[0]
  foreach($stale in @($sessions|Where-Object {!$_.closed -and !$_.ready -and ((Get-Date)-$_.connectedAt).TotalSeconds -gt 8})){$stale.closed=$true}
  $alive=@($sessions|Where-Object {!$_.closed -and $_.targetId -eq [string]$target.id})
  if($alive.Count){return}
  $ws=[uri]$target.webSocketDebuggerUrl
  if($ws.Scheme -ne 'ws' -or $ws.Host -notin @('127.0.0.1','localhost','::1') -or $ws.Port -ne 9707){$report.lastReadReason='invalid_local_page';return}
  # A replaced chat target must get its own Network subscription and reply reader.
  try{
   $socket=New-Object DoubaoCaptureSocket;$socket.Connect($ws.AbsoluteUri)
   $newSession=@{connectedAt=(Get-Date);targetId=[string]$target.id;page='doubao-chat';socket=$socket;nextId=0;requests=@{};videoRequests=@{};pending=@{};historyPending=@{};historyRequest=$historyTemplate;readerPending=@{};ready=$false;closed=$false}
   $newSession.enableId=Command $newSession 'Network.enable' @{maxTotalBufferSize=16777216;maxResourceBufferSize=8388608}
   [void]$sessions.Add($newSession);$report.pageReconnects++
  }catch{if($socket){$socket.Dispose()};$report.pageDiscoveryFailures++}
  foreach($old in @($sessions|Where-Object {$_.closed})){$old.socket.Dispose();[void]$sessions.Remove($old)}
 }
 $deadline=if($sendSince){[DateTimeOffset]::FromUnixTimeMilliseconds($sendSince).LocalDateTime.AddMinutes(15)}else{(Get-Date).AddMinutes(15)}
 while (($isVideo -or (Get-Date) -lt $deadline) -and !$script:originals.Count -and !$generationRefused) {
  if((Get-Date) -ge $nextPageCheck){Refresh-ChatConnection;$nextPageCheck=(Get-Date).AddSeconds(2)}
  foreach($s in $sessions) {
   $message=$null
   while($s.socket.Messages.TryDequeue([ref]$message)) {
    try {$event=$message | ConvertFrom-Json} catch {continue}
    if($event.id -eq $s.enableId -and !$event.error){$s.ready=$true;continue}
    if ($event.captureSocketClosed) {$s.closed=$true;continue}
    if ($event.method -eq 'Network.requestWillBeSent') {
     $request=$event.params.request
     if($parallelTracking -and @($event.params.initiator.stack.callFrames|Where-Object {$_.functionName -eq 'shenlangReadBoundConversation'}).Count){continue}
     if($isVideo -and $sendSince -and ([double]$event.params.wallTime*1000) -ge $sendSince){$s.videoRequests[[string]$event.params.requestId]=$true}
     $uri=$null
     if($sendSince -and $request.method -eq 'POST' -and [uri]::TryCreate($request.url,[UriKind]::Absolute,[ref]$uri) -and $uri.Scheme -eq 'https' -and $uri.Host -eq 'www.doubao.com' -and $uri.AbsolutePath -match '^/im/(chain/single|message/chain_by_block)/?$' -and $request.postData){
      $requestConversationId=Request-ConversationId $request
      if($s.page -eq 'doubao-chat' -and $requestConversationId){$activeConversationId=$requestConversationId}
      if($parallelTracking -and $boundConversationId -and $requestConversationId -ne $boundConversationId){continue}
      # chain_by_block can contain only uploaded references; prefer the full reply chain.
      if(!$s.historyRequest -or $uri.AbsolutePath -match '^/im/chain/single/?$'){
       $s.historyRequest=@{url=[string]$request.url;method='POST';postData=[string]$request.postData;headers=$request.headers}
      }
      if($s.page -eq 'doubao-chat'){$historyTemplate=$s.historyRequest;if($parallelTracking -and $boundConversationId){$boundRequest=$s.historyRequest}}
      $s.historyPending[[string]$event.params.requestId]=@{url=[string]$request.url;method='POST';postData=[string]$request.postData;headers=$request.headers}
      $report.historyRequests++
     }
     $uri=$null
     if ([uri]::TryCreate($request.url,[UriKind]::Absolute,[ref]$uri) -and $uri.Scheme -eq 'https' -and $uri.Host -match '(^|\.)(doubao\.com|snssdk\.com)$' -and $uri.AbsolutePath -match '(completion|generate|/im/message/send)' -and $uri.AbsolutePath -notmatch '(history|list|recommend|config)' -and $request.method -eq 'POST') {
      if (!$parallelTracking -and $s.requests.Count -and !$s.requests.ContainsKey($event.params.requestId) -and !$isVideo) {throw '同时发现多次生成，请一次只生成一张图片。'}
      if($parallelTracking -and !$boundConversationId -and !$submission.submissionIdentity -and $request.postData -and ([double]$event.params.wallTime*1000) -ge $sendSince){
       $identityRequest=@{url=[string]$request.url;method='POST';postData=[string]$request.postData}|ConvertTo-Json -Depth 12 -Compress
       $expression=$readerSource+';shenlangDoubaoResultReader.submissionIdentity('+$identityRequest+','+($feedbackRequest.prompt|ConvertTo-Json -Compress)+')'
       $identityId=Command $s 'Runtime.evaluate' @{expression=$expression;returnByValue=$true}
       $identityPending[$s.targetId+':'+[string]$identityId]=$s.page
       $report['identityRequestsObserved']=[int]$report['identityRequestsObserved']+1
      }
      $s.requests[$event.params.requestId]=$true
      $report.generationRequests++
     }
    } elseif ($isVideo -and $sendSince -and $event.method -eq 'Network.responseReceived' -and $event.params.response.mimeType -like 'video/*' -and $s.videoRequests.ContainsKey([string]$event.params.requestId)) {
     $uri=$null;$mediaURL=[string]$event.params.response.url
     if([uri]::TryCreate($mediaURL,[UriKind]::Absolute,[ref]$uri) -and $uri.Scheme -eq 'https' -and $uri.Host -match '(^|\.)(doubaovideo\.com|bytevcloudcn\.com|bytevcloud\.com|volces\.com|volcvideo\.com|bytevideo\.com|ixigua\.com|v[0-9]+-vdl\.doubao\.com)$'){$videoNetworkCandidates[$mediaURL]=$true}
    } elseif ($event.method -eq 'Network.eventSourceMessageReceived' -and $s.requests.ContainsKey($event.params.requestId)) {
     if(!$parallelTracking){Parse-Body $event.params.data}
     elseif(!$isVideo -and !$boundConversationId -and $streamBindingBacklog.Count -lt 64 -and ($streamBindingBytes+([string]$event.params.data).Length) -le 1048576){
      $streamBindingBacklog.Enqueue([string]$event.params.data);$streamBindingBytes+=([string]$event.params.data).Length
     }
    } elseif ($event.method -eq 'Network.loadingFinished' -and $s.historyPending.ContainsKey([string]$event.params.requestId)) {
     $requestForReply=$s.historyPending[[string]$event.params.requestId]
     [void]$s.historyPending.Remove([string]$event.params.requestId)
     $id=Command $s 'Network.getResponseBody' @{requestId=$event.params.requestId}
     $s.pending[[string]$id]=@{kind='history';request=$requestForReply}
    } elseif ($event.method -eq 'Network.loadingFinished' -and $s.requests.ContainsKey($event.params.requestId)) {
     $id=Command $s 'Network.getResponseBody' @{requestId=$event.params.requestId}
     $s.pending[[string]$id]=$true
    } elseif ($event.id -and $streamBindingPending.ContainsKey($s.targetId+':'+[string]$event.id)) {
     [void]$streamBindingPending.Remove($s.targetId+':'+[string]$event.id)
     if(!$event.error -and !$event.result.exceptionDetails -and $event.result.result.value){$earlyBinding=$event.result.result.value}
    } elseif ($event.id -and $identityPending.ContainsKey($s.targetId+':'+[string]$event.id)) {
     $identityPage=$identityPending[$s.targetId+':'+[string]$event.id];[void]$identityPending.Remove($s.targetId+':'+[string]$event.id)
     $identity=$event.result.result.value
     if(!$event.error -and !$event.result.exceptionDetails -and !$submission.submissionIdentity -and ($identity.localMessageId -or $identity.messageId)){
      $submission|Add-Member -NotePropertyName submissionIdentity -NotePropertyValue $identity -Force
      $submission|ConvertTo-Json -Depth 15|Set-Content -Encoding UTF8 (Join-Path $OutputDirectory 'submission.json')
      $report['identitySource']=$identityPage;$report['identityReadStatus']='matched';$nextHistoryPoll=Get-Date;Save-Report
     }else{$report['identityReadStatus']='request_identity_not_matched';Save-Report}
    } elseif ($event.id -and $downloadPending.ContainsKey($s.targetId+':'+[string]$event.id)) {
     [void]$downloadPending.Remove($s.targetId+':'+[string]$event.id)
     $report['downloadButtonCheck']=$event.result.result.value;
     if($event.result.result.value.hover){[void](Command $s 'Input.dispatchMouseEvent' @{type='mouseMoved';x=$event.result.result.value.hover.x;y=$event.result.result.value.hover.y})}
     if($event.result.result.value.clicked){$downloadButtonTriggered=$true;$report['downloadButtonTriggered']=$true;Save-Report}
    } elseif ($event.id -and $feedbackPending.ContainsKey($s.targetId+':'+[string]$event.id)) {
     [void]$feedbackPending.Remove($s.targetId+':'+[string]$event.id)
     $feedback=$event.result.result.value
     if($event.error -or $event.result.exceptionDetails){$report['feedbackReadError']=if($event.result.exceptionDetails){[string]$event.result.exceptionDetails.text}else{[string]$event.error.message};Save-Report}
     if(($event.error -or $event.result.exceptionDetails) -and $feedbackRequest.confirmationMode -eq 'auto' -and $report['feedbackStage'] -eq 'checking'){$feedbackRequest|Add-Member -NotePropertyName confirmationMode -NotePropertyValue 'manual' -Force;[Console]::Out.WriteLine('DOUBAO_AUTO_MANUAL')}
     if(($event.error -or $event.result.exceptionDetails) -and $feedbackRequest.autoUseMembership -eq $true -and $report['feedbackStage'] -eq 'quota_checking'){$feedbackRequest|Add-Member -NotePropertyName autoUseMembership -NotePropertyValue $false -Force;[Console]::Out.WriteLine('DOUBAO_MEMBERSHIP_MANUAL')}
     if($feedback.manualRequired){$feedbackRequest|Add-Member -NotePropertyName confirmationMode -NotePropertyValue 'manual' -Force;[Console]::Out.WriteLine('DOUBAO_AUTO_MANUAL')}
     if($feedback.diagnostic){$report['feedbackDiagnostic']=$feedback.diagnostic;Save-Report}
     # Allow transient render text to settle before treating it as an unexecuted request.
     if(!$event.error -and !$event.result.exceptionDetails){
      if($feedback.stage -eq 'rendering_unverified' -and !$script:originals.Count -and !$resultSeenAt -and $report['feedbackStage'] -notin @('generating','result')){
       if(!$unverifiedRenderingSince){$unverifiedRenderingSince=Get-Date}
       if(((Get-Date)-$unverifiedRenderingSince).TotalSeconds -ge 30){
        $feedback.stage='not_executed';$feedback.message='豆包仅返回“正在渲染，请稍候”，未确认有效的视频生成任务，请核对账号权限后重试。'
       }
      }elseif($feedback.stage -ne 'unknown'){$unverifiedRenderingSince=$null}
     }
     if($isVideo -and !$event.error -and !$event.result.exceptionDetails -and $feedback.stage -eq 'needs_check' -and $feedback.messageId -and $feedback.messageId -ne $lastUncertainReply){
      $lastUncertainReply=[string]$feedback.messageId
      [Console]::Out.WriteLine('DOUBAO_WAIT:'+(@{reason='feedback';message='豆包反馈需核实，请选择继续等待或取消等待';feedback=[string]$feedback.message}|ConvertTo-Json -Compress));Save-Report
     }
     if(!$event.error -and !$event.result.exceptionDetails -and $feedback.stage -in @('refused','failed','quota_exhausted','not_executed')){$generationRefused=$true;$report['refusalMessage']=[string]$feedback.message;[Console]::Out.WriteLine('DOUBAO_REFUSED:'+(@{message=$feedback.message;kind=$feedback.stage}|ConvertTo-Json -Compress));Save-Report}
     if(!$event.error -and !$event.result.exceptionDetails -and $feedback.stage -in @('checking','quota_checking','generating','result')){
      if($feedback.stage -eq 'result' -and !$resultSeenAt){$resultSeenAt=Get-Date}
      if($report['feedbackStage'] -ne $feedback.stage){$report['feedbackStage']=$feedback.stage;[Console]::Out.WriteLine('DOUBAO_STAGE:'+$feedback.stage);Save-Report}
      if($feedback.message){[Console]::Out.WriteLine('DOUBAO_FEEDBACK:'+(@{message=$feedback.message;model=$feedback.model;stage=$feedback.stage}|ConvertTo-Json -Compress))}
      if($feedback.sent){Remove-Item -LiteralPath $confirmPath -Force -ErrorAction SilentlyContinue;$submission|Add-Member -NotePropertyName followupSent -NotePropertyValue $true -Force;$submission|ConvertTo-Json -Depth 15|Set-Content -Encoding UTF8 (Join-Path $OutputDirectory 'submission.json');[Console]::Out.WriteLine('DOUBAO_STAGE:resent')}
     }
    } elseif ($event.id -and $fallbackPending.ContainsKey($s.targetId+':'+[string]$event.id)) {
     $vid=$fallbackPending[$s.targetId+':'+[string]$event.id];[void]$fallbackPending.Remove($s.targetId+':'+[string]$event.id)
     $decoded=$event.result.result.value
     if(!$event.error -and !$event.result.exceptionDetails -and $decoded.ok -and $decoded.url){$script:originals[[string]$decoded.url]=$true;$script:videoPrimaryURL=[string]$decoded.url;$report['videoId']=$vid;$report.resultSource='current_video_fallback_api'}else{$report['fallbackReadReason']=if($decoded.reason){$decoded.reason}else{'fallback_decode_failed'}}
    } elseif ($event.id -and $s.readerPending.ContainsKey([string]$event.id)) {
     $readerRequest=$s.readerPending[[string]$event.id];[void]$s.readerPending.Remove([string]$event.id)
     if($parallelTracking -and $boundConversationId -and (Request-ConversationId $readerRequest) -ne $boundConversationId){continue}
     $value=$event.result.result.value
     if($nativeShadowEnabled -and !$event.error -and !$event.result.exceptionDetails -and $value.nativeShadow){try{[Console]::Out.WriteLine('DOUBAO_NATIVE_SHADOW:'+($value.nativeShadow|ConvertTo-Json -Depth 8 -Compress))}catch{}}
     if($value.followupSubmissionIds){$followupSubmissionIds=@($value.followupSubmissionIds);$report['followupSubmissionIds']=$followupSubmissionIds}
     if($parallelTracking -and !$boundConversationId -and $value.binding.conversationId -and $value.binding.submissionId){
      $boundConversationId=[string]$value.binding.conversationId;$boundSubmissionId=[string]$value.binding.submissionId;$boundSubmissionCreatedAt=$value.binding.createdAt;$boundRequest=if($s.historyRequest -and (Request-ConversationId $s.historyRequest) -eq $boundConversationId -and ([uri]$s.historyRequest.url).AbsolutePath -match '/chain/single/?$'){$s.historyRequest}else{$readerRequest}
      $report['boundConversationId']=$boundConversationId;$report['boundSubmissionId']=$boundSubmissionId
      [Console]::Out.WriteLine('DOUBAO_BOUND:'+(@{conversationId=$boundConversationId;submissionId=$boundSubmissionId}|ConvertTo-Json -Compress));Save-Report
     }
     if($parallelTracking -and (!$boundConversationId -or !$boundSubmissionId)){continue}
     if($value.feedback -and !$value.images.Count){
      $report['messageFeedbackId']=[string]$value.feedback.messageId
      $feedbackPayload=@{boundSubmissionId=$boundSubmissionId;reply=$value.feedback;prompt=$feedbackRequest.prompt;modelName=$feedbackRequest.modelName;type=$(if($isVideo){'video'}else{'image'})}|ConvertTo-Json -Depth 12 -Compress
      $expression=$feedbackBootstrap+'typeof shenlangDoubaoGeneration === "undefined" ? ({stage:"unknown"}) : shenlangDoubaoGeneration.run("acceptFeedback",'+$feedbackPayload+')'
      $id=Command $s 'Runtime.evaluate' @{expression=$expression;returnByValue=$true};$feedbackPending[$s.targetId+':'+[string]$id]=$true
     }
     if($isVideo -and $value.fallbacks.Count){
      foreach($item in $value.fallbacks){
       $vid=[string]$item.vid
       if($fallbackAttempts.ContainsKey($vid) -and ((Get-Date)-$fallbackAttempts[$vid]).TotalSeconds -lt 10){continue}
       $fallbackAttempts[$vid]=Get-Date
       try{
        $uri=[uri]$item.api;if($uri.Scheme -ne 'https' -or $uri.Host -notmatch '^vas-[a-z0-9-]+\.snssdk\.com$' -or $uri.UserInfo){throw 'Invalid media API'}
        $response=Invoke-WebRequest -UseBasicParsing -Uri $uri.AbsoluteUri -Headers @{Accept='application/json,text/plain,*/*'} -TimeoutSec 10 -MaximumRedirection 0
        $body=[string]$response.Content;if($body.Length -gt 8388608){throw 'Media API response too large'}
        $expression=$decoderSource+';shenlangDoubaoVideoDecoder.decode(JSON.parse('+($body|ConvertTo-Json -Compress)+'),'+([string]$item.api|ConvertTo-Json -Compress)+').then(result=>({...result,url:shenlangDoubaoResultReader.unwatermarkedURL(result.url)}))'
        $id=Command $s 'Runtime.evaluate' @{expression=$expression;returnByValue=$true;awaitPromise=$true};$fallbackPending[$s.targetId+':'+[string]$id]=$vid;$report['fallbackRequested']=$true
       }catch{$report['fallbackReadReason']='fallback_request_failed'}
      }
     }
     if($isVideo -and $value.diagnostic){$report['withoutWatermarkDiagnostic']=$value.diagnostic}
     if($event.error -or $event.result.exceptionDetails -or !$value.ok){$report.historyReadFailures++;$report.lastReadReason=if($value.reason){[string]$value.reason}elseif($event.result.exceptionDetails){'page_reader_exception'}else{'page_reader_failed'}}
     elseif($value.images.Count){
      foreach($image in $value.images){if(!$isVideo -or $image.unwatermarked){$script:originals[[string]$image.url]=$true;if($isVideo){$script:videoPrimaryURL=[string]$image.url;$report['videoId']=[string]$image.vid}}}
      $report.resultSource=if($isVideo){'current_video_without_watermark'}else{'conversation_original'}
     }
    } elseif ($event.id -and $s.pending.ContainsKey([string]$event.id)) {
     $pending=$s.pending[[string]$event.id]
     [void]$s.pending.Remove([string]$event.id)
     if (!$event.error -and $event.result.body) {
      $body=$event.result.body
      if ($event.result.base64Encoded) {$body=[Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($body))}
      $report.responseBodies++
      if($pending -is [hashtable] -and $pending.kind -eq 'history'){
       $payload=$pending.request|ConvertTo-Json -Depth 12 -Compress
       $options=@{nativeShadow=$nativeShadowEnabled;followupSubmissionIds=$followupSubmissionIds;requireBinding=$parallelTracking;requireIdentity=$parallelTracking;expectedLocalMessageId=$submission.submissionIdentity.localMessageId;expectedMessageId=$submission.submissionIdentity.messageId;expectedConversationId=$submission.submissionIdentity.conversationId;boundSubmissionId=$boundSubmissionId;boundSubmissionCreatedAt=$boundSubmissionCreatedAt;boundConversationId=$boundConversationId;prompt=$feedbackRequest.prompt;since=$sendSince;type=$(if($isVideo){'video'}else{'image'})}|ConvertTo-Json -Compress
       $expression=$readerSource+';shenlangDoubaoResultReader.analyze('+($body|ConvertTo-Json -Compress)+','+$payload+','+$options+')'
       $id=Command $s 'Runtime.evaluate' @{expression=$expression;returnByValue=$true;awaitPromise=$true}
       $s.readerPending[[string]$id]=$pending.request
      }elseif(!$parallelTracking){Parse-Body $body}
     }
    }
   }
  }
  # Release image submission only after the exact sent message is echoed by
  # the server AND its read-only conversation request has been observed.
  # Keep early stream receipts until the asynchronous request identity is available.
  if($parallelTracking -and !$isVideo -and !$boundConversationId -and $submission.submissionIdentity -and $streamBindingBacklog.Count){
   $bindingSession=@($sessions|Where-Object {$_.ready -and !$_.closed -and $_.page -eq 'doubao-chat'})|Select-Object -First 1
   if($bindingSession){
    $bindingOptions=@{expectedLocalMessageId=$submission.submissionIdentity.localMessageId;expectedMessageId=$submission.submissionIdentity.messageId;expectedConversationId=$submission.submissionIdentity.conversationId;since=$sendSince}|ConvertTo-Json -Compress
    while($streamBindingBacklog.Count -and $streamBindingPending.Count -lt 8){
     $bindingData=$streamBindingBacklog.Dequeue();$streamBindingBytes-=$bindingData.Length
     $expression=$readerSource+';shenlangDoubaoResultReader.streamSubmissionBinding('+($bindingData|ConvertTo-Json -Compress)+','+$bindingOptions+')'
     $id=Command $bindingSession 'Runtime.evaluate' @{expression=$expression;returnByValue=$true};$streamBindingPending[$bindingSession.targetId+':'+[string]$id]=$true
    }
   }
  }
  if($parallelTracking -and !$isVideo -and !$boundConversationId -and $earlyBinding){
   $readerSession=@($sessions|Where-Object {$_.historyRequest -and (Request-ConversationId $_.historyRequest) -eq [string]$earlyBinding.conversationId})|Select-Object -First 1
   if($readerSession){
    $boundConversationId=[string]$earlyBinding.conversationId;$boundSubmissionId=[string]$earlyBinding.submissionId;$boundSubmissionCreatedAt=$earlyBinding.createdAt;$boundRequest=$readerSession.historyRequest
    $report['boundConversationId']=$boundConversationId;$report['boundSubmissionId']=$boundSubmissionId;$report['bindingSource']='verified_submission_stream'
    [Console]::Out.WriteLine('DOUBAO_BOUND:'+(@{conversationId=$boundConversationId;submissionId=$boundSubmissionId}|ConvertTo-Json -Compress));Save-Report
   }
  }
  if($parallelTracking -and $isVideo -and $sendSince -and !$report['feedbackStage'] -and !$report['feedbackFallbackOffered'] -and ([DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds()-$sendSince) -ge 60000){
   $report['feedbackFallbackOffered']=$true
   [Console]::Out.WriteLine('DOUBAO_WAIT:'+(@{reason='feedback';message='尚未识别到本次任务反馈，请核对豆包，可继续等待或取消等待。'}|ConvertTo-Json -Compress));Save-Report
  }
  if($feedbackRequest -and (Get-Date) -ge $nextFeedbackPoll){
   foreach($s in $sessions){if($s.page -ne 'doubao-chat' -or $s.closed -or !$s.ready){continue}
    if($backgroundTracking){
     if((Test-Path -LiteralPath $confirmPath) -and $activeConversationId -eq $boundConversationId){$backgroundTracking=$false;$handoffAnnounced=$false}
     else{if(Test-Path -LiteralPath $confirmPath){[Console]::Out.WriteLine('DOUBAO_NEEDS_PAGE')};continue}
    }
    if(@($feedbackPending.Keys|Where-Object {$_ -like ($s.targetId+':*')}).Count){continue}
    $payload=$feedbackRequest|ConvertTo-Json -Depth 12 -Compress
    $confirmed=Test-Path -LiteralPath $confirmPath
    if($confirmed){$feedbackRequest|Add-Member -NotePropertyName userConfirmed -NotePropertyValue $true -Force;$payload=$feedbackRequest|ConvertTo-Json -Depth 12 -Compress}
    if($report['feedbackStage'] -eq 'result' -and !$downloadButtonTriggered -and (Get-Date) -ge $nextDownloadAttempt -and !$downloadPending.Count){
     $expr='shenlangDoubaoGeneration.run("requestVideoDownload",'+$payload+')'
     $id=Command $s 'Runtime.evaluate' @{expression=$expr;returnByValue=$true};$downloadPending[$s.targetId+':'+[string]$id]=$true;$nextDownloadAttempt=(Get-Date).AddSeconds(5)
    }
    $feedbackStep=if(!$isVideo){'terminalFeedback'}elseif($confirmed -or $feedbackRequest.confirmationMode -eq 'auto' -or $feedbackRequest.autoUseMembership -eq $true){'videoFollowup'}else{'videoFeedback'}
    $expression=$feedbackBootstrap+'typeof shenlangDoubaoGeneration === "undefined" ? ({stage:"unknown"}) : Promise.resolve(shenlangDoubaoGeneration.run('+($feedbackStep|ConvertTo-Json -Compress)+','+$payload+')).then(value=>({...value,diagnostic:shenlangDoubaoGeneration.run("feedbackDiagnostic",'+$payload+')}))'
    try{$id=Command $s 'Runtime.evaluate' @{expression=$expression;returnByValue=$true;awaitPromise=$true};$feedbackPending[$s.targetId+':'+[string]$id]=$true}catch{}
   }
   $nextFeedbackPoll=(Get-Date).AddMilliseconds(500)
  }
  $report.chatConnections=@($sessions|Where-Object {!$_.closed -and $_.ready -and $_.page -eq 'doubao-chat'}).Count
  if($sendSince -and !$script:originals.Count -and (Get-Date) -ge $nextHistoryPoll){
   foreach($s in $sessions){
    if($s.page -ne 'doubao-chat' -or $s.closed -or !$s.ready){continue}
    if($parallelTracking -and $boundRequest){$s.historyRequest=$boundRequest}
    if(!$s.historyRequest -and $historyTemplate){$s.historyRequest=$historyTemplate}
    if(!$s.historyRequest){$report.lastReadReason='no_conversation_request';continue}
    if($s.readerPending.Count){$report.lastReadReason='waiting_for_reader';continue}
    # Replay only the observed, read-only conversation request. Never submit generation.
    $payload=$s.historyRequest|ConvertTo-Json -Depth 12 -Compress
    $options=@{nativeShadow=$nativeShadowEnabled;followupSubmissionIds=$followupSubmissionIds;requireBinding=$parallelTracking;requireIdentity=$parallelTracking;expectedLocalMessageId=$submission.submissionIdentity.localMessageId;expectedMessageId=$submission.submissionIdentity.messageId;expectedConversationId=$submission.submissionIdentity.conversationId;boundSubmissionId=$boundSubmissionId;boundSubmissionCreatedAt=$boundSubmissionCreatedAt;boundConversationId=$boundConversationId;prompt=$feedbackRequest.prompt;since=$sendSince;type=$(if($isVideo){'video'}else{'image'})}|ConvertTo-Json -Compress
    $expression=$readerSource+';shenlangDoubaoResultReader.read('+$payload+','+$options+')'
    try{
     $id=Command $s 'Runtime.evaluate' @{expression=$expression;returnByValue=$true;awaitPromise=$true}
     $s.readerPending[[string]$id]=$s.historyRequest;$report.historyPolls++
    }catch{$s.closed=$true;$report.lastReadReason='chat_transport_closed'}
   }
   $nextHistoryPoll=(Get-Date).AddSeconds(1);Save-Report
  }
  if(!$parallelTracking -and $isVideo -and $downloadButtonTriggered -and !$script:originals.Count){$marked=@($videoNetworkCandidates.Keys|Where-Object {$_ -match '[?&]lr=unwatermarked(?:&|$)'});if($marked.Count -eq 1){$script:originals[$marked[0]]=$true;$script:videoPrimaryURL=$marked[0];$report.resultSource='current_result_download_button'}}
  if($isVideo -and !$deliveryWarningShown -and $resultSeenAt -and !$script:originals.Count -and ((Get-Date)-$resultSeenAt).TotalSeconds -gt 90){
   $report['downloadReadReason']=$report['lastReadReason'];$report['lastReadReason']='video_delivered_url_not_found';Save-Report
   $report|ConvertTo-Json -Depth 8|Set-Content -Encoding UTF8 (Join-Path ([Environment]::GetFolderPath('Desktop')) '豆包视频交付诊断.json')
   $deliveryWarningShown=$true
   [Console]::Out.WriteLine('DOUBAO_WAIT:'+(@{message='视频已生成，下载地址尚未接收，请选择继续接收或取消等待后自行核对。'}|ConvertTo-Json -Compress))
  }
  if($parallelTracking -and !$handoffAnnounced -and $boundConversationId -and $boundSubmissionId -and (!$isVideo -or $report['feedbackStage'] -eq 'generating')){
   $backgroundTracking=$true;$handoffAnnounced=$true;$report['backgroundTracking']=$true
   [Console]::Out.WriteLine('DOUBAO_BACKGROUND:'+(@{conversationId=$boundConversationId;submissionId=$boundSubmissionId}|ConvertTo-Json -Compress));Save-Report
  }
  Start-Sleep -Milliseconds 120
 }
 $report.originalImages=$script:originals.Count
 if($generationRefused){$report.status='generation_refused';$report.message='豆包未生成：'+$report['refusalMessage']}elseif ($script:originals.Count) {
  [Console]::Out.WriteLine('DOUBAO_STAGE:downloading')
  if(!$isVideo){[Console]::Out.WriteLine('DOUBAO_IMAGES:'+([Math]::Min($script:originals.Count,20)))}
  if($isVideo){
   foreach($candidate in $script:originals.Keys){if($candidate -notmatch '[?&]lr=unwatermarked(?:&|$)'){throw '未收到无水印视频地址。'}}
   $report['watermarkSource']='server_unwatermarked';
   $folder=Join-Path $OutputDirectory 'videos';[void](New-Item -ItemType Directory -Path $folder -Force)
   $file=Join-Path $folder 'video_1.mp4'
   $url=if($script:videoPrimaryURL){$script:videoPrimaryURL}else{@($script:originals.Keys)[0]}
   Invoke-WebRequest -UseBasicParsing -Uri $url -Headers @{Referer='https://www.doubao.com/'} -OutFile $file -TimeoutSec 180 -MaximumRedirection 0
   if((Get-Item $file).Length -le 12 -or (Get-Item $file).Length -gt 1073741824){Remove-Item $file;throw '视频文件大小无效。'}
   $stream=[IO.File]::OpenRead($file);try{$header=New-Object byte[] 12;[void]$stream.Read($header,0,12);if([Text.Encoding]::ASCII.GetString($header,4,4) -ne 'ftyp'){throw '结果不是有效MP4视频。'}}finally{$stream.Dispose()}
   $report.files+=@{file='video_1.mp4';bytes=(Get-Item $file).Length}
   $report.status='video_downloaded';$report.message='本次视频的无水印地址文件已下载。'
  }else{
  $folder=Join-Path $OutputDirectory 'images'
  [void](New-Item -ItemType Directory -Path $folder)
  $i=0
  foreach($url in ($script:originals.Keys | Select-Object -First 20)) {
   $i++;$file=Join-Path $folder ('image_'+$i+'.png')
   Invoke-WebRequest -UseBasicParsing -Uri $url -OutFile $file -TimeoutSec 40 -MaximumRedirection 0
   if ((Get-Item $file).Length -gt 104857600) {Remove-Item $file;throw '原图超过保存上限。'}
   $im=[Drawing.Image]::FromFile($file)
   try {$width=$im.Width;$height=$im.Height;$extension=if($im.RawFormat.Guid -eq [Drawing.Imaging.ImageFormat]::Jpeg.Guid){'.jpg'}elseif($im.RawFormat.Guid -eq [Drawing.Imaging.ImageFormat]::Png.Guid){'.png'}else{throw '原图格式暂不支持。'}} finally {$im.Dispose()}
   $correct=Join-Path $folder ('image_'+$i+$extension);if($correct -ne $file){Move-Item $file $correct;$file=$correct}
   $report.files+=@{file=[IO.Path]::GetFileName($file);width=$width;height=$height;bytes=(Get-Item $file).Length}
  }
  $report.status='original_downloaded_needs_visual_check'
  $report.message='已从本次生成响应读取原图字段并下载。请检查右下角水印，尚未确认无水印。'

  }
 } else {
  $report.status='no_original_captured'
  $report.message='任务已发送，但未确认本次生成的原图。请保留豆包结果并检查本地连接，勿重复生成。'
 }
} catch {$report.status='failed';$report.message=if($report['lastReadReason'] -eq 'video_delivered_url_not_found'){'豆包已交付，但无水印下载地址接收失败；桌面已保存豆包视频交付诊断.json，请保留结果勿重复生成。'}else{'捕获未完成：请检查桌面豆包及本地连接是否正常后重试。'}}
finally {foreach($s in $sessions){if($trackingKey -and !$s.closed -and $s.page -eq 'doubao-chat'){try{[void](Command $s 'Runtime.evaluate' @{expression=('delete globalThis.'+$trackingKey);returnByValue=$true})}catch{}};$s.socket.Dispose()}}
Save-Report
Write-Host $report.message
Write-Host '请反馈 doubao-result-capture-report.json，以及成功下载时的图片文件。'

