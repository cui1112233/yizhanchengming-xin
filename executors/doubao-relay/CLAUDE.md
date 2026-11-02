@AGENTS.md

## Claude Code 补充

- 用户在用 relay 跑真实任务时,不要改 Rust 源码后重启 relay;需要重启时先确认没有进行中的任务
  (`sqlite3 ~/Library/Application\ Support/doubao-relay/relay.db "select id,status from tasks where status not in ('completed','failed','cancelled')"`)。
- relay 日志在 `~/Library/Logs/doubao-relay/doubao-relay.<日期>.log`(JSON lines),排查任务先按 `task_id` 过滤。
- 临时探测脚本放会话 scratchpad,不要放进仓库。
