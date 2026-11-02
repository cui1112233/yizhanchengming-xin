param([string]$InputFile,[string]$OutputFile)
$request=Get-Content -LiteralPath $InputFile -Raw -Encoding UTF8|ConvertFrom-Json
$Scenario='normal'
$ErrorActionPreference='Stop'
$report=[ordered]@{schemaVersion=11;scenario=$Scenario;status='started';noForegroundCalls=$true;noMouseActions=$true;noAccountSwitch=$true;accounts=@();sources=@();completeListVerified=$false;diagnostics=@{};message=''}
Add-Type -TypeDefinition @'
using System;
using System.IO;
using System.Text;
using System.Net.WebSockets;
using System.Threading;
using System.Threading.Tasks;
using System.Collections.Concurrent;
public sealed class AccountReadSocket : IDisposable {
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
function Add-Accounts($items,[string]$source){
 $count=0
 foreach($item in @($items)){
  $label=[string]$item.label
  if($label -notmatch '^avatar\s+(.+)$' -or $label.Length -gt 300){continue}
  $name=($Matches[1] -split '\s+(体验套餐|标准套餐|加强套餐|高级套餐|专业套餐|基础套餐|(?<=\s)[\u4e00-\u9fffA-Za-z0-9]{1,24}套餐)')[0].Trim()
  if(!$name -or $name.Length -gt 120 -or $name -in @('豆包订阅','豆包订阅升级','免费领额度','免费领取额度')){continue}
  $subscription=if($label -match '\s(体验套餐|标准套餐|加强套餐|高级套餐|专业套餐|基础套餐|(?<=\s)[\u4e00-\u9fffA-Za-z0-9]{1,24}套餐)(?:\s|$)'){$Matches[1]}else{''}
  $count++
  if(!@($report.accounts|Where-Object {$_.label -ceq $label}).Count){$report.accounts+=@{name=$name;subscription=$subscription;label=$label}}
 }
 return $count
}
$script:pageCommandSequence=0
function Evaluate($socket,[string]$expression){
 $script:pageCommandSequence++;$commandId=$script:pageCommandSequence
 $socket.Send((@{id=$commandId;method='Runtime.evaluate';params=@{expression=$expression;returnByValue=$true;awaitPromise=$false;silent=$true}}|ConvertTo-Json -Depth 8 -Compress))
 $deadline=[DateTime]::UtcNow.AddSeconds(5)
 while([DateTime]::UtcNow -lt $deadline){
  $message=$null
  while($socket.Messages.TryDequeue([ref]$message)){
   $data=$message|ConvertFrom-Json
   if($data.id -eq $commandId){if($data.error -or $data.result.exceptionDetails){throw '页面读取失败'};return $data.result.result.value}
  }
  Start-Sleep -Milliseconds 25
 }
 throw '页面读取超时'
}
function Set-PageMetrics($socket,$viewport,[bool]$clear){
 $script:pageCommandSequence++;$commandId=$script:pageCommandSequence
 $method=if($clear){'Emulation.clearDeviceMetricsOverride'}else{'Emulation.setDeviceMetricsOverride'}
 $params=if($clear){@{}}else{@{width=[int]$viewport.width;height=[int]$viewport.height;deviceScaleFactor=[double]$viewport.scale;mobile=$false}}
 $socket.Send((@{id=$commandId;method=$method;params=$params}|ConvertTo-Json -Depth 8 -Compress))
 $deadline=[DateTime]::UtcNow.AddSeconds(2)
 while([DateTime]::UtcNow -lt $deadline){
  $message=$null
  while($socket.Messages.TryDequeue([ref]$message)){$data=$message|ConvertFrom-Json;if($data.id -eq $commandId){return !$data.error}}
  Start-Sleep -Milliseconds 25
 }
 return $false
}
function Resume-Page($socket){
 $script:pageCommandSequence++;$commandId=$script:pageCommandSequence
 $socket.Send((@{id=$commandId;method='Page.setWebLifecycleState';params=@{state='active'}}|ConvertTo-Json -Depth 8 -Compress))
 $deadline=[DateTime]::UtcNow.AddSeconds(2)
 while([DateTime]::UtcNow -lt $deadline){
  $message=$null
  while($socket.Messages.TryDequeue([ref]$message)){$data=$message|ConvertFrom-Json;if($data.id -eq $commandId){return !$data.error}}
  Start-Sleep -Milliseconds 25
 }
 return $false
}
function Set-PageFocus($socket,[bool]$enabled){
 $script:pageCommandSequence++;$commandId=$script:pageCommandSequence
 $socket.Send((@{id=$commandId;method='Emulation.setFocusEmulationEnabled';params=@{enabled=$enabled}}|ConvertTo-Json -Depth 8 -Compress))
 $deadline=[DateTime]::UtcNow.AddSeconds(2)
 while([DateTime]::UtcNow -lt $deadline){
  $message=$null
  while($socket.Messages.TryDequeue([ref]$message)){$data=$message|ConvertFrom-Json;if($data.id -eq $commandId){if($data.error){return $false};return $true}}
  Start-Sleep -Milliseconds 25
 }
 return $false
}
function Send-PageMouse($socket,[string]$type,[double]$x,[double]$y){
 $script:pageCommandSequence++;$commandId=$script:pageCommandSequence
 $socket.Send((@{id=$commandId;method='Input.dispatchMouseEvent';params=@{type=$type;x=$x;y=$y;button='left';clickCount=1}}|ConvertTo-Json -Depth 8 -Compress))
 $deadline=[DateTime]::UtcNow.AddSeconds(5)
 while([DateTime]::UtcNow -lt $deadline){
  $message=$null
  while($socket.Messages.TryDequeue([ref]$message)){$data=$message|ConvertFrom-Json;if($data.id -eq $commandId){if($data.error){throw '豆包页面未接受切换操作。'};return}}
  Start-Sleep -Milliseconds 25
 }
 throw '豆包页面切换操作超时。'
}
Add-Type -TypeDefinition @'
using System;
using System.Runtime.InteropServices;
public static class BackgroundFocusRead {
 [DllImport("user32.dll")] public static extern IntPtr GetForegroundWindow();
}
'@
function Test-DoubaoConnection {
 $listeners=@(Get-NetTCPConnection -LocalPort 9707 -State Listen -ErrorAction SilentlyContinue)
 if(!$listeners.Count){return $false}
 foreach($listener in $listeners){if($listener.LocalAddress -notin @('127.0.0.1','::1') -or (Get-Process -Id $listener.OwningProcess -ErrorAction SilentlyContinue).ProcessName -ine 'Doubao'){return $false}}
 try{$targets=Invoke-RestMethod 'http://127.0.0.1:9707/json/list' -TimeoutSec 2;return @($targets|Where-Object {try{$u=[uri]$_.url;$_.type -eq 'page' -and $u.Scheme -in @('doubao','chrome') -and $u.Host -eq 'doubao-chat'}catch{$false}}).Count -gt 0}catch{return $false}
}
function Ensure-DoubaoStarted {
 $running=@(Get-Process -Name Doubao -ErrorAction SilentlyContinue)
 if($running.Count -and (Test-DoubaoConnection)){$report.startup.action='already_connected';return}
 # Allow an app already starting to finish before deciding it needs a restart.
 if($running.Count){for($attempt=0;$attempt -lt 3;$attempt++){Start-Sleep -Milliseconds $(if($request.operation -eq 'switch'){100}else{400});if(Test-DoubaoConnection){$report.startup.action='already_connected';return}}}
 $report.startup.action='locating_installation'
 $paths=@();foreach($process in $running){try{if($process.Path){$paths+=,$process.Path}}catch{}}

 foreach($key in @('HKCU:\Software\Microsoft\Windows\CurrentVersion\App Paths\Doubao.exe','HKLM:\Software\Microsoft\Windows\CurrentVersion\App Paths\Doubao.exe')){if(Test-Path $key){$paths+=,(Get-Item $key).GetValue('')}}
 foreach($key in @('HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\*','HKLM:\Software\Microsoft\Windows\CurrentVersion\Uninstall\*','HKLM:\Software\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\*')){foreach($item in @(Get-ItemProperty $key -ErrorAction SilentlyContinue|Where-Object {$_.DisplayName -match '豆包|Doubao'})){if($item.InstallLocation){$paths+=,(Join-Path $item.InstallLocation 'Doubao.exe');$paths+=,(Join-Path $item.InstallLocation 'app\Doubao.exe')};if($item.DisplayIcon){$paths+=,($item.DisplayIcon -replace ',\d+$','').Trim('"')}}}
 $paths+=,@("$env:LOCALAPPDATA\Doubao\app\Doubao.exe","$env:LOCALAPPDATA\Programs\Doubao\Doubao.exe","$env:ProgramFiles\Doubao\Doubao.exe")
 foreach($drive in @(Get-PSDrive -PSProvider FileSystem)){$paths+=,(Join-Path $drive.Root 'Doubao\app\Doubao.exe')}
 $exe=@($paths|Where-Object {$_ -and (Split-Path $_ -Leaf) -ieq 'Doubao.exe' -and (Test-Path -LiteralPath $_ -PathType Leaf)}|Select-Object -Unique)
 if(!$exe.Count){$report.startup.error='installation_not_found';throw '未找到豆包安装程序'}
 # The user requested reconnect and sync as one action. Restart only this installation.
 $running=@(Get-Process -Name Doubao -ErrorAction SilentlyContinue)
 if($running.Count -and (Test-DoubaoConnection)){$report.startup.action='already_connected';return}
 if($running.Count){
  $report.startup.action='restarting'
  foreach($process in $running){try{if($process.Path -ine $exe[0]){$report.startup.error='restart_failed';throw '无法唯一确认豆包进程'}}catch{$report.startup.error='restart_failed';throw}}
  foreach($process in $running){if($process.MainWindowHandle -ne 0){[void]$process.CloseMainWindow()}}
  Start-Sleep -Milliseconds 800
  foreach($process in @(Get-Process -Name Doubao -ErrorAction SilentlyContinue)){
   try{if($process.Path -ine $exe[0]){throw '豆包进程路径发生变化'};Stop-Process -Id $process.Id -ErrorAction Stop}catch{$report.startup.error='restart_failed';throw}
  }
  $exitDeadline=[DateTime]::UtcNow.AddSeconds(5)
  while(@(Get-Process -Name Doubao -ErrorAction SilentlyContinue).Count -and [DateTime]::UtcNow -lt $exitDeadline){Start-Sleep -Milliseconds 200}
  if(@(Get-Process -Name Doubao -ErrorAction SilentlyContinue).Count){$report.startup.error='restart_failed';throw '豆包未完成退出'}
 }
 $listeners=@(Get-NetTCPConnection -LocalPort 9707 -State Listen -ErrorAction SilentlyContinue)
 if($listeners.Count){$report.startup.error='port_occupied';throw '本地连接端口已被占用'}
 $report.startup.action='starting'
 try{Start-Process -FilePath $exe[0] -ArgumentList @('--remote-debugging-port=9707','--remote-debugging-address=127.0.0.1','--force-renderer-accessibility=complete') -ErrorAction Stop|Out-Null}catch{$report.startup.error='launch_failed';throw}
 $deadline=[DateTime]::UtcNow.AddSeconds(20)
 do{
  try{
   $targets=Invoke-RestMethod 'http://127.0.0.1:9707/json/list' -TimeoutSec 1
   $pages=@($targets|Where-Object {try{$u=[uri]$_.url;$_.type -eq 'page' -and $u.Scheme -in @('doubao','chrome') -and $u.Host -eq 'doubao-chat'}catch{$false}})
   if($pages.Count -and (Test-DoubaoConnection)){$report.startup.action='started';Start-Sleep -Milliseconds 1000;return}
  }catch{}
  Start-Sleep -Milliseconds 250
 }while([DateTime]::UtcNow -lt $deadline)
 $report.startup.error='startup_timeout';throw '等待豆包启动超时'
}
$report.Add('startup',@{action='checking';error=''})
$initialForeground=[BackgroundFocusRead]::GetForegroundWindow()
$report.Add('menuAttempts',@())
$report.Add('foregroundChanged',$false)
try{
 if($request.operation -eq 'connect'){
  Ensure-DoubaoStarted
  @{ok=$true;connected=$true;syncMode='background';message='豆包已连接，请点击同步账号。';diagnostics=@{status='connected';startup=$report.startup}}|ConvertTo-Json -Depth 5|Set-Content -LiteralPath $OutputFile -Encoding UTF8
  exit 0
 }
 if(!(Test-DoubaoConnection)){$report.startup.error='connection_required';throw '请先连接豆包建立连接'}
 # Opening a stopped app may activate its window. Background-reading checks start after launch.
 $initialForeground=[BackgroundFocusRead]::GetForegroundWindow()
 $windows=@()
 if($request.operation -ne 'switch'){
 Add-Type -AssemblyName UIAutomationClient
 Add-Type -AssemblyName UIAutomationTypes
 $windows=@(Get-Process -Name Doubao -ErrorAction SilentlyContinue|Where-Object {$_.MainWindowHandle -ne 0})
 foreach($p in $windows){
  try{
   $w=[Windows.Automation.AutomationElement]::FromHandle($p.MainWindowHandle)
   $buttons=$w.FindAll([Windows.Automation.TreeScope]::Descendants,(New-Object Windows.Automation.PropertyCondition([Windows.Automation.AutomationElement]::ControlTypeProperty,[Windows.Automation.ControlType]::Button)))
   $rect=$w.Current.BoundingRectangle
   $validWindow=(!$rect.IsEmpty -and $rect.Width -gt 300 -and $rect.Height -gt 200 -and $rect.Left -gt -30000 -and $rect.Top -gt -30000)
   $region=@($buttons|Where-Object {$c=$_.Current;$b=$c.BoundingRectangle;$validWindow -and !$c.IsPassword -and !$b.IsEmpty -and $b.Height -ge 20 -and $b.Height -le 85 -and $b.Width -ge 100 -and $b.Width -le 360 -and $b.Left -ge ($rect.Left-2) -and $b.Left -lt ($rect.Left+80) -and $b.Top -gt ($rect.Bottom-100) -and $b.Bottom -le ($rect.Bottom+2)})
   $candidates=@($region|Where-Object {$_.Current.IsEnabled -and ![string]::IsNullOrWhiteSpace($_.Current.Name)})
   $report.menuAttempts+=@{stage='locate';candidateCount=$candidates.Count;windowUsable=$validWindow;windowRect=@{left=$rect.Left;top=$rect.Top;width=$rect.Width;height=$rect.Height};regionButtons=@($region|Select-Object -First 10|ForEach-Object {$c=$_.Current;$b=$c.BoundingRectangle;@{name=$c.Name;automationId=$c.AutomationId;offscreen=$c.IsOffscreen;enabled=$c.IsEnabled;left=$b.Left;top=$b.Top;width=$b.Width;height=$b.Height}});sidebarControls=@($buttons|Where-Object {$_.Current.Name -match '^(展开侧边栏|收起侧边栏|展开左侧栏|收起左侧栏|展开导航|展开菜单)$'}|ForEach-Object {@{name=$_.Current.Name;automationId=$_.Current.AutomationId}})}
   $invoked=$false
   if($candidates.Count -eq 1){
    $e=$candidates[0];$supported=@();$patternErrors=@()
    foreach($patternType in @([Windows.Automation.InvokePattern]::Pattern,[Windows.Automation.ExpandCollapsePattern]::Pattern,[Windows.Automation.SelectionItemPattern]::Pattern)){
     try{[object]$pattern=$null;if($e.TryGetCurrentPattern($patternType,[ref]$pattern)){$supported+=$patternType.ProgrammaticName}}catch{$patternErrors+=@{type=$_.Exception.GetType().FullName;message=$_.Exception.Message}}
    }
    $report.menuAttempts+=@{stage='capabilities_only';supportedPatterns=$supported;errors=$patternErrors;action='none'}
   }
   $items=@();foreach($e in $w.FindAll([Windows.Automation.TreeScope]::Descendants,(New-Object Windows.Automation.PropertyCondition([Windows.Automation.AutomationElement]::ControlTypeProperty,[Windows.Automation.ControlType]::MenuItem)))){
    if(!$e.Current.IsPassword -and $e.Current.Name -match '^avatar\s+'){$items+=@{label=$e.Current.Name}}
   }
   $count=Add-Accounts $items 'uia';$report.sources+=@{kind='uia';status='read';accountEntries=$count}
  }catch{$report.sources+=@{kind='uia';status='unavailable';accountEntries=0;errorType=$_.Exception.GetType().FullName;errorMessage=$_.Exception.Message}}
 }
 if(!$windows.Count){$report.sources+=@{kind='uia';status='no_window';accountEntries=0}}
 }
 $processes=@(Get-Process -ErrorAction SilentlyContinue|Where-Object {$_.ProcessName -match '(?i)doubao'})
 $report.diagnostics=@{processCount=$processes.Count;processes=@($processes|ForEach-Object {@{name=$_.ProcessName;hasMainWindow=($_.MainWindowHandle -ne 0)}});listenerCheck='started';httpCheck='not_tested'}
 $listeners=@();try{$listeners=@(Get-NetTCPConnection -LocalPort 9707 -State Listen -ErrorAction Stop);$report.diagnostics.listenerCheck=if($listeners.Count){'found'}else{'none'}}catch{$report.diagnostics.listenerCheck='unavailable_or_no_listener'}
 $valid=$true

 foreach($listener in $listeners){if($listener.LocalAddress -notin @('127.0.0.1','::1') -or (Get-Process -Id $listener.OwningProcess -ErrorAction SilentlyContinue).ProcessName -ine 'Doubao'){$valid=$false}}
 if($valid){
  try{$version=Invoke-RestMethod 'http://127.0.0.1:9707/json/version' -TimeoutSec 3;$report.diagnostics.httpCheck='connected'}catch{$report.diagnostics.httpCheck='unreachable';throw '本地连接不可访问'}
  $targets=Invoke-RestMethod 'http://127.0.0.1:9707/json/list' -TimeoutSec 3
  foreach($target in @($targets|Where-Object {$_.type -eq 'page'})){
   $u=[uri]$target.url;if($request.operation -eq 'switch' -and $u.Host -ne 'doubao-chat'){continue};if($u.Scheme -notin @('doubao','chrome') -or $u.Host -notin @('doubao-chat','doubao-background')){continue}
   $socket=$null;$sidebarTemporarilyOpened=$false;$pageFocusEnabled=$false;$pageMetricsEnabled=$false
   try{
    $ws=[uri]$target.webSocketDebuggerUrl
    if($ws.Scheme -ne 'ws' -or $ws.Host -notin @('127.0.0.1','localhost','::1') -or $ws.Port -ne 9707){throw '连接地址无效'}
    $socket=New-Object AccountReadSocket;$socket.Connect($ws.AbsoluteUri)
    if($request.operation -eq 'switch' -and $u.Host -eq 'doubao-chat'){$pageFocusEnabled=Set-PageFocus $socket $true;$pageResumed=Resume-Page $socket;$report.menuAttempts+=@{stage='switch_page_focus';result=@{action=$(if($pageFocusEnabled){'enabled'}else{'unsupported'})}}}
    $accountIds=@($report.menuAttempts|Where-Object {$_.stage -eq 'locate'}|ForEach-Object {$_.regionButtons}|ForEach-Object {$_.automationId}|Where-Object {$_})
    $idsJson=ConvertTo-Json -InputObject @($accountIds) -Compress
    $expression=@'
(()=>{const nodes=[...document.querySelectorAll('[aria-label^="avatar "],[role="menuitem"],button')];const accounts=[];for(const el of nodes){const aria=el.getAttribute('aria-label')||'';const text=(el.textContent||'').trim();const label=/^avatar\s+/.test(aria)?aria:/^avatar\s+/.test(text)?text:'';if(label&&label.length<=300&&!accounts.some(a=>a.label===label))accounts.push({label});}const ids=__ACCOUNT_IDS__;const accountButtons=ids.map(id=>document.getElementById(id)).filter(Boolean).map(el=>({tag:el.tagName,role:el.getAttribute('role'),hasPopup:el.getAttribute('aria-haspopup'),expanded:el.getAttribute('aria-expanded')}));return {accountButtons,frameCount:document.querySelectorAll('iframe').length,accounts:accounts.slice(0,100),readyState:document.readyState,visibility:document.visibilityState}})()
'@
    $expressionTemplate=$expression
    $expression=$expressionTemplate.Replace('__ACCOUNT_IDS__',$idsJson)
    $value=Evaluate $socket $expression
    $loginExpression=@'
(()=>{const buttons=[...document.querySelectorAll('button,[role="button"]')];return {notLoggedIn:buttons.some(el=>{const r=el.getBoundingClientRect();const text=(el.getAttribute('aria-label')||el.innerText||el.textContent||'').trim();return /^(登录|立即登录|登录豆包)$/.test(text)&&r.width>=100&&r.height>=20&&r.left>=0&&r.left<80&&r.top>innerHeight-110&&r.bottom<=innerHeight+2})}})()
'@
    if($u.Host -eq 'doubao-chat'){$loginState=Evaluate $socket $loginExpression;if($loginState.notLoggedIn){$report['notLoggedIn']=$true;$report.sources+=@{kind='dom';page=$u.Host;status='not_logged_in'};continue}}
    if($u.Host -eq 'doubao-chat' -and @($value.accountButtons).Count -eq 0){
     $locateExpression=@'
(()=>{const nodes=[...document.querySelectorAll('button[aria-haspopup="menu"]')].filter(el=>{const r=el.getBoundingClientRect();return r.width>=100&&r.width<=360&&r.height>=20&&r.height<=85&&r.left>=0&&r.left<80&&r.top>innerHeight-110&&r.bottom<=innerHeight+2});const controls=[...document.querySelectorAll('button')].filter(el=>/^(展开侧边栏|收起侧边栏|展开左侧栏|收起左侧栏|展开导航|展开菜单)$/.test((el.getAttribute('aria-label')||el.textContent||'').trim())).map(el=>({name:(el.getAttribute('aria-label')||el.textContent||'').trim(),id:el.id}));return {candidates:nodes.map(el=>({id:el.id,label:(el.textContent||'').trim().slice(0,120),expanded:el.getAttribute('aria-expanded')})),sidebarControls:controls}})()
'@
     $found=Evaluate $socket $locateExpression;$report.menuAttempts+=@{stage='dom_locate_without_window';result=$found}
     if(@($found.candidates).Count -eq 0){
      $sidebarExpression=@'
(()=>{const matches=[...document.querySelectorAll('button,[role="button"]')].filter(el=>{const r=el.getBoundingClientRect();return r.width>=15&&r.width<=60&&r.height>=15&&r.height<=60&&r.left>=0&&r.left<50&&r.top>=0&&r.top<80&&!!el.querySelector('svg')});if(matches.length!==1)return {action:'none',count:matches.length,viewport:{width:innerWidth,height:innerHeight,visibility:document.visibilityState},buttonCount:document.querySelectorAll('button,[role="button"]').length,headerButtons:[...document.querySelectorAll('button,[role="button"]')].map(el=>{const r=el.getBoundingClientRect();return {tag:el.tagName,id:el.id,svg:!!el.querySelector('svg'),left:r.left,top:r.top,width:r.width,height:r.height}}).filter(r=>r.top<80&&r.left<100).slice(0,12)};const el=matches[0];el.setAttribute('data-canvas-sidebar-probe','v10');el.click();return {action:'temporary_expand',count:1}})()
'@
      $sidebarResult=Evaluate $socket $sidebarExpression;$report.menuAttempts+=@{stage='sidebar_expand';result=$sidebarResult}
      $sidebarTemporarilyOpened=$sidebarResult.action -eq 'temporary_expand'
      if($sidebarTemporarilyOpened){for($attempt=0;$attempt -lt 6;$attempt++){Start-Sleep -Milliseconds 250;$found=Evaluate $socket $locateExpression;if(@($found.candidates).Count -eq 1){break}};$report.menuAttempts+=@{stage='dom_locate_after_expand';result=$found}}
     }

     if(@($found.candidates).Count -eq 1 -and $found.candidates[0].id){
      $idsJson=ConvertTo-Json -InputObject @($found.candidates[0].id) -Compress
      $expression=$expressionTemplate.Replace('__ACCOUNT_IDS__',$idsJson)
      $value=Evaluate $socket $expression
     }
    }
    if($u.Host -eq 'doubao-chat'  -and @($value.accountButtons).Count -eq 1){
     $openExpression=@'
(()=>{const ids=__ACCOUNT_IDS__;const matches=ids.map(id=>document.getElementById(id)).filter(el=>el&&el.tagName==='BUTTON'&&el.getAttribute('aria-haspopup')==='menu');if(matches.length!==1)return {action:'ambiguous',count:matches.length};const el=matches[0];if(el.getAttribute('aria-expanded')==='true')return {action:'already_open'};el.dispatchEvent(new PointerEvent('pointerdown',{bubbles:true,cancelable:true,pointerId:1,pointerType:'mouse',button:0,buttons:1,isPrimary:true}));el.dispatchEvent(new PointerEvent('pointerup',{bubbles:true,cancelable:true,pointerId:1,pointerType:'mouse',button:0,buttons:0,isPrimary:true}));return {action:'page_pointer_events',expanded:el.getAttribute('aria-expanded')}})()
'@
     $opened=Evaluate $socket ($openExpression.Replace('__ACCOUNT_IDS__',$idsJson))
     $report.menuAttempts+=@{stage='dom_open';result=$opened}
     Start-Sleep -Milliseconds $(if($request.operation -eq 'switch'){80}else{650})
     $report.foregroundChanged=([BackgroundFocusRead]::GetForegroundWindow() -ne $initialForeground)
     if($true){
      $moreExpression=@'
(()=>{const entries=[...document.querySelectorAll('[role="menuitem"]')];const more=entries.filter(el=>(el.textContent||'').trim()==='更多账号');if(more.length!==1)return {action:'none',count:more.length};const el=more[0];el.dispatchEvent(new PointerEvent('pointermove',{bubbles:true,pointerType:'mouse',pointerId:1}));return {action:'submenu_hover',expanded:el.getAttribute('aria-expanded')}})()
'@
      $moreResult=Evaluate $socket $moreExpression;$report.menuAttempts+=@{stage='dom_more';result=$moreResult};Start-Sleep -Milliseconds $(if($request.operation -eq 'switch'){80}else{650})
     }
     $value=Evaluate $socket $expression
     # Read visible account menu labels, without selecting or clicking any account.
     $menuExpression=@'
(()=>{return {entries:[...document.querySelectorAll('[role="menuitem"]')].filter(el=>{const r=el.getBoundingClientRect();return r.width>0&&r.height>0}).map(el=>({label:(()=>{const aria=el.getAttribute('aria-label')||'';if(/^avatar\s+/.test(aria))return aria.trim().slice(0,300);const full=(el.innerText||el.textContent||'').trim();const badges=[...el.querySelectorAll('span,div,small')].filter(n=>!n.children.length&&/^[\u4e00-\u9fffA-Za-z0-9]{1,24}套餐$/.test((n.textContent||'').trim()));const badge=badges.length===1?(badges[0].textContent||'').trim():'';return badge&&full.endsWith(badge)?(full.slice(0,-badge.length).trim()+' '+badge).slice(0,300):full.slice(0,300)})(),hasPopup:el.getAttribute('aria-haspopup'),avatar:!!el.querySelector('img[alt="avatar"]'),imageAlts:[...el.querySelectorAll('img')].map(im=>im.getAttribute('alt')||'').slice(0,3)})).slice(0,100)}})()
'@
     $menuValue=$null;$previousCount=-1;$stableRounds=0
     $menuLimit=if($request.operation -eq 'switch'){40}else{8}
     for($menuAttempt=0;$menuAttempt -lt $menuLimit;$menuAttempt++){
      [void](Evaluate $socket $moreExpression)
      Start-Sleep -Milliseconds $(if($request.operation -eq 'switch'){100}else{400})
      $menuValue=Evaluate $socket $menuExpression
      $entryCount=@($menuValue.entries).Count
      if($entryCount -eq $previousCount){$stableRounds++}else{$stableRounds=0}
      $previousCount=$entryCount
      if($request.operation -eq 'switch'){
       $foundTarget=@($menuValue.entries|Where-Object {$name=([string]$_.label -replace '^avatar\s*','' -replace '(体验套餐|标准套餐|加强套餐|高级套餐|专业套餐|基础套餐|(?<=\s)[\u4e00-\u9fffA-Za-z0-9]{1,24}套餐)$','').Trim();$_.avatar -and !$_.hasPopup -and ($name -replace '\s+',' ') -ceq $request.name})
       if($foundTarget.Count -eq 1){break}
      }elseif($menuAttempt -ge 3 -and $stableRounds -ge 2){break}
     }
     $report.menuAttempts+=@{stage='dom_menu_labels';entries=$menuValue.entries}

     $excluded=@('设置','外观设置','收藏夹','豆包官网','检查更新','帮助与反馈','更多账号','添加账号','退出登录','豆包订阅','豆包订阅升级','免费领额度','免费领取额度')
     $parsed=@()
     foreach($entry in @($menuValue.entries)){
      $label=[string]$entry.label
      if(!$entry.avatar -or !$label -or $entry.hasPopup -or $label -in $excluded -or $label -match '^当前时段额度用量'){continue}
      $label=(($label -replace '^avatar\s+','') -replace '\s+',' ').Trim()
      if($label -in $excluded){continue}
      $subscription='';if($label -match '(体验套餐|标准套餐|加强套餐|高级套餐|专业套餐|基础套餐|(?<=\s)[\u4e00-\u9fffA-Za-z0-9]{1,24}套餐)$'){$subscription=$Matches[1];$label=$label.Substring(0,$label.Length-$subscription.Length).Trim()}
      if($label.Length -gt 0 -and $label.Length -le 120){
       $parsed+=@{name=$label;subscription=$subscription;label=$entry.label;avatar=$entry.avatar}
       if(!@($report.accounts|Where-Object {$_.name -ceq $label}).Count){$report.accounts+=@{name=$label;subscription=$subscription;label=$entry.label;source='dom_menu_candidate'}}
      }
     }
     $report.menuAttempts+=@{stage='parsed_account_candidates';count=$parsed.Count;accounts=$parsed}
     $identityExpression=@'
(()=>{
 const visible=el=>{const r=el.getBoundingClientRect();return r.width>0&&r.height>0};
 const label=el=>(el.getAttribute('aria-label')||el.innerText||el.textContent||'').trim();
 const buttons=[...document.querySelectorAll('button,[role="button"]')].filter(el=>{const r=el.getBoundingClientRect();return visible(el)&&r.left>=0&&r.left<80&&r.top>innerHeight-110&&r.bottom<=innerHeight+2&&r.width>=100&&r.width<=360&&r.height>=20&&r.height<=85});
 const menus=[...document.querySelectorAll('[role="menuitem"]')].filter(visible);
 const login=buttons.some(el=>/^(登录|立即登录|登录豆包)$/.test(label(el)));
 const logout=menus.some(el=>/^退出登录$/.test(label(el)));
 const identities=buttons.filter(el=>el.getAttribute('aria-haspopup')==='menu'&&!/^(登录|立即登录|登录豆包)$/.test(label(el)));
 if(logout&&identities.length===1){
  const el=identities[0];let name=(el.innerText||el.textContent||'').trim().split(/\r?\n/).map(x=>x.trim()).filter(Boolean)[0]||'';
  const full=(el.innerText||el.textContent||'').trim();const subscription=full.match(/(体验套餐|标准套餐|加强套餐|高级套餐|专业套餐|基础套餐|(?<=\s)[\u4e00-\u9fffA-Za-z0-9]{1,24}套餐)\s*$/)?.[1]||'';
  name=name.replace(/^avatar\s+/,'').replace(/(体验套餐|标准套餐|加强套餐|高级套餐|专业套餐|基础套餐|(?<=\s)[\u4e00-\u9fffA-Za-z0-9]{1,24}套餐)\s*$/,'').replace(/\s*[›〉»>]\s*$/,'').replace(/\s+/g,' ').trim();
  const lines=full.split(/\r?\n/).map(x=>x.replace(/\s*[›〉»>]\s*$/,'').trim()).filter(Boolean);
  const company=!subscription&&lines.length===2?lines[1]:'';
  if(company)name=name+' '+company;
  if(name&&name.length<=120&&!/^(登录|添加账号|豆包订阅|豆包订阅升级|免费领额度|免费领取额度)$/.test(name))return {status:'logged_in',name,subscription,evidence:'account_button_and_logout'};
 }
 return {status:login&&!logout?'not_logged_in':'unknown',evidence:login?'login_button':'none'};
})()
'@
     $identity=Evaluate $socket $identityExpression
     $report.menuAttempts+=@{stage='current_account_identity';result=$identity}
     if($identity.status -eq 'logged_in'){
      $report['currentAccountName']=[string]$identity.name
      if(!$report.accounts.Count){$report.accounts+=@{name=[string]$identity.name;subscription=[string]$identity.subscription;label=('avatar '+$identity.name+$(if($identity.subscription){' '+$identity.subscription}else{''}));source='verified_current_account'}}
     }
     if($request.operation -eq 'switch'){
      $targetName=[string]$request.name
      if(!$targetName){$targetName=([string]$request.label -replace '^avatar\s+','' -replace '\s+(体验套餐|标准套餐|加强套餐|高级套餐|专业套餐|基础套餐|(?<=\s)[\u4e00-\u9fffA-Za-z0-9]{1,24}套餐)$','').Trim()}
      if(!$targetName){throw '目标账号无效，请重新同步账号。'}
      $report['requestedAccountName']=$targetName
      $switchExpression=@'
(()=>{
 const wanted=__TARGET_NAME__;
 const clean=s=>s.replace(/^avatar\s*/,'').replace(/(体验套餐|标准套餐|加强套餐|高级套餐|专业套餐|基础套餐|(?<=\s)[\u4e00-\u9fffA-Za-z0-9]{1,24}套餐)\s*$/,'').trim();
 const items=[...document.querySelectorAll('[role="menuitem"]')].filter(el=>{const r=el.getBoundingClientRect();return r.width>0&&r.height>0&&!el.getAttribute('aria-haspopup')&&el.querySelector('img[alt="avatar"]')&&clean((el.getAttribute('aria-label')||el.innerText||el.textContent||'').trim())===wanted});
 if(items.length!==1)return {action:'ambiguous',count:items.length};
 const el=items[0];if(el.getAttribute('aria-disabled')==='true')return {action:'disabled'};
 const name=clean((el.getAttribute('aria-label')||el.textContent||'').trim());el.click();return {action:'clicked_element',name};
})()
'@
      $verifyExpression=@'
(()=>{
 const clean=s=>s.replace(/^avatar\s*/,'').replace(/(体验套餐|标准套餐|加强套餐|高级套餐|专业套餐|基础套餐|(?<=\s)[\u4e00-\u9fffA-Za-z0-9]{1,24}套餐)\s*$/,'').replace(/\s*[›〉»>]\s*$/,'').replace(/\s+/g,' ').trim();
 const known=__KNOWN_NAMES__;
 const wanted=__VERIFY_TARGET__;
 const all=[...document.querySelectorAll('button,[role="button"]')].filter(el=>el.getAttribute('aria-haspopup')==='menu'&&!el.closest('[role="menu"]'));
 const namesFor=el=>{const full=(el.innerText||el.textContent||'').trim();const lines=full.split(/\r?\n/).map(clean).filter(Boolean);const exact=known.filter(name=>clean(lines.join(' '))===name||clean(full)===name||clean(el.getAttribute('aria-label')||'')===name);return exact.length?exact:known.filter(name=>lines.includes(name))};
 const positioned=all.filter(el=>{const r=el.getBoundingClientRect();return r.width>=35&&r.width<=360&&r.height>=20&&r.height<=85&&r.left>=0&&r.left<80&&r.top>innerHeight-110&&r.bottom<=innerHeight+2});
 // Hidden/minimized windows may have no usable layout. Require a unique avatar identity instead.
 const semantic=all.filter(el=>el.querySelector('img[alt="avatar"]')&&namesFor(el).length===1);
 const buttons=positioned.length===1?positioned:semantic;
 if(buttons.length!==1)return {name:'',targetName:wanted,targetMatched:false,candidateCount:buttons.length,reason:'identity_not_unique',visibility:document.visibilityState,focused:document.hasFocus(),viewport:{width:innerWidth,height:innerHeight}};
 const el=buttons[0];const full=(el.innerText||el.textContent||'').trim();const lines=full.split(/\r?\n/).map(clean).filter(Boolean);const names=namesFor(el);
 return {name:names.length===1?names[0]:'',targetName:wanted,targetMatched:names.length===1&&names[0]===wanted,candidateCount:1,width:el.getBoundingClientRect().width,reason:names.length===1?'matched':'identity_text_unmatched',source:positioned.length===1?'layout':'avatar_identity',labels:lines.slice(0,4),visibility:document.visibilityState,focused:document.hasFocus()};
})()
'@
      $verifyExpression=$verifyExpression.Replace('__KNOWN_NAMES__',(ConvertTo-Json -InputObject @($report.accounts|ForEach-Object {$_.name}|Select-Object -Unique) -Compress))

      $verifyExpression=$verifyExpression.Replace('__VERIFY_TARGET__',($targetName|ConvertTo-Json -Compress))
      $originalViewport=Evaluate $socket '({width:innerWidth,height:innerHeight,scale:devicePixelRatio||1})'
      $report['originalViewport']=$originalViewport
      if($report.currentAccountName -cne $targetName){
       $selected=Evaluate $socket ($switchExpression.Replace('__TARGET_NAME__',($targetName|ConvertTo-Json -Compress)))
       $report.menuAttempts+=@{stage='switch_target';result=$selected;candidateCount=$selected.count}
       if($selected.action -ne 'clicked_element'){throw '未能唯一找到目标账号，请重新同步账号后重试。'}
       $report.menuAttempts+=@{stage='switch_click';result=@{action='exact_element_click'}}
      }
      $verifySidebarExpression=@'
(()=>{const matches=[...document.querySelectorAll('button,[role="button"]')].filter(el=>{const r=el.getBoundingClientRect();return r.width>=15&&r.width<=60&&r.height>=15&&r.height<=60&&r.left>=0&&r.left<50&&r.top>=0&&r.top<80&&!!el.querySelector('svg')});if(matches.length!==1)return {action:'none',count:matches.length,viewport:{width:innerWidth,height:innerHeight,visibility:document.visibilityState},buttonCount:document.querySelectorAll('button,[role="button"]').length,headerButtons:[...document.querySelectorAll('button,[role="button"]')].map(el=>{const r=el.getBoundingClientRect();return {tag:el.tagName,id:el.id,svg:!!el.querySelector('svg'),left:r.left,top:r.top,width:r.width,height:r.height}}).filter(r=>r.top<80&&r.left<100).slice(0,12)};const el=matches[0];el.setAttribute('data-canvas-sidebar-probe','v10');el.click();return {action:'temporary_expand',count:1}})()
'@
      $verified=$false;$verifyDeadline=[DateTime]::UtcNow.AddSeconds(20);$reconnectNeeded=$false;$reconnectCount=0;$sidebarChecked=$false;$verificationSamples=@();$verifyStarted=[DateTime]::UtcNow
      # Switching can reload the page and destroy its old debugger target. Never click twice.
      do{
       Start-Sleep -Milliseconds 100
       try{
        if($reconnectNeeded){
         if($socket){if($pageMetricsEnabled){try{[void](Set-PageMetrics $socket $null $true)}catch{}};if($pageFocusEnabled){try{[void](Set-PageFocus $socket $false)}catch{}};$socket.Dispose();$socket=$null};$pageFocusEnabled=$false;$pageMetricsEnabled=$false
         if(!(Test-DoubaoConnection)){continue}
         $freshTargets=Invoke-RestMethod 'http://127.0.0.1:9707/json/list' -TimeoutSec 1
         $freshPages=@($freshTargets|Where-Object {try{$pageUrl=[uri]$_.url;$_.type -eq 'page' -and $pageUrl.Scheme -in @('doubao','chrome') -and $pageUrl.Host -eq 'doubao-chat'}catch{$false}})
         if($freshPages.Count -ne 1){continue}
         $freshWs=[uri]$freshPages[0].webSocketDebuggerUrl
         if($freshWs.Scheme -ne 'ws' -or $freshWs.Host -notin @('127.0.0.1','localhost','::1') -or $freshWs.Port -ne 9707){throw '账号确认连接地址无效。'}
         $socket=New-Object AccountReadSocket;$socket.Connect($freshWs.AbsoluteUri)
         $pageFocusEnabled=Set-PageFocus $socket $true
         $pageResumed=Resume-Page $socket
         $report.menuAttempts+=@{stage='switch_page_focus_after_reload';result=@{action=$(if($pageFocusEnabled){'enabled'}else{'unsupported'})}}
         $reconnectNeeded=$false;$reconnectCount++;$sidebarChecked=$false
         $report.menuAttempts+=@{stage='switch_reconnect';result=@{action='connected'}}
        }
        $actual=Evaluate $socket $verifyExpression
        if($actual.viewport -and ($actual.viewport.width -eq 0 -or $actual.viewport.height -eq 0) -and !$pageMetricsEnabled -and $originalViewport.width -ge 300 -and $originalViewport.height -ge 200){
         $pageMetricsEnabled=Set-PageMetrics $socket $originalViewport $false
         $report.menuAttempts+=@{stage='switch_restore_render_size';result=@{action=$(if($pageMetricsEnabled){'applied'}else{'unsupported'})}}
         if($pageMetricsEnabled){continue}
        }
        $report['verificationIdentity']=$actual
        if($verificationSamples.Count -lt 30){$verificationSamples+=@{elapsedMs=[int]([DateTime]::UtcNow-$verifyStarted).TotalMilliseconds;identity=$actual}}
        $report['verificationSamples']=$verificationSamples
        if($actual.targetMatched -eq $true -and $actual.targetName -ceq $targetName){$verified=$true;break}
        if(!$actual.name -and !$sidebarChecked -and ($actual.candidateCount -eq 0 -or $actual.width -lt 100)){
         $sidebarChecked=$true;$expanded=Evaluate $socket $verifySidebarExpression
         $report.menuAttempts+=@{stage='switch_sidebar_check';result=$expanded}
        }
       }catch{$reconnectNeeded=$true}
      }while([DateTime]::UtcNow -lt $verifyDeadline)
      $report.menuAttempts+=@{stage='switch_verify';result=@{action=$(if($verified){'verified'}else{'not_verified'})}}
      if(!$verified){throw '尚未确认账号切换成功，请同步账号后重试。'}
      $report['switchVerified']=$true;$report['currentAccountName']=$targetName
     }
     # Close only the account menu opened by this test. Do not dismiss a previously open menu.
     if($opened.action -eq 'page_pointer_events'){
      $closeExpression=@'
(()=>{const ids=__ACCOUNT_IDS__;const buttons=ids.map(id=>document.getElementById(id)).filter(Boolean);if(buttons.length!==1)return {action:'none'};const el=buttons[0];if(el.getAttribute('aria-expanded')!=='true')return {action:'already_closed'};el.dispatchEvent(new PointerEvent('pointerdown',{bubbles:true,cancelable:true,pointerId:1,pointerType:'mouse',button:0,buttons:1,isPrimary:true}));el.dispatchEvent(new PointerEvent('pointerup',{bubbles:true,cancelable:true,pointerId:1,pointerType:'mouse',button:0,buttons:0,isPrimary:true}));return {action:'toggle_close',expanded:el.getAttribute('aria-expanded')}})()
'@
      $closed=Evaluate $socket ($closeExpression.Replace('__ACCOUNT_IDS__',$idsJson));Start-Sleep -Milliseconds 200
      $checked=Evaluate $socket $expression;$report.menuAttempts+=@{stage='cleanup';result=$closed;buttons=$checked.accountButtons}
     }

    }
    $count=Add-Accounts $value.accounts 'dom' 
    $report.sources+=@{kind='dom';page=$u.Host;status='read';accountEntries=$count;visibility=$value.visibility;readyState=$value.readyState;accountButtons=$value.accountButtons;frameCount=$value.frameCount}
   }catch{$report.sources+=@{kind='dom';page=$u.Host;status='unavailable';accountEntries=0;errorType=$_.Exception.GetType().FullName;errorMessage=$_.Exception.Message}}
   finally{if($socket){if($pageMetricsEnabled){try{[void](Set-PageMetrics $socket $null $true)}catch{}};if($pageFocusEnabled){try{[void](Set-PageFocus $socket $false)}catch{}};$socket.Dispose()}}
  }
 }else{$report.sources+=@{kind='connection';status='not_available';accountEntries=0}}
 $report.status=if($report.notLoggedIn){'not_logged_in'}elseif($report.accounts.Count){'candidates_found'}elseif($report.diagnostics.httpCheck -eq 'connected'){'connected_no_accounts'}else{'connection_blocked'}
 $report.message=if($report.accounts.Count){'只读检查完成，读取到候选账号，仍需核对完整性。'}elseif($report.diagnostics.httpCheck -eq 'connected'){'页面菜单测试完成，尚未读取到账号，请发送报告。'}else{'本地连接被阻塞，请发送报告排查。'}
}catch{
 $report.status=if($report.startup.error){$report.startup.error}elseif($report.diagnostics.processCount -eq 0){'doubao_not_running'}elseif($report.diagnostics.httpCheck -eq 'unreachable'){'interface_unreachable'}else{'probe_failed'}
 $report.message=switch($report.status){'doubao_not_running'{'没有检测到豆包进程，请确认豆包仍然运行。'} 'interface_unreachable'{'检测到豆包进程，但本地接口无法访问。'} default{'后台探测未完成，请将报告交回排查。'}}
}

$raw=@{ok=$false;accounts=@();currentLabel='';syncMode='background';message='未读取到账号，请确认豆包已登录后重试。'}
$raw.message=switch($report.status){
 'not_logged_in'{'豆包尚未登录，请登录后再次同步。'}
 'connection_required'{'豆包尚未连接，请先点击“连接豆包”，再同步账号。'}
 'installation_not_found'{'未找到豆包，请确认已安装豆包。'}
 'port_occupied'{'本地连接端口已被占用，请关闭占用连接的程序后重试。'}
 'restart_failed'{'豆包重新连接失败，请确认已退出豆包后再次同步。'}
 'launch_failed'{'豆包启动失败，请检查安装是否完整后重试。'}
 'startup_timeout'{'豆包连接尚未就绪，请确认启动完成并已登录后重试。'}
 'doubao_not_running'{'豆包尚未完成启动，请稍后再次同步。'}
 'interface_unreachable'{'豆包尚未连接，请先点击“连接豆包”，再同步账号。'}
 'connection_blocked'{'豆包本地连接不可用，请检查本地连接后重试。'}
 default{'未读取到账号，请确认豆包已登录后重试。'}
}
$raw.diagnostics=@{startup=$report.startup;status=$report.status;connection=$report.diagnostics.httpCheck;accountCount=$report.accounts.Count;requestedAccountName=$report.requestedAccountName;verificationIdentity=$report.verificationIdentity;verificationSamples=$report.verificationSamples;pageResumed=$pageResumed;originalViewport=$report.originalViewport;stages=@($report.menuAttempts|ForEach-Object {@{stage=$_.stage;candidateCount=$_.candidateCount;action=$_.result.action;identityStatus=$_.result.status;evidence=$_.result.evidence}});sources=@($report.sources|ForEach-Object {@{kind=$_.kind;status=$_.status;errorType=$_.errorType;errorMessage=$_.errorMessage;page=$_.page}})}
if($report.accounts.Count -and !$report.notLoggedIn){
 $raw.accounts=@($report.accounts|ForEach-Object {@{name=$_.name;label=('avatar '+$_.name+$(if($_.subscription){' '+$_.subscription}else{''}))}})
 $currentButton=@($report.menuAttempts|Where-Object {$_.stage -eq 'dom_locate_after_expand' -or $_.stage -eq 'dom_locate_without_window'}|ForEach-Object {$_.result.candidates}|Select-Object -First 1)
 $currentName=if($currentButton.Count){[string]$currentButton[0].label}else{[string](@($report.menuAttempts|Where-Object {$_.stage -eq 'locate'}|ForEach-Object {$_.regionButtons}|Select-Object -First 1)[0].name)}
 $currentName=($currentName -replace '^avatar\s+','' -replace '(体验套餐|标准套餐|加强套餐|高级套餐|专业套餐|基础套餐|(?<=\s)[\u4e00-\u9fffA-Za-z0-9]{1,24}套餐)$','').Trim()
 if($report.currentAccountName){$currentName=[string]$report.currentAccountName}
 $raw.currentName=$currentName
 $matches=@($raw.accounts|Where-Object {$_.name -ceq $currentName});if($matches.Count -eq 1){$raw.currentLabel=$matches[0].label}
 if($request.operation -ne 'switch' -or $report.switchVerified){$raw.ok=$true;$raw.message=if($request.operation -eq 'switch'){'账号已切换。'}else{'账号已同步。'}}
}
if($request.operation -eq 'switch' -and !$report.switchVerified){
 $raw.ok=$false;$failure=@($report.sources|Where-Object {$_.page -eq 'doubao-chat' -and $_.errorMessage}|Select-Object -Last 1)
 $raw.message=if($failure.Count -and $failure[0].errorMessage -match '[\u4e00-\u9fff]'){$failure[0].errorMessage}else{'账号切换未完成，请重新同步账号后重试。'}
}
$raw|ConvertTo-Json -Depth 5|Set-Content -LiteralPath $OutputFile -Encoding UTF8
