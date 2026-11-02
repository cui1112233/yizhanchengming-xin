param([string]$InputFile,[string]$ReportFile)
$ErrorActionPreference='Stop'
[Console]::OutputEncoding=New-Object System.Text.UTF8Encoding($false)
$OutputEncoding=[Console]::OutputEncoding
$request=Get-Content -LiteralPath $InputFile -Raw -Encoding UTF8|ConvertFrom-Json
$report=@{submitAttempted=$false;submitted=$false;message='';stage='prepare'}
function Save {$report|ConvertTo-Json -Depth 15|Set-Content -LiteralPath $ReportFile -Encoding UTF8}
$script:submissionObservers=New-Object System.Collections.ArrayList
$script:submitArmed=$false;$script:submissionRequests=New-Object System.Collections.ArrayList
$socket=$null;$sequence=0;$focusEnabled=$false;$fileChooserIntercepted=$false;$script:fileChooserNodeId=$null
function Read-SubmissionObservers {
 foreach($observer in $script:submissionObservers){$raw=$null;while($observer.Messages.TryDequeue([ref]$raw)){try{$e=$raw|ConvertFrom-Json}catch{continue};if($script:submitArmed -and $e.method -eq 'Network.requestWillBeSent' -and $e.params.request.method -eq 'POST' -and $e.params.request.postData){[void]$script:submissionRequests.Add(@{url=[string]$e.params.request.url;method='POST';postData=[string]$e.params.request.postData})}}}
}
function Prepare-SubmissionObservers {
 $live=Invoke-RestMethod 'http://127.0.0.1:9707/json/list' -TimeoutSec 3
 foreach($target in $live){if($target.type -ne 'page' -or ([uri]$target.url).Host -ne 'doubao-background'){continue};$ws=[uri]$target.webSocketDebuggerUrl;if($ws.Scheme -ne 'ws' -or $ws.Host -notin @('localhost','127.0.0.1','::1') -or $ws.Port -ne 9707){continue}
  $observer=New-Object DoubaoCaptureSocket;$observer.Connect($ws.AbsoluteUri);[void]$script:submissionObservers.Add($observer)
  $observer.Send('{"id":1,"method":"Network.enable","params":{"maxPostDataSize":16777216}}')
  $ready=$false;$limit=[DateTime]::UtcNow.AddSeconds(3)
  while(!$ready -and [DateTime]::UtcNow -lt $limit){$raw=$null;while($observer.Messages.TryDequeue([ref]$raw)){try{$e=$raw|ConvertFrom-Json;if($e.id -eq 1 -and !$e.error){$ready=$true}}catch{}};if(!$ready){Start-Sleep -Milliseconds 25}}
  if(!$ready){throw '后台消息观察未就绪，任务未发送。'}
 }
 Read-SubmissionObservers
}
function Command([string]$method,$params,[int]$timeout=10000){
 $script:sequence++;$id=$script:sequence
 $socket.Send((@{id=$id;method=$method;params=$params}|ConvertTo-Json -Depth 15 -Compress))
 $deadline=[DateTime]::UtcNow.AddMilliseconds($timeout)
 while([DateTime]::UtcNow -lt $deadline){
  Read-SubmissionObservers
  $message=$null
  while($socket.Messages.TryDequeue([ref]$message)){
   $data=$message|ConvertFrom-Json
   if($script:submitArmed -and $data.method -eq 'Network.requestWillBeSent' -and $data.params.request.method -eq 'POST' -and $data.params.request.postData){[void]$script:submissionRequests.Add(@{url=[string]$data.params.request.url;method='POST';postData=[string]$data.params.request.postData})}
   if($data.method -eq 'Page.fileChooserOpened'){$script:fileChooserNodeId=$data.params.backendNodeId;continue}
   if($data.captureSocketClosed){throw '豆包本地连接已断开。'}
   if($data.id -eq $id){if($data.error){throw '豆包页面操作失败。'};return $data.result}
  }
  Start-Sleep -Milliseconds 25
 }
 throw '豆包页面操作超时，未重复发送。'
}
function Evaluate([string]$expression,[int]$timeout=15000){
 $value=Command 'Runtime.evaluate' @{expression=$expression;returnByValue=$true;awaitPromise=$true} $timeout
 if($value.exceptionDetails){$message=[string]$value.exceptionDetails.exception.description;if($message -match 'Error: ([^\r\n]+)'){$message=$Matches[1]};if($message -match '[\u4e00-\u9fff]'){throw $message};throw '豆包页面操作失败，请检查当前输入区。'}
 return $value.result.value
}
function Step([string]$name,[int]$timeout=20000){
 $report['operation']=$name
 $json=$request|ConvertTo-Json -Depth 8 -Compress
 return Evaluate ('shenlangDoubaoGeneration.run('+($name|ConvertTo-Json -Compress)+','+$json+')') $timeout
}
function Select-NativeModel {
 # Some desktop versions expose model options as native menu items, not page DOM.
 Add-Type -AssemblyName UIAutomationClient
 Add-Type -AssemblyName UIAutomationTypes
 $root=[Windows.Automation.AutomationElement]::RootElement
 $condition=New-Object Windows.Automation.PropertyCondition([Windows.Automation.AutomationElement]::ControlTypeProperty,[Windows.Automation.ControlType]::MenuItem)
 $matches=@();$wanted=([string]$request.modelName -replace '\s','')
 foreach($el in $root.FindAll([Windows.Automation.TreeScope]::Descendants,$condition)){
  try{
   if($el.Current.IsOffscreen -or !$el.Current.IsEnabled){continue}
   $label=([string]$el.Current.Name -replace '\s','')
   if($label -notmatch ('^'+[regex]::Escape($wanted)+'(?:[^A-Za-z0-9.]|$)')){continue}
   if((Get-Process -Id $el.Current.ProcessId -ErrorAction Stop).ProcessName -ine 'Doubao'){continue}
   $matches+=,$el
  }catch{}
 }
 if($matches.Count -ne 1){return $false}
 $pattern=$null
 if($matches[0].TryGetCurrentPattern([Windows.Automation.InvokePattern]::Pattern,[ref]$pattern)){$pattern.Invoke();return $true}
 if($matches[0].TryGetCurrentPattern([Windows.Automation.SelectionItemPattern]::Pattern,[ref]$pattern)){$pattern.Select();return $true}
 return $false
}
function Select-NativeUploadFile([string]$file){
 # Electron's native Open dialog is not a DOM file input or a CDP chooser event.
 Add-Type -AssemblyName UIAutomationClient
 Add-Type -AssemblyName UIAutomationTypes
 $deadline=[DateTime]::UtcNow.AddSeconds(8)
 while([DateTime]::UtcNow -lt $deadline){
  $windows=@();$root=[Windows.Automation.AutomationElement]::RootElement
  $condition=New-Object Windows.Automation.PropertyCondition([Windows.Automation.AutomationElement]::ControlTypeProperty,[Windows.Automation.ControlType]::Window)
  foreach($window in $root.FindAll([Windows.Automation.TreeScope]::Children,$condition)){
   try{if($window.Current.ClassName -ne '#32770' -or $window.Current.IsOffscreen){continue};if((Get-Process -Id $window.Current.ProcessId -ErrorAction Stop).ProcessName -ine 'Doubao'){continue};$windows+=,$window}catch{}
  }
  if($windows.Count -gt 1){throw '豆包有多个文件选择窗口，未选择文件'}
  if($windows.Count -eq 1){
   $dialog=$windows[0]
   $edits=@();$open=@()
   foreach($el in $dialog.FindAll([Windows.Automation.TreeScope]::Descendants,[Windows.Automation.Condition]::TrueCondition)){
    try{
     if($el.Current.IsOffscreen -or !$el.Current.IsEnabled){continue}
     if($el.Current.ControlType -eq [Windows.Automation.ControlType]::Edit -and ($el.Current.Name -match '^文件名|^File name' -or $el.Current.AutomationId -eq '1148')){$edits+=,$el}
     if($el.Current.ControlType -eq [Windows.Automation.ControlType]::Button -and ($el.Current.AutomationId -eq '1' -or $el.Current.Name -match '^打开|^Open')){$open+=,$el}
    }catch{}
   }
   if($edits.Count -eq 1 -and $open.Count -eq 1){
    $value=$null;$invoke=$null
    if($edits[0].TryGetCurrentPattern([Windows.Automation.ValuePattern]::Pattern,[ref]$value) -and $open[0].TryGetCurrentPattern([Windows.Automation.InvokePattern]::Pattern,[ref]$invoke)){
     $value.SetValue($file);$invoke.Invoke();$report['uploadSource']='doubao_native_file_dialog';return $true
    }
   }
   $report['nativeFileDialogFound']=$true;$report['nativeFilenameFields']=$edits.Count;$report['nativeOpenButtons']=$open.Count
  }
  Start-Sleep -Milliseconds 200
 }
 return $false
}
try{
 if(!$request.prompt -or $request.prompt.Length -gt 10000){throw '提示词无效。'}
 if($request.modelName -notin @('Seedream 5.0 Flash','Seedream 5.0 Pro','Seedream 4.5','Seedance 2.5','Seedance 2.0','Seedance 2.0 Fast','Seedance 2.0 Mini')){throw '豆包图片模型无效。'}
 if($request.type -eq 'video'){
  $max=if($request.modelName -eq 'Seedance 2.5'){30}else{15}
  if($request.duration -lt 4 -or $request.duration -gt $max -or [int]$request.duration -ne $request.duration){throw '视频时长无效。'}
  if($request.submissionMode -eq 'conversation'){$request.prompt='请使用 Seedance 2.0 Mini 生成'+$request.duration+'秒视频，比例'+$request.ratio+'。'+"`n"+$request.prompt}
 }
 if($request.ratio -notin @('自动','1:1','16:9','9:16','4:3','3:4','3:2','2:3','21:9')){throw '豆包生成比例无效。'}
 $files=@($request.referenceFiles);$maxReferences=if($request.type -eq 'video' -and $request.modelName -eq 'Seedance 2.0 Fast'){11}else{10};if($files.Count -gt $maxReferences){throw '参考图片最多10张，Fast参考视频最多1个。'}
 foreach($file in $files){if(!(Test-Path -LiteralPath $file -PathType Leaf)){throw '参考图片文件不存在。'}}
 if(!('DoubaoCaptureSocket' -as [type])){throw '本地生成组件未初始化，请重新发送。'}
 $targets=Invoke-RestMethod 'http://127.0.0.1:9707/json/list' -TimeoutSec 3
 $pages=@($targets|Where-Object {try{$u=[uri]$_.url;$_.type -eq 'page' -and $u.Scheme -in @('doubao','chrome') -and $u.Host -eq 'doubao-chat'}catch{$false}})
 if($pages.Count -ne 1){throw '未找到唯一豆包聊天页面，请检查本地连接。'}
 $ws=[uri]$pages[0].webSocketDebuggerUrl
 if($ws.Scheme -ne 'ws' -or $ws.Host -notin @('127.0.0.1','localhost','::1') -or $ws.Port -ne 9707){throw '豆包本地连接地址无效。'}
 $socket=New-Object DoubaoCaptureSocket;$socket.Connect($ws.AbsoluteUri)
 try{[void](Command 'Page.setWebLifecycleState' @{state='active'});[void](Command 'Emulation.setFocusEmulationEnabled' @{enabled=$true});$focusEnabled=$true}catch{}
 $source=[IO.File]::ReadAllText((Join-Path $PSScriptRoot 'doubao-generation-page.js'),[Text.Encoding]::UTF8)
 [void](Evaluate $source)
 if($request.submissionMode -eq 'work'){$request.prompt=Evaluate ('shenlangDoubaoGeneration.buildTextPrompt('+($request|ConvertTo-Json -Depth 12 -Compress)+')')}
 [Console]::Out.WriteLine('DOUBAO_STAGE:mode')
 try{$prepared=Step 'prepare' 60000}catch{
  $prepareError=$_.Exception.Message
  if($prepareError -match '未识别到所选(?:图片|视频)模型菜单项'){
   $nativeSelected=$false;try{$nativeSelected=Select-NativeModel}catch{}
   if(!$nativeSelected){throw $prepareError}
   $report['modelMenuSource']='native';$prepared=Step 'prepare' 60000
  }else{throw}
 }
 if($prepared.needsDurationInput){
  # Coarse drag once, then verify and adjust a single live slider step at a time.
  $matched=$false;$dragAttempts=0;$keyboardDisabled=$false;$keyboardBefore=$null;$keyboardKey=$null
  for($attempt=0;$attempt -lt 40;$attempt++){
   $plan=Step 'durationDragPlan'
   if($plan.matched){$matched=$true;break}
   if($null -ne $keyboardBefore){
    $expectedDelta=if($keyboardKey -eq 'ArrowRight'){1}else{-1}
    if(([int]$plan.duration-[int]$keyboardBefore) -ne $expectedDelta){$keyboardDisabled=$true}
    $keyboardBefore=$null
   }
   if($attempt -eq 0 -and $plan.endpointKey){
    $code=if($plan.endpointKey -eq 'End'){35}else{36}
    [void](Command 'Input.dispatchKeyEvent' @{type='keyDown';key=$plan.endpointKey;code=$plan.endpointKey;windowsVirtualKeyCode=$code;nativeVirtualKeyCode=$code})
    [void](Command 'Input.dispatchKeyEvent' @{type='keyUp';key=$plan.endpointKey;code=$plan.endpointKey;windowsVirtualKeyCode=$code;nativeVirtualKeyCode=$code})
    Start-Sleep -Milliseconds 250
    continue
   }
   if($dragAttempts -gt 0 -and !$keyboardDisabled -and $plan.fineTuneKey){
    $keyboardBefore=[int]$plan.duration;$keyboardKey=[string]$plan.fineTuneKey
    $code=if($keyboardKey -eq 'ArrowRight'){39}else{37}
    [void](Command 'Input.dispatchKeyEvent' @{type='keyDown';key=$keyboardKey;code=$keyboardKey;windowsVirtualKeyCode=$code;nativeVirtualKeyCode=$code})
    [void](Command 'Input.dispatchKeyEvent' @{type='keyUp';key=$keyboardKey;code=$keyboardKey;windowsVirtualKeyCode=$code;nativeVirtualKeyCode=$code})
    Start-Sleep -Milliseconds 180
    continue
   }
   if($dragAttempts -ge 8){break}
   $dragAttempts++
   $startX=[double]$plan.startX;$endX=[double]$plan.endX;$y=[double]$plan.y
   [void](Command 'Input.dispatchMouseEvent' @{type='mouseMoved';x=$startX;y=$y})
   [void](Command 'Input.dispatchMouseEvent' @{type='mousePressed';x=$startX;y=$y;button='left';buttons=1;clickCount=1})
   try{
    for($move=1;$move -le 12;$move++){
     $x=$startX+($endX-$startX)*$move/12
     [void](Command 'Input.dispatchMouseEvent' @{type='mouseMoved';x=$x;y=$y;button='left';buttons=1})
     Start-Sleep -Milliseconds 20
    }
   }finally{[void](Command 'Input.dispatchMouseEvent' @{type='mouseReleased';x=$endX;y=$y;button='left';buttons=0;clickCount=1})}
   Start-Sleep -Milliseconds 300
  }
  $lastPlan=Step 'durationDragPlan';if(!$lastPlan.matched){throw ('视频时长未生效：期望'+$request.duration+'秒，实际'+$lastPlan.duration+'秒；任务未发送。')}
  [void](Step 'verifyVideoSettings' 10000)
 }
 $report.stage='upload';Save
 if($files.Count){
  [void](Command 'Page.enable' @{})
  [void](Command 'Page.setInterceptFileChooserDialog' @{enabled=$true});$fileChooserIntercepted=$true
  [Console]::Out.WriteLine('DOUBAO_STAGE:upload')
  # Wait for each attachment before introducing the next; preserve canvas reference order.
  for($index=0;$index -lt $files.Count;$index++){
   [Console]::Out.WriteLine(('DOUBAO_UPLOAD:'+($index+1)+'/'+$files.Count))
   $script:fileChooserNodeId=$null;$nativeUploaded=$false
   try{[void](Step 'fileInput')}catch{
    $uploadError=$_.Exception.Message
    if(!$script:fileChooserNodeId){$nativeUploaded=Select-NativeUploadFile $files[$index];if(!$nativeUploaded){throw $uploadError}}
   }
   if(!$nativeUploaded){
   $dom=Command 'DOM.getDocument' @{depth=0}
   $node=Command 'DOM.querySelector' @{nodeId=$dom.root.nodeId;selector='input[data-shenlang-reference-upload="true"]'}
   if($script:fileChooserNodeId){[void](Command 'DOM.setFileInputFiles' @{backendNodeId=$script:fileChooserNodeId;files=@($files[$index])})}
   elseif($node.nodeId){[void](Command 'DOM.setFileInputFiles' @{nodeId=$node.nodeId;files=@($files[$index])})}
   else{throw '参考图上传入口已变化，未发送。'}
   }
   $request|Add-Member -NotePropertyName uploadedCount -NotePropertyValue ($index+1) -Force
   [void](Step 'uploadReady' 70000)
  }
 }
 [Console]::Out.WriteLine('DOUBAO_STAGE:prompt');$report.stage='prompt';Save
 [void](Step 'fill');[void](Step 'checkSend')
 if(Get-Command Sync-CaptureBeforeSend -CommandType Function -ErrorAction SilentlyContinue){Sync-CaptureBeforeSend}
 if($request.parallelTracking){Prepare-SubmissionObservers;[void](Command 'Network.enable' @{maxPostDataSize=16777216});[void](Evaluate 'null');$script:submissionRequests.Clear();$script:submitArmed=$true}
 $report.submitAttempted=$true;$report['sendStartedAt']=[DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds();$report.stage='sending';Save
 [void](Step 'send');$report.submitted=$true;$report.stage='submitted';$report.message='已发送，请等待生成结果。';Save
 if($request.parallelTracking -and $request.trackingKey -match '^shenlangDoubaoTask_[a-zA-Z0-9_]+$'){[void](Evaluate ('globalThis.'+$request.trackingKey+'=shenlangDoubaoGeneration'))}
 [Console]::Out.WriteLine('DOUBAO_STAGE:sent')
 if($request.parallelTracking){
  $identitySource=[IO.File]::ReadAllText((Join-Path $PSScriptRoot 'doubao-result-reader.js'),[Text.Encoding]::UTF8)
  $identityDeadline=[DateTime]::UtcNow.AddSeconds(2)
  do{
   [void](Evaluate 'null')
   $candidates=@($script:submissionRequests)
   foreach($candidate in $candidates){
    $identity=Evaluate ($identitySource+';shenlangDoubaoResultReader.submissionIdentity('+($candidate|ConvertTo-Json -Depth 15 -Compress)+','+($request.prompt|ConvertTo-Json -Compress)+', {verifiedSend:true})')
    if($identity){$report['submissionIdentity']=$identity;Save;break}
   }
   if($report['submissionIdentity']){break};Start-Sleep -Milliseconds 50
  }while([DateTime]::UtcNow -lt $identityDeadline)
  $script:submitArmed=$false;$script:submissionRequests.Clear()
 }

 # Return immediately: the capture loop reads both current-message data and DOM feedback.
 # Waiting here would prevent queued conversation replies from being processed.

}catch{
 $failureMessage=$_.Exception.Message
 $failedOperation=$report['operation']
 if($request.type -eq 'video' -and $socket){try{$report['videoSettingsDiagnostic']=Step 'videoSettingsDiagnostic' 3000}catch{};try{$report['uploadDiagnostic']=Step 'fileInputDiagnostic' 3000}catch{}}
 if($request.type -eq 'video'){
  try{
   $diagnostic=[ordered]@{version='1.1.7';model=$request.modelName;operation=$failedOperation;expectedRatio=$request.ratio;expectedDuration=$request.duration;message=$failureMessage;controls=$report['videoSettingsDiagnostic'];upload=$report['uploadDiagnostic'];nativeFileDialogFound=$report['nativeFileDialogFound'];nativeFilenameFields=$report['nativeFilenameFields'];nativeOpenButtons=$report['nativeOpenButtons']}
   $diagnostic|ConvertTo-Json -Depth 15|Set-Content -LiteralPath (Join-Path ([Environment]::GetFolderPath('Desktop')) '豆包视频操作诊断.json') -Encoding UTF8
   $diagnostic|ConvertTo-Json -Depth 15|Set-Content -LiteralPath (Join-Path ([Environment]::GetFolderPath('Desktop')) '豆包视频时长诊断.json') -Encoding UTF8
  }catch{}
 }
 $report.message=if($failureMessage -match '[\u4e00-\u9fff]'){$failureMessage}else{'图片发送失败，请检查豆包输入区。'};Save;throw $report.message
}finally{
 foreach($observer in $script:submissionObservers){$observer.Dispose()}
 if($socket){if($fileChooserIntercepted){try{[void](Command 'Page.setInterceptFileChooserDialog' @{enabled=$false} 2000)}catch{}};if($focusEnabled){try{[void](Command 'Emulation.setFocusEmulationEnabled' @{enabled=$false} 2000)}catch{}};$socket.Dispose()}
}
