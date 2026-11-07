# Production Host Stabilization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在不删除任何有效业务媒体的前提下，将生产服务器系统盘从 95% 降到 80% 以下，并关闭不必要的公网暴露。

**Architecture:** 先记录基线和建立最小回滚包，再仅清理已明确过期的备份、历史二进制、APT 缓存和日志。Nginx 改动使用新文件、`nginx -t` 和原子替换；UFW 在放行 22/80/443 后才启用。

**Tech Stack:** Ubuntu 24.04, systemd, Nginx, UFW, MySQL 8, Redis, Go 单二进制, PowerShell + Paramiko SSH

## Global Constraints

- 不删除 `/opt/www/qiantie-artifacts` 中的 PNG/JPG/MP3/MP4。
- 不删除当前运行目标 `/opt/www/qiantie/qiantie-202611071612`。
- 不修改 MySQL binlog 保留周期，不执行 Ubuntu 26.04 大版升级。
- SSH 密码登录本计划不关闭。
- 每个配置变更都必须有原文件副本、语法检查和回滚命令。
- 任何时候 `/health` 或 HTTPS 异常，立即停止后续步骤并回滚当前任务。

---

## File and System Map

- `/opt/www/qiantie/qiantie`: 当前运行二进制符号链接。
- `/opt/www/qiantie/qiantie.env`: 生产环境变量，不打印凭证值。
- `/opt/www/qiantie-artifacts`: 业务媒体目录，本计划只读。
- `/opt/www/nginx/nginx.conf`: Nginx 主配置。
- `/opt/www/nginx/conf.d/qiantie.conf`: 站点、80/443/3000 配置。
- `/etc/systemd/system/qiantie.service`: Qiantie systemd 单元，本计划只读。
- `/root/backup-20261107-host-stabilization`: 本次操作的配置、数据库和文件清单回滚包。
- `F:\脚本测试\一战晟铭-v88\work\20261107-qiantie-host-audit`: 既有审计证据，不覆盖。

### Task 1: Freeze the Baseline and Build a Minimal Rollback Bundle

**Files:**
- Create remotely: `/root/backup-20261107-host-stabilization/`
- Read: `/opt/www/qiantie/qiantie.env`
- Read: `/opt/www/nginx/nginx.conf`
- Read: `/opt/www/nginx/conf.d/qiantie.conf`
- Read: `/etc/systemd/system/qiantie.service`

**Interfaces:**
- Consumes: 当前运行主机状态。
- Produces: 可回滚配置、压缩 MySQL 备份、删除前文件清单和基线检查结果。

- [ ] **Step 1: Confirm exact host and active binary**

Run on the server:

```bash
set -euo pipefail
hostnamectl --static
ip -brief address
test "$(readlink -f /opt/www/qiantie/qiantie)" = "/opt/www/qiantie/qiantie-202611071612"
systemctl is-active --quiet qiantie nginx mysql redis-server
curl --fail --silent --show-error http://127.0.0.1:4000/health
```

Expected: SSH is connected to the requested public target `115.190.156.223`; the printed local address may be a cloud private address, so it is recorded but is not compared to the public IP. All service checks exit 0 and health returns a successful JSON response.

- [ ] **Step 2: Release recoverable package cache before creating the backup**

Run:

```bash
set -euo pipefail
apt-get clean
rm -rf -- /var/lib/apt/lists/partial/*
df -h /
```

Expected: root filesystem has more than 1GB available. This step does not remove installed packages.

- [ ] **Step 3: Create the exact rollback directory and configuration copies**

Run:

```bash
set -euo pipefail
rollback=/root/backup-20261107-host-stabilization
test ! -e "$rollback"
install -d -m 700 "$rollback"
install -m 600 /opt/www/qiantie/qiantie.env "$rollback/qiantie.env"
install -m 644 /opt/www/nginx/nginx.conf "$rollback/nginx.conf"
install -m 644 /opt/www/nginx/conf.d/qiantie.conf "$rollback/qiantie.conf"
install -m 644 /etc/systemd/system/qiantie.service "$rollback/qiantie.service"
readlink -f /opt/www/qiantie/qiantie > "$rollback/active-binary.txt"
sha256sum /opt/www/qiantie/qiantie-202611071612 > "$rollback/active-binary.sha256"
```

Expected: rollback directory exists with mode 700 and the env copy has mode 600.

- [ ] **Step 4: Record deletion candidates without deleting them**

Run:

