param([string]$OwnerFile = '', [string]$ResourceFile = '', [string]$SourceSHA = '', [string]$TestPattern = '^TestNovelParityRealMySQL', [string]$Timeout = '25m')
$ErrorActionPreference='Stop'
$npRepo=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../../..'))
$npEvidence=Join-Path $npRepo '_local-archive-20261028/novel-parity-20261117'
New-Item -ItemType Directory -Force $npEvidence | Out-Null
$npStamp=Get-Date -Format 'yyyyMMdd-HHmmss'
$npContainer='ycm-mysql'; $npAddress='127.0.0.1:3306'
if($ResourceFile){
  $npResource=Get-Content -LiteralPath $ResourceFile -Raw | ConvertFrom-Json
  $npContainer=[string]$npResource.containerID; $npAddress=[string]$npResource.address
  if($npContainer -notmatch '^[a-f0-9]{64}$' -or $npAddress -notmatch '^127\.0\.0\.1:[0-9]{4,5}$'){throw 'Invalid dedicated MySQL identity'}
  $npActual=(& docker inspect $npContainer --format '{{.Id}}|{{.Image}}|{{.Name}}|{{json .HostConfig.Tmpfs}}').Trim()
  if($LASTEXITCODE -ne 0 -or $npActual -notlike "$npContainer|$($npResource.imageID)|/$($npResource.containerName)|*" -or $npActual -notmatch '/var/lib/mysql' -or (& docker port $npContainer 3306/tcp).Trim() -ne $npAddress){throw 'Dedicated MySQL resource identity changed'}
  $npMemory=[long]((& docker inspect $npContainer --format '{{.HostConfig.Memory}}').Trim())
  if($npMemory -ne [long]$npResource.memoryLimitBytes -or $npMemory -lt 536870912 -or $npMemory -gt 2147483648){throw 'Dedicated MySQL memory limit changed or is unbounded'}
}
if($OwnerFile){
  $npOwner=Get-Content -LiteralPath $OwnerFile -Raw | ConvertFrom-Json
  $npSchema=[string]$npOwner.schema; $npUser=[string]$npOwner.user; $npNonce=[string]$npOwner.nonce
  if($npOwner.containerID -and ($npOwner.containerID -ne $npContainer -or $npOwner.address -ne $npAddress)){throw 'Owned database container/address changed'}
}else{
  $npSuffix=[Guid]::NewGuid().ToString('N').Substring(0,12)
  $npSchema='novelparity20261117'+$npSuffix; $npUser='np'+$npSuffix; $npNonce=[Guid]::NewGuid().ToString('N')
}
if($npSchema -notmatch '^novelparity20261117[a-f0-9]{12}$' -or $npUser -ne ('np'+$npSchema.Substring(19)) -or $npNonce -notmatch '^[a-f0-9]{32}$'){throw 'Refusing non-owned resource'}
$npPassword=[Guid]::NewGuid().ToString('N')+[Guid]::NewGuid().ToString('N')
if($OwnerFile){
  $npCheck="SELECT COUNT(*) FROM $npSchema.novel_parity_fixture_owner WHERE task_id='novel-parity-task8' AND nonce='$npNonce';"
  $npCount=$npCheck | docker exec -i $npContainer sh -c 'MYSQL_PWD="$MYSQL_ROOT_PASSWORD" exec mysql -uroot --batch --skip-column-names'
  if($LASTEXITCODE -ne 0 -or $npCount -ne '1'){throw 'Owner marker does not match'}
  $npSQL="ALTER USER '$npUser'@'%' IDENTIFIED BY '$npPassword';"
}else{
  $npSQL="CREATE DATABASE $npSchema CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci; CREATE USER '$npUser'@'%' IDENTIFIED BY '$npPassword'; GRANT ALL PRIVILEGES ON $npSchema.* TO '$npUser'@'%';"
}
$npSQL | docker exec -i $npContainer sh -c 'MYSQL_PWD="$MYSQL_ROOT_PASSWORD" exec mysql -uroot --batch --skip-column-names'
if($LASTEXITCODE -ne 0){throw 'Owned resource preparation failed; do not blindly retry'}
$env:QIANTIE_TEST_NOVEL_MYSQL_DSN="${npUser}:${npPassword}@tcp(${npAddress})/${npSchema}?parseTime=true&loc=UTC&multiStatements=true&timeout=5s&readTimeout=5m&writeTimeout=2m"
$env:QIANTIE_TEST_NOVEL_MYSQL_ADDR=$npAddress
$env:QIANTIE_TEST_NOVEL_SCHEMA=$npSchema; $env:QIANTIE_TEST_NOVEL_NONCE=$npNonce; $env:QIANTIE_TEST_NOVEL_REQUIRE='1'
if(!$SourceSHA){$SourceSHA=(& git -C $npRepo rev-parse HEAD).Trim()}
if($SourceSHA -notmatch '^[a-f0-9]{40}$'){throw 'Explicit source SHA must be a full commit hash'}
$env:QIANTIE_TEST_NOVEL_SOURCE_SHA=$SourceSHA
Write-Output "VERIFY_SOURCE_SHA=$SourceSHA OWNED_SCHEMA=$npSchema ADDRESS=$npAddress"
@{schema=$npSchema;user=$npUser;nonce=$npNonce;sourceSHA=$env:QIANTIE_TEST_NOVEL_SOURCE_SHA;containerID=$npContainer;address=$npAddress;createdAt=(Get-Date).ToUniversalTime().ToString('o')} | ConvertTo-Json | Set-Content -LiteralPath (Join-Path $npEvidence "$npSchema-owner.json") -Encoding utf8
Push-Location (Join-Path $npRepo 'backend')
try{
  go test -buildvcs=false ./internal/novelcreation -run $TestPattern -count=1 "-timeout=$Timeout" -v 2>&1 | Tee-Object -FilePath (Join-Path $npEvidence "$npSchema-$npStamp-test.log")
  $npExit=$LASTEXITCODE
}finally{
  Pop-Location
  foreach($npName in @('QIANTIE_TEST_NOVEL_MYSQL_DSN','QIANTIE_TEST_NOVEL_MYSQL_ADDR','QIANTIE_TEST_NOVEL_SCHEMA','QIANTIE_TEST_NOVEL_NONCE','QIANTIE_TEST_NOVEL_REQUIRE','QIANTIE_TEST_NOVEL_SOURCE_SHA')){Remove-Item -LiteralPath "Env:$npName" -ErrorAction SilentlyContinue}
}
Write-Output "TEST_EXIT=$npExit OWNED_SCHEMA=$npSchema"
if($npExit -ne 0){throw "Real MySQL acceptance failed: $npExit"}
