$repo = Split-Path -Parent $PSScriptRoot
$env:QIANTIE_MYSQL_DSN = 'root:@tcp(127.0.0.1:3307)/qiantie_v88?parseTime=true'
$env:QIANTIE_BRIDGE_SECRET = '123456'
# No Node/3001 fallback in the local Go runtime.
$env:QIANTIE_GO_LISTEN_ADDR = ':4001'
$env:QIANTIE_GO_AUTH_ENABLED = '1'
$env:QIANTIE_BATCH_FACTORY_V11_SLICE = '6'
$env:QIANTIE_BATCH_FACTORY_V11_PRODUCTION_ENABLED = '1'
$env:QIANTIE_BATCH_FACTORY_V11_MERGE_ENABLED = '1'
$env:QIANTIE_BATCH_FACTORY_V11_LOCAL_MERGE_ENABLED = '1'
$env:QIANTIE_BATCH_FACTORY_V11_DYNAMIC_TEXT_PROVIDER = '1'
$env:QIANTIE_DATA_DIR = Join-Path $repo 'backend/.data-v88-review'
$env:QIANTIE_SCRIPT_HISTORY_DIR = Join-Path $repo 'data/users'
$env:QIANTIE_MEMBER_SYSTEM_DIR = Join-Path $repo 'data/system'
$env:QIANTIE_MEMBER_AVATAR_DIR = Join-Path $repo 'data/avatars'
$env:QIANTIE_LOCAL_EXECUTOR_ARTIFACT_DIR = Join-Path $repo 'backend/.data-v88-review/artifacts'
$env:QIANTIE_LOCAL_EXECUTOR_DOWNLOADS_DIR = Join-Path $repo 'backend/.data-v88-review/downloads'
& (Join-Path $repo 'backend/.build-verify/qiantie-local-v88-current.exe')