```bash
set -euo pipefail
rollback=/root/backup-20261107-host-stabilization
find /root/backup-20261104 /root/backup-20261105 -xdev -printf '%y\t%s\t%TY-%Tm-%TdT%TH:%TM:%TS\t%p\n' | sort > "$rollback/legacy-backups.manifest"
find /opt/www/qiantie -maxdepth 1 -type f -name 'qiantie-*' -printf '%s  %f\n' | sort > "$rollback/release-binaries.manifest"
sha256sum /opt/www/qiantie/qiantie-* > "$rollback/release-binaries.sha256"
```

Expected: manifests are non-empty and the active binary appears in both release files.

- [ ] **Step 5: Create a fresh compressed database backup**

Run:

```bash
set -euo pipefail
rollback=/root/backup-20261107-host-stabilization
mysqldump --single-transaction --quick --routines --events --triggers qiantie_v88 | gzip -1 > "$rollback/qiantie_v88.sql.gz"
gzip -t "$rollback/qiantie_v88.sql.gz"
sha256sum "$rollback/qiantie_v88.sql.gz" > "$rollback/qiantie_v88.sql.gz.sha256"
```

Expected: `gzip -t` exits 0 and the backup is larger than 1MB.

- [ ] **Step 6: Save the baseline report**

Run:

```bash
set -euo pipefail
rollback=/root/backup-20261107-host-stabilization
{
  date -Is
  df -h /
  free -h
  systemctl --no-pager --full status qiantie nginx mysql redis-server
  ss -ltnp
  du -sh /opt/www/qiantie /opt/www/qiantie-artifacts /root/backup-* /var/cache/apt /var/lib/apt/lists /var/log 2>/dev/null
} > "$rollback/baseline.txt"
```

Expected: report records 22/80/443/3000 listeners and the root filesystem baseline.

- [ ] **Step 7: Commit checkpoint**

No Git commit is made for remote runtime data. Record the evidence under the existing audit case after execution.

### Task 2: Remove Only Verified Legacy Backups and Release Binaries

**Files:**
- Delete remotely: `/root/backup-20261104`
- Delete remotely: `/root/backup-20261105`
- Delete remotely: the explicit historical binary list below
- Preserve: `/root/backup-20261106-25`, `/root/backup-20261106-26`, `/root/backup-20261107-host-stabilization`

**Interfaces:**
- Consumes: Task 1 manifests and fresh database backup.
- Produces: at least 3GB additional free space without touching business media.

- [ ] **Step 1: Verify every preserved rollback target before deletion**

Run:

```bash
set -euo pipefail
test -x /opt/www/qiantie/qiantie-202611071612
test -x /opt/www/qiantie/qiantie-202611071559
test -x /opt/www/qiantie/qiantie-202611071547
test -x /opt/www/qiantie/qiantie-202611071513
test -x /opt/www/qiantie/qiantie-202611032018
test -s /root/backup-20261106-25/qiantie-202611061407
test -s /root/backup-20261106-26/qiantie-202611061625
gzip -t /root/backup-20261107-host-stabilization/qiantie_v88.sql.gz
```

Expected: all checks exit 0.

- [ ] **Step 2: Remove the two superseded backup directories**

Run:

```bash
set -euo pipefail
for target in /root/backup-20261104 /root/backup-20261105; do
  resolved="$(readlink -f "$target")"
  case "$resolved" in
    /root/backup-20261104|/root/backup-20261105) rm -rf --one-file-system -- "$resolved" ;;
    *) echo "refusing unexpected target: $resolved" >&2; exit 1 ;;
  esac
done
```

Expected: both exact directories are absent; later backup directories remain.

- [ ] **Step 3: Remove only the explicit superseded release binaries**

Run:

```bash
set -euo pipefail
cd /opt/www/qiantie
active="$(readlink -f qiantie)"
test "$active" = "/opt/www/qiantie/qiantie-202611071612"
remove=(
  qiantie-202611071414 qiantie-202611071331 qiantie-202611071324
  qiantie-202611071320 qiantie-202611071306 qiantie-202611071210
  qiantie-202611071116 qiantie-202611071037 qiantie-202611070228
  qiantie-202611061717 qiantie-202611061625 qiantie-202611061407
  qiantie-202611061334 qiantie-202611061043 qiantie-202611060214
  qiantie-202611060151 qiantie-202611060052 qiantie-202611060022
)
for name in "${remove[@]}"; do
  target="$(readlink -f -- "$name")"
  test "$target" = "/opt/www/qiantie/$name"
  test "$target" != "$active"
done
rm -f -- "${remove[@]}"
```

Expected: the five preserved releases remain and the active symlink still resolves to `qiantie-202611071612`.

- [ ] **Step 4: Bound journal size and rebuild APT indexes**

Run:

```bash
set -euo pipefail
journalctl --vacuum-time=7d --vacuum-size=100M
apt-get update
df -h /
```

Expected: root filesystem usage is below 80%; if not, stop and report before touching media or MySQL.

