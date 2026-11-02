param([string]$InputFile,[string]$OutputFile)
$ErrorActionPreference='Stop'
$request=Get-Content -LiteralPath $InputFile -Raw -Encoding UTF8|ConvertFrom-Json
if($request.operation -in @('read','connect','switch')){
 try{
  & (Join-Path $PSScriptRoot 'doubao-accounts-background.ps1') -InputFile $InputFile -OutputFile $OutputFile
 }catch{
  @{ok=$false;accounts=@();currentLabel='';syncMode='background';message='后台同步组件执行失败，请查看同步诊断后重试。';diagnostics=@{status='component_error';errorType=$_.Exception.GetType().FullName}}|ConvertTo-Json -Depth 5|Set-Content -LiteralPath $OutputFile -Encoding UTF8
 }
 # Read always ends here, including failures. Foreground operations below are for account switching only.
 exit 0
}
$result=@{ok=$false;accounts=@();currentLabel='';message='账号读取失败，请检查桌面豆包是否开启。'}
Add-Type -TypeDefinition @'
using System;
using System.Runtime.InteropServices;
public static class AccountMouse {
 [DllImport("user32.dll")] public static extern IntPtr GetForegroundWindow();
 [DllImport("user32.dll")] public static extern bool ShowWindow(IntPtr h,int n);
 [DllImport("user32.dll")] public static extern bool SetForegroundWindow(IntPtr h);
 [DllImport("user32.dll")] public static extern bool IsWindow(IntPtr h);
 [DllImport("user32.dll")] public static extern bool IsIconic(IntPtr h);
 [DllImport("user32.dll")] public static extern bool SetCursorPos(int x,int y);
 [DllImport("user32.dll")] public static extern void mouse_event(uint f,uint x,uint y,uint d,UIntPtr e);
 public static void Click(int x,int y){SetCursorPos(x,y);mouse_event(2,0,0,0,UIntPtr.Zero);mouse_event(4,0,0,0,UIntPtr.Zero);}
}
'@
function Hover($e){if([AccountMouse]::GetForegroundWindow() -ne $script:targetWindow){throw '请将桌面豆包切到前台后重试。'};$r=$e.Current.BoundingRectangle;if($r.Width -le 0 -or $r.Height -le 0){throw '账号菜单不可见。'};[void][AccountMouse]::SetCursorPos([int]($r.Left+$r.Width/2),[int]($r.Top+$r.Height/2))}
function Click($e){if([AccountMouse]::GetForegroundWindow() -ne $script:targetWindow){throw '请将桌面豆包切到前台后重试。'};$r=$e.Current.BoundingRectangle;if($r.Width -le 0 -or $r.Height -le 0){throw '账号按钮不可见。'};[AccountMouse]::Click([int]($r.Left+$r.Width/2),[int]($r.Top+$r.Height/2))}
function Processes {return @(Get-Process -Name Doubao -ErrorAction SilentlyContinue|Where-Object {$_.MainWindowHandle -ne 0})}
function LaunchDoubao {
 $paths=@();foreach($p in @(Get-Process -Name Doubao -ErrorAction SilentlyContinue)){try{if($p.Path){$paths+=,$p.Path}}catch{}}
 foreach($key in @('HKCU:\Software\Microsoft\Windows\CurrentVersion\App Paths\Doubao.exe','HKLM:\Software\Microsoft\Windows\CurrentVersion\App Paths\Doubao.exe')){if(Test-Path $key){$paths+=,(Get-Item $key).GetValue('')}}
 foreach($key in @('HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\*','HKLM:\Software\Microsoft\Windows\CurrentVersion\Uninstall\*','HKLM:\Software\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\*')){foreach($item in @(Get-ItemProperty $key -ErrorAction SilentlyContinue|Where-Object {$_.DisplayName -match '豆包|Doubao'})){if($item.InstallLocation){$paths+=,(Join-Path $item.InstallLocation 'Doubao.exe');$paths+=,(Join-Path $item.InstallLocation 'app\Doubao.exe')};if($item.DisplayIcon){$paths+=,($item.DisplayIcon -replace ',\d+$','').Trim('"')}}}
 $paths+=,@("$env:LOCALAPPDATA\Doubao\app\Doubao.exe","$env:LOCALAPPDATA\Programs\Doubao\Doubao.exe","$env:ProgramFiles\Doubao\Doubao.exe")
 foreach($drive in @(Get-PSDrive -PSProvider FileSystem)){ $paths+=,(Join-Path $drive.Root 'Doubao\app\Doubao.exe') }
 $exe=@($paths|Where-Object {$_ -and (Split-Path $_ -Leaf) -ieq 'Doubao.exe' -and (Test-Path -LiteralPath $_ -PathType Leaf)}|Select-Object -Unique)
 if(!$exe.Count){throw '未找到已安装的桌面豆包，请先安装或打开豆包后重新同步。'}
 Start-Process -FilePath $exe[0]
}
function Window {
 $p=Processes
 if($p.Count -ne 1){throw '暂未找到唯一的豆包主窗口，请完成登录并仅保留一个主窗口。'}
 $script:targetWindow=$p[0].MainWindowHandle
 return [Windows.Automation.AutomationElement]::FromHandle($script:targetWindow)
}
function Locate {
 if(!(Processes).Count){LaunchDoubao}
 for($i=0;$i -lt 40;$i++){
  $p=Processes
  if($p.Count -eq 1){$script:targetWindow=$p[0].MainWindowHandle;[void][AccountMouse]::ShowWindow($script:targetWindow,9);[void][AccountMouse]::SetForegroundWindow($script:targetWindow);Start-Sleep -Milliseconds 350;return (Window)}
  if($p.Count -gt 1){throw '检测到多个豆包主窗口，请仅保留一个后同步。'}
  Start-Sleep -Milliseconds 500
 }
 throw '豆包启动后尚未出现主窗口，请完成启动和登录后重新同步。'
}
function Elements($w,$type){$c=New-Object Windows.Automation.PropertyCondition([Windows.Automation.AutomationElement]::ControlTypeProperty,$type);return $w.FindAll([Windows.Automation.TreeScope]::Descendants,$c)}
function Act($e){$p=$null;if($e.TryGetCurrentPattern([Windows.Automation.ExpandCollapsePattern]::Pattern,[ref]$p)){$p.Expand();return};if($e.TryGetCurrentPattern([Windows.Automation.InvokePattern]::Pattern,[ref]$p)){$p.Invoke();return};throw '账号菜单暂不支持直接打开，请在豆包中展开头像与更多账号后重试。'}
function Current($w){$wr=$w.Current.BoundingRectangle;$found=@();foreach($e in (Elements $w ([Windows.Automation.ControlType]::Button))){$c=$e.Current;$r=$c.BoundingRectangle;if(!$c.IsOffscreen -and !$c.IsPassword -and $c.IsEnabled -and $r.Left -lt ($wr.Left+80) -and $r.Top -gt ($wr.Bottom-85) -and $r.Width -ge 100 -and ![string]::IsNullOrWhiteSpace($c.Name)){$found+=,$e}};if($found.Count -ne 1){throw '无法唯一识别当前豆包账号，请保持豆包主窗口可见。'};return $found[0]}
function Accounts($w){$list=@();foreach($e in (Elements $w ([Windows.Automation.ControlType]::MenuItem))){$c=$e.Current;if(!$c.IsOffscreen -and !$c.IsPassword -and $c.IsEnabled -and $c.Name -match '^avatar\s+(.+)$'){$name=($Matches[1] -split '\s+(体验套餐|标准套餐|加强套餐|高级套餐|专业套餐|基础套餐|(?<=\s)[\u4e00-\u9fffA-Za-z0-9]{1,24}套餐)')[0].Trim();$list+=@{name=$name;label=$c.Name;element=$e}}};return $list}
try {
 Add-Type -AssemblyName UIAutomationClient
 Add-Type -AssemblyName UIAutomationTypes
 $request=Get-Content -LiteralPath $InputFile -Raw -Encoding UTF8|ConvertFrom-Json
 $w=Locate;$current=$null
 for($attempt=0;$attempt -lt 3;$attempt++){
  try{$w=Window;$current=Current $w;break}catch{
   if($attempt -eq 2){throw}
   $w=Locate
   $expand=@(Elements $w ([Windows.Automation.ControlType]::Button)|Where-Object {!$_.Current.IsOffscreen -and $_.Current.Name -match '^(展开侧边栏|展开左侧栏|展开导航|展开菜单)$'})
   if($expand.Count -eq 1){Act $expand[0]}
   Start-Sleep -Milliseconds 800
  }
 }
 $result.currentLabel=$current.Current.Name
 $result.currentName=($result.currentLabel -replace '^avatar\s+','').Trim()
 $accounts=@()
 for($menuAttempt=0;$menuAttempt -lt 3;$menuAttempt++){
  $w=Window
  if(@(Accounts $w).Count -eq 0){$current=Current $w;Click $current;Start-Sleep -Milliseconds 600}
  $w=Window;$more=@(Elements $w ([Windows.Automation.ControlType]::MenuItem)|Where-Object {!$_.Current.IsOffscreen -and $_.Current.Name -ceq '更多账号'})
  if($more.Count -eq 1 -and @(Accounts $w).Count -le 2){Hover $more[0];Start-Sleep -Milliseconds 900}
  $w=Window;$accounts=@(Accounts $w)
  if($accounts.Count){break}
  $w=Locate;Start-Sleep -Milliseconds 600
 }
 if(!$accounts.Count){throw '已找到豆包窗口，但未读取到账号列表，请确认已登录并展开左侧栏后重新同步。'}
 $currentName=($result.currentLabel -replace '^avatar\s+','' -split '\s+(体验套餐|标准套餐|加强套餐|高级套餐|专业套餐|基础套餐|(?<=\s)[\u4e00-\u9fffA-Za-z0-9]{1,24}套餐)')[0].Trim()
 $currentMatches=@($accounts|Where-Object {$_.name -ceq $currentName});$result.currentLabel=if($currentMatches.Count -eq 1){$currentMatches[0].label}else{''}
 $result.accounts=@($accounts|ForEach-Object {@{name=$_.name;label=$_.label}})
 if($request.operation -eq 'switch'){
  $targets=@($accounts|Where-Object {$_.label -ceq $request.label})
  if($targets.Count -ne 1){throw '目标账号未唯一匹配，请重新获取账号列表。'}
  $target=$targets[0];$p=$null
  if($target.element.TryGetCurrentPattern([Windows.Automation.SelectionItemPattern]::Pattern,[ref]$p)){$p.Select()}elseif($target.element.TryGetCurrentPattern([Windows.Automation.InvokePattern]::Pattern,[ref]$p)){$p.Invoke()}else{throw '此账号暂不支持自动切换。'}
  $verified=$false
  for($i=0;$i -lt 12;$i++){Start-Sleep -Milliseconds 500;try{$button=Current (Window);$text=$button.Current.Name;$escaped=[regex]::Escape($target.name);if($text -match ('^(avatar\s+)?'+$escaped+'(\s+(体验套餐|标准套餐|加强套餐|高级套餐|专业套餐|基础套餐|(?<=\s)[\u4e00-\u9fffA-Za-z0-9]{1,24}套餐).*)?$')){$result.currentLabel=$target.label;$verified=$true;break}}catch{}}
  if(!$verified){throw '账号切换尚未确认，请查看豆包左下角当前账号后重新获取。'}
 }
 $result.ok=$true;$result.message=if($request.operation -eq 'switch'){'账号已切换。'}else{'账号已获取。'}
}catch{$result.message=if($_.Exception.Message -match '[\u4e00-\u9fff]'){$_.Exception.Message}else{'无法读取账号，请保持桌面豆包主窗口可见，并开启本地连接后重试。'}}
finally {
 if($request.operation -eq 'read' -and $request.returnWindowHandle){
  try{
   $returnHandle=[IntPtr]([long]$request.returnWindowHandle)
   if([AccountMouse]::IsWindow($returnHandle)){
    if([AccountMouse]::IsIconic($returnHandle)){[void][AccountMouse]::ShowWindow($returnHandle,9)}else{[void][AccountMouse]::ShowWindow($returnHandle,5)}
    for($returnAttempt=0;$returnAttempt -lt 3;$returnAttempt++){
     [void][AccountMouse]::SetForegroundWindow($returnHandle)
     Start-Sleep -Milliseconds 120
     if([AccountMouse]::GetForegroundWindow() -eq $returnHandle){break}
    }
   }
  }catch{}
 }
}

$result|ConvertTo-Json -Depth 5|Set-Content -LiteralPath $OutputFile -Encoding UTF8
