# 本机单二进制验收手册

本手册只用于 Mac 本机隔离验收。最终运行时是一个 Go 二进制，同时提供 Go API、用户 React SPA 和管理 React SPA。Vite 只参与构建，不是验收服务；不要运行 `npm run dev`、`vite preview`、Node production backend、GitHub Actions，也不要连接 ECS、公网服务、生产 MySQL、生产 Redis 或生产 TOS。

## 1. 验收边界和前置检查

从仓库根目录记录当前版本与工作树。正式验收应从干净提交构建；脚本在工作树不干净时会把 build-info 的 SHA 标成 `<sha>-dirty`，不得把它当成干净发布物。

```bash
git rev-parse HEAD
git status --short
lsof -nP -iTCP:18080 -sTCP:LISTEN
```

`lsof` 无输出才表示 `127.0.0.1:18080` 可用。另行确认现有 `8080` API 或 `5174` Vite 进程没有被关闭或复用；本次所有 HTTP 请求必须显式访问 `127.0.0.1:18080`。

验收只使用以下隔离资源：

- MySQL 数据库固定为 `ycm_staging`，DSN 的 host 必须是本机且库名必须是 `ycm_staging`。
- Redis 必须是本机独立实例；业务 key 前缀固定为 `ycm:staging:`。前缀不能把生产 Redis 变成本地安全资源，因此不要填写生产 Redis 地址。
- TOS object key 前缀固定为 `staging/`。默认不配置 TOS endpoint、bucket 或凭据，并且不执行上传、图片、TTS、视频等 Provider 操作。
- 日志写入本机独立目录，不写进仓库。

## 2. 创建数据库并仅迁移 `ycm_staging`

先用本机 MySQL 管理账号创建独立数据库。`-p` 会在终端安全地询问密码，不要把密码写进命令或文档。

```bash
mysql -h 127.0.0.1 -u root -p \
  -e 'CREATE DATABASE IF NOT EXISTS ycm_staging CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;'
```

在当前 shell 设置只指向本机 `ycm_staging` 的 Go MySQL DSN。下面的用户名和密码是占位符，必须换成本机专用账号，不能使用生产凭据。

```bash
export QIANTIE_MYSQL_DSN='<local-staging-user>:<local-staging-password>@tcp(127.0.0.1:3306)/ycm_staging?parseTime=true&charset=utf8mb4'
```

迁移前先确认 DSN 连接到预期数据库，再用已安装的 Goose 仅执行仓库迁移目录：

```bash
mysql --protocol=TCP -h 127.0.0.1 -u '<local-staging-user>' -p \
  -Nse 'SELECT DATABASE()' ycm_staging
goose -dir api/db/migrations mysql "$QIANTIE_MYSQL_DSN" status
goose -dir api/db/migrations mysql "$QIANTIE_MYSQL_DSN" up
goose -dir api/db/migrations mysql "$QIANTIE_MYSQL_DSN" status
```

输出数据库名不是 `ycm_staging` 时立即停止。Go 服务本身不会自动执行 Goose migration。

## 3. 构建唯一验收产物

从仓库根目录执行：

```bash
./scripts/build-embedded-ui.sh
test -x api/.staging-bin/ycm-server
```

脚本依次构建 `前台`、`后台`，只替换 `api/internal/webui/dist/user` 和 `api/internal/webui/dist/admin` 两个生成目录，再以 ldflags 注入当前 Git SHA 并构建 `api/.staging-bin/ycm-server`。它不会启动 Node 服务、访问 ECS 或调用 Actions。若 `/api/build-info` 的 `gitSha` 带 `-dirty`，先提交或清理本任务改动，再重新构建用于正式验收。

## 4. 启动与停止

只在本机专用终端设置下面的环境变量。首次空数据库启动需要一次性 bootstrap 账号；值只放在本机 shell 或不纳入版本控制的私有 env 文件中，不要打印或提交。数据库中已有可登录用户后，这两个 bootstrap 值不会重置账号。