- [ ] **Step 5: Verify application and media counts**

Run:

```bash
set -euo pipefail
systemctl is-active --quiet qiantie nginx mysql redis-server
curl --fail --silent --show-error http://127.0.0.1:4000/health
find /opt/www/qiantie-artifacts -maxdepth 1 -type f | wc -l
mysql -Nse "SELECT COUNT(*) FROM batch_factory_v12_comic_page_images; SELECT COUNT(*) FROM batch_factory_v12_book_asset_images;" qiantie_v88
```

Expected: artifact file count is not lower than the Task 1 baseline and database counts remain unchanged.

- [ ] **Step 6: Commit checkpoint**

No source commit. Append before/after disk evidence to `work/20261107-qiantie-host-audit/evidence/` using a new evidence file.

### Task 3: Enforce HTTPS and Remove Public Port 3000

**Files:**
- Modify remotely: `/opt/www/nginx/nginx.conf`
- Modify remotely: `/opt/www/nginx/conf.d/qiantie.conf`
- Preserve copies: `/root/backup-20261107-host-stabilization/nginx.conf`, `/root/backup-20261107-host-stabilization/qiantie.conf`

**Interfaces:**
- Consumes: healthy Nginx baseline from Task 2.
- Produces: HTTP-to-HTTPS redirect, loopback-only port 3000, hidden Nginx version, basic response headers.

- [ ] **Step 1: Prepare the main configuration candidate**

Create `/tmp/nginx.conf.new` from the current main file and add the following inside `http {}` immediately after `default_type`:

```nginx
server_tokens off;
```

Expected candidate SHA differs from the current file only by this directive.

- [ ] **Step 2: Prepare the site configuration candidate**

Copy the current site file to `/tmp/qiantie.conf.new`, then make these exact changes:

```nginx
# In the 443 server block:
add_header X-Content-Type-Options "nosniff" always;
add_header Referrer-Policy "strict-origin-when-cross-origin" always;

# Replace the current port-80 proxy locations with:
location ^~ /.well-known/acme-challenge/ {
    root /opt/www/nginx/acme;
    try_files $uri =404;
}
location / {
    return 308 https://$host$request_uri;
}

# Replace the port-3000 listen directive:
listen 127.0.0.1:3000 default_server;
```

Expected: HTTPS proxy behavior is unchanged; ACME remains available; every other HTTP path redirects.

- [ ] **Step 3: Syntax-test candidates before replacement**

Run:

```bash
set -euo pipefail
install -d -m 700 /tmp/qiantie-nginx-test/conf.d
install -m 644 /tmp/nginx.conf.new /tmp/qiantie-nginx-test/nginx.conf
install -m 644 /tmp/qiantie.conf.new /tmp/qiantie-nginx-test/conf.d/qiantie.conf
sed -i 's#/opt/www/nginx/conf.d/\*.conf#/tmp/qiantie-nginx-test/conf.d/*.conf#' /tmp/qiantie-nginx-test/nginx.conf
nginx -t -c /tmp/qiantie-nginx-test/nginx.conf
```

Expected: syntax test is successful.

- [ ] **Step 4: Install atomically and reload**

Run:

```bash
set -euo pipefail
install -m 644 /tmp/nginx.conf.new /opt/www/nginx/nginx.conf.new
install -m 644 /tmp/qiantie.conf.new /opt/www/nginx/conf.d/qiantie.conf.new
mv -f /opt/www/nginx/nginx.conf.new /opt/www/nginx/nginx.conf
mv -f /opt/www/nginx/conf.d/qiantie.conf.new /opt/www/nginx/conf.d/qiantie.conf
nginx -t
systemctl reload nginx
```

Expected: reload exits 0 without restarting Qiantie.

- [ ] **Step 5: Verify locally and externally**

Run locally on the server:

```bash
set -euo pipefail
curl -sS -o /dev/null -D - http://127.0.0.1/health -H 'Host: yzcm.site'
curl --fail --silent --show-error https://yzcm.site/health
ss -ltnp | grep -E '127\.0\.0\.1:3000|0\.0\.0\.0:(80|443)'
```

Run from the Windows workstation:

```powershell
curl.exe -I http://yzcm.site/settings
curl.exe -I https://yzcm.site/settings
Test-NetConnection 115.190.156.223 -Port 3000
```

Expected: HTTP returns 308 with the same path on HTTPS, HTTPS responds, port 3000 is not reachable externally.

- [ ] **Step 6: Roll back on any failure**

Run only if validation fails:

```bash
set -euo pipefail
install -m 644 /root/backup-20261107-host-stabilization/nginx.conf /opt/www/nginx/nginx.conf
install -m 644 /root/backup-20261107-host-stabilization/qiantie.conf /opt/www/nginx/conf.d/qiantie.conf
nginx -t
systemctl reload nginx
```

Expected: original behavior is restored.

- [ ] **Step 7: Commit checkpoint**

No source commit. Save current Nginx SHA256, `nginx -t`, redirect and listener output in the audit evidence directory.

### Task 4: Enable a Minimal Firewall Without Losing SSH

**Files:**
- Modify remotely: UFW rules under `/etc/ufw`
- Backup remotely: `/root/backup-20261107-host-stabilization/ufw-before.tar.gz`

**Interfaces:**
- Consumes: loopback-only port 3000 from Task 3.
- Produces: default-deny inbound firewall with 22/80/443 allowed.

- [ ] **Step 1: Back up current firewall state**

Run:

```bash
set -euo pipefail
tar -C /etc -czf /root/backup-20261107-host-stabilization/ufw-before.tar.gz ufw
ufw status verbose > /root/backup-20261107-host-stabilization/ufw-status-before.txt
```

Expected: backup archive passes `gzip -t`.

- [ ] **Step 2: Add rules before enabling UFW**

Run while keeping the current SSH session open:

```bash
set -euo pipefail
ufw default deny incoming
ufw default allow outgoing
ufw allow 22/tcp comment 'SSH'
ufw allow 80/tcp comment 'HTTP redirect and ACME'
ufw allow 443/tcp comment 'HTTPS'
ufw --force enable
ufw status numbered
```

Expected: only 22/80/443 are allowed inbound.

- [ ] **Step 3: Open a second SSH connection before closing the first**

From the workstation, establish a new authenticated SSH session and run:

```bash
echo second-session-ok
curl --fail --silent http://127.0.0.1:4000/health
```

Expected: second session succeeds. If it fails, run `ufw disable` in the original session.

- [ ] **Step 4: Verify public ports**

Run from the workstation:

```powershell
22,80,443,3000,4000 | ForEach-Object {
  Test-NetConnection 115.190.156.223 -Port $_ |
    Select-Object RemotePort,TcpTestSucceeded
}
```

Expected: 22/80/443 true; 3000/4000 false.

- [ ] **Step 5: Commit checkpoint**

No source commit. Save `ufw status numbered`, listener and external port evidence.

### Task 5: Apply Ubuntu 24.04 Security Updates and Final Verification

**Files:**
- Modify remotely: installed Ubuntu 24.04 packages only
- Read: `/var/run/reboot-required`

**Interfaces:**
- Consumes: recovered disk headroom and working rollback bundle.
- Produces: patched host, or an explicit list of deferred packages if an update fails.

- [ ] **Step 1: Record exact upgrade candidates**

Run:

```bash
set -euo pipefail
apt-get update
apt list --upgradable 2>/dev/null | tee /root/backup-20261107-host-stabilization/upgradable-before.txt
df -h /
```

Expected: at least 2GB remains free before package installation. Stop otherwise.

- [ ] **Step 2: Install Ubuntu 24.04 updates without release upgrade**

Run:

```bash
set -euo pipefail
export DEBIAN_FRONTEND=noninteractive
apt-get -y upgrade
apt-get -y autoremove --purge
apt-get clean
```

Expected: command exits 0; no `do-release-upgrade` command is used.

- [ ] **Step 3: Reboot only when required**

Run:

```bash
if test -f /var/run/reboot-required; then
  cat /var/run/reboot-required.pkgs | tee /root/backup-20261107-host-stabilization/reboot-required.pkgs
  systemctl reboot
fi
```

Expected: if reboot occurs, SSH returns within 3 minutes.

- [ ] **Step 4: Run full post-change verification**

Run after reconnecting:

```bash
set -euo pipefail
systemctl is-active --quiet qiantie nginx mysql redis-server fail2ban
test "$(readlink -f /opt/www/qiantie/qiantie)" = "/opt/www/qiantie/qiantie-202611071612"
curl --fail --silent --show-error http://127.0.0.1:4000/health
curl --fail --silent --show-error https://yzcm.site/health
nginx -t
ufw status verbose
df -h /
free -h
journalctl -p err --since '-15 min' --no-pager
```

Expected: all services active, root filesystem below 80%, no new Qiantie/Nginx/MySQL/Redis errors.

- [ ] **Step 5: Verify user-visible pages**

Use the authenticated browser to open:

```text
https://yzcm.site/settings
https://yzcm.site/batch-factory
```

Expected: both pages load, existing comic images display, and existing merged media routes still respond.

- [ ] **Step 6: Record final evidence**

Append the exact before/after disk, package, service, firewall, HTTPS and browser results to the existing audit case. Do not claim the TOS migration is complete; it belongs to the second implementation plan.