```bash
export QIANTIE_GO_LISTEN_ADDR='127.0.0.1:18080'
export QIANTIE_ENV='local'
export QIANTIE_COOKIE_SECURE='false'
unset REDIS_ADDR TASK9_REDIS_ADDR
export QIANTIE_REDIS_ADDR='127.0.0.1:6379'
export QIANTIE_REDIS_PREFIX='ycm:staging:'
export TOS_KEY_PREFIX='staging/'
export QIANTIE_BOOTSTRAP_ADMIN_USERNAME='<local-staging-admin>'
export QIANTIE_BOOTSTRAP_ADMIN_PASSWORD='<local-staging-password>'
export YCM_STAGING_LOG_DIR="$HOME/Library/Logs/ycm-staging"
mkdir -p "$YCM_STAGING_LOG_DIR"
unset TOS_ENDPOINT TOS_REGION TOS_BUCKET TOS_ACCESS_KEY TOS_SECRET_KEY TOS_PUBLIC_BASE_URL
unset QIANTIE_TEXT_API_BASE_URL QIANTIE_TEXT_API_KEY QIANTIE_TEXT_MODEL
unset SHUIHUO_IMAGE_PROVIDER_BASE_URL SHUIHUO_IMAGE_PROVIDER_API_KEY SHUIHUO_IMAGE_PROVIDER_MODEL
unset SHUIHUO_TTS_PROVIDER_BASE_URL SHUIHUO_TTS_PROVIDER_API_KEY SHUIHUO_TTS_PROVIDER_MODEL
set -o pipefail
./api/.staging-bin/ycm-server 2>&1 | tee "$YCM_STAGING_LOG_DIR/server.log"
```

保持该前台进程运行，在另一个终端执行下面的检查。验收结束后回到服务终端按 `Ctrl-C` 停止；不要使用后台 `&`、`nohup` 或复用 Vite 端口。

应用选择 Redis 地址的优先级是 `REDIS_ADDR`、`QIANTIE_REDIS_ADDR`、`TASK9_REDIS_ADDR`。因此启动前必须清除前后两个兼容变量，再显式设置 `QIANTIE_REDIS_ADDR`；本机验收只允许 `127.0.0.1:<port>`，不得使用主机名、局域网地址、公网地址或生产 Redis。`QIANTIE_COOKIE_SECURE=false` 也必须显式设置，防止父 shell 中遗留的 `true` 让本机 HTTP 无法保存 Cookie。

此时需要外部执行器的操作必须显示真实的 `executor_unavailable` 或安全错误；不得为了验收触发任何付费 Provider。没有 TOS 配置时上传必须真实失败，不能伪造成功。

## 5. Go embed、深链和版本身份

```bash
curl -i http://127.0.0.1:18080/
curl -i http://127.0.0.1:18080/history/deep-link
curl -i http://127.0.0.1:18080/admin/
curl -i http://127.0.0.1:18080/admin/prompts/deep-link
curl -sS http://127.0.0.1:18080/api/build-info
```

四个页面响应都应为 HTML，并带 `X-YCM-Static-Source: go-embed`；用户深链不能返回管理端 HTML，管理深链也不能返回用户端 HTML。`/api/build-info` 的 `gitSha` 必须与构建前记录的 `git rev-parse HEAD` 完全一致且不能带 `-dirty`。这些结果只证明本机 Go embed 产物，不代表公网或 ECS 已发布。

可在浏览器打开 `http://127.0.0.1:18080/`，开发者工具 Network 中的文档和 `/assets/...` 必须来自 `127.0.0.1:18080`，不能来自 `5174`。

## 6. Cookie Session、CSRF、401 和 capability 403

匿名访问当前用户接口必须得到 `401`：

```bash
curl -i http://127.0.0.1:18080/api/auth/current-user
```

不带 `Origin` 或 `Referer` 的写请求应在任何写入前得到 CSRF `403`：

```bash
curl -i -X PUT \
  -H 'Content-Type: application/json' \
  --data '{}' \
  http://127.0.0.1:18080/api/v1/workspace/settings
```

响应 code 应为 `CSRF_REJECTED`。这只证明 CSRF，不等于 capability 403。

`QIANTIE_ENV=local` 只用于允许本机 HTTP 回环地址接收 Cookie；不得复制到公网部署配置。在浏览器用本机 bootstrap 账号登录，刷新页面后确认仍为同一账号，且 Cookie Session 能恢复；不要截图、复制或记录 Cookie 值。需要用命令行复核时，把账号密码保存在当前 shell 环境，通过标准输入提交，并将 cookie jar 放在本机独立日志目录：

```bash
export YCM_STAGING_USERNAME='<local-staging-admin>'
export YCM_STAGING_PASSWORD='<local-staging-password>'
jq -n --arg username "$YCM_STAGING_USERNAME" --arg password "$YCM_STAGING_PASSWORD" \
  '{username:$username,password:$password}' | \
  curl -i -c "$YCM_STAGING_LOG_DIR/cookies.txt" \
    -H 'Origin: http://127.0.0.1:18080' \
    -H 'Content-Type: application/json' \
    --data-binary @- http://127.0.0.1:18080/api/auth/login
curl -i -b "$YCM_STAGING_LOG_DIR/cookies.txt" \
  http://127.0.0.1:18080/api/auth/current-user
```

真正的 capability `403` 必须使用隔离数据库中的非管理员账号及其 Cookie；不能把匿名 401 或 CSRF 403 写成权限验收通过。服务首次启动并创建 bootstrap owner 后，在另一个终端加载相同的本机 `QIANTIE_MYSQL_DSN` 和 `QIANTIE_ENV=local`，再用受保护的维护命令创建一次性 member。该命令会拒绝非 `ycm_staging` 数据库、非 `127.0.0.1` MySQL、非 local 环境以及不带 `ycm-staging-member-` 前缀的用户名，并由 Go 代码生成密码哈希：

```bash
export QIANTIE_ENV='local'
export QIANTIE_AUTH_ADMIN_ACTION='local-create-member'
export QIANTIE_AUTH_ADMIN_USERNAME='ycm-staging-member-review'
export QIANTIE_AUTH_ADMIN_DISPLAY_NAME='Local acceptance member'
export QIANTIE_AUTH_ADMIN_PASSWORD='<local-member-password-at-least-12-characters>'
(cd api && go run ./cmd/auth-admin)
unset QIANTIE_AUTH_ADMIN_ACTION QIANTIE_AUTH_ADMIN_DISPLAY_NAME QIANTIE_AUTH_ADMIN_PASSWORD
```

不要在 bootstrap owner 之前创建 fixture，否则空库首次启动不会再创建 owner。创建完成后，用该 member 登录到独立 cookie jar，并访问它没有 capability 的管理 API：

```bash
export YCM_STAGING_MEMBER_USERNAME='ycm-staging-member-review'
export YCM_STAGING_MEMBER_PASSWORD='<local-member-password-at-least-12-characters>'
jq -n --arg username "$YCM_STAGING_MEMBER_USERNAME" --arg password "$YCM_STAGING_MEMBER_PASSWORD" \
  '{username:$username,password:$password}' | \
  curl -i -c "$YCM_STAGING_LOG_DIR/member-cookies.txt" \
    -H 'Origin: http://127.0.0.1:18080' \
    -H 'Content-Type: application/json' \
    --data-binary @- http://127.0.0.1:18080/api/auth/login
curl -i -b "$YCM_STAGING_LOG_DIR/member-cookies.txt" \
  http://127.0.0.1:18080/api/v1/admin/prompts
```

最后一个请求必须返回 `403` 和管理端专用错误码 `ADMIN_CAPABILITY_REQUIRED`。密码只存在于当前本机 shell，不作为命令参数、日志或文档内容输出。

## 7. 清理

停止 Go 进程后，删除本机 cookie jar，并仅清理本次专用 Redis namespace 与 `ycm_staging` 数据库。删除数据库前再次人工核对名称；不要使用通配符，不要清理生产或共享实例。因为默认不配置 TOS，本流程不产生 TOS 对象。

```bash
rm -f "$YCM_STAGING_LOG_DIR/cookies.txt" "$YCM_STAGING_LOG_DIR/member-cookies.txt"
export QIANTIE_ENV='local'
export QIANTIE_AUTH_ADMIN_ACTION='local-delete-member'
export QIANTIE_AUTH_ADMIN_USERNAME='ycm-staging-member-review'
(cd api && go run ./cmd/auth-admin)
unset QIANTIE_AUTH_ADMIN_ACTION QIANTIE_AUTH_ADMIN_USERNAME
unset YCM_STAGING_USERNAME YCM_STAGING_PASSWORD YCM_STAGING_MEMBER_USERNAME YCM_STAGING_MEMBER_PASSWORD
unset QIANTIE_BOOTSTRAP_ADMIN_USERNAME QIANTIE_BOOTSTRAP_ADMIN_PASSWORD
```

`local-delete-member` 只会删除 `ycm_staging` 本机库中匹配保留前缀、角色仍为 `member` 且未加入团队的 fixture；条件不匹配时失败，不会扩大删除范围。数据库和 Redis 的整体删除属于破坏性动作，按实际本机工具逐个确认后执行；本手册不提供可误操作生产资源的一键删除命令。
