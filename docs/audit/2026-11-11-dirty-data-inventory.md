# V88 本地库存量坏数据盘点（只读）

- 盘点时间：2026-11-11（快照）
- 数据源：本机 `ycm-mysql` / `qiantie_go_local`（3306），只读连接，未连线上库
- 执行方式：仅 `SELECT / DESCRIBE / SHOW`；未执行任何 UPDATE / DELETE / INSERT / DDL
- 本轮未改代码、未构建、未重启 4001、未跑真实生成

> 与上一轮审计的差异：数字会变。最明显的是「6 页卡生成中」已在 2026-11-11 11:21:04 被现有回收逻辑自动转成 `generation_failed`，库内已无 `status='generating'` 的页。

## 一、总览

| 类别 | 上一轮 | 本次实测 | 结论方向 |
|---|---|---|---|
| `IMAGE_BILLED_NOT_RETRIEVED` | 23 条 | 23 条，全部已 resolved；活跃未取回 0 | 纯历史审计记录，建议保留 |
| 卡在生成中的页 | 6 页（最长 11.7 天） | 0 页；6 页已于 11-11 11:21:04 回收为失败态 | 已被自动回收，无需再清 |
| 没有主图的页 | 28 页 | 30 页（且这 30 页是“零张图”，不是“有图无主图”） | 建议按书补图或保留失败态 |
| 没有主图的资产 | 91 张 | 91 个资产“一张图都没有”；有图的资产 100% 有主图 | 不建议删，按需补图 |
| zombie `book_stage_runs` running | 15 条（≤11-03） | 15 条，日期全部 ≤2026-11-03 | 建议标记失败（仅账本） |

补充口径（本次实测）：
- `comic_pages` 1466 页：`image_ready` 1435、`generation_failed` 26、`prompt_ready` 4、`ready` 1。
- `comic_page_images` 1444 行：主图 1436、非主图 8；没有“一个页面有多张主图”的情况。
- `book_asset_images` 323 行：主图 306、非主图 17；`book_assets` 397 个，其中 306 个有主图、91 个没有任何图。

## 二、分类清单

### A. `IMAGE_BILLED_NOT_RETRIEVED`（23 条）

- 归属：唯一 job `go-auto-1794061558368134000`，batch `batch_5f35e4adab7b146ef5d712051b7b475c`（P0??-??-20261103），书 `book_31ec09d8158f60fc42690f637b296216`（下凡渡情劫，他表妹说我在钓金龟），owner `choushiyiguai`。
- 现状：job `completed`；书状态 `completed / published`，消息「漫画合集已发布」；`billedNotRetrievedCount=0`；全书 35 页全部 `image_ready`、每页 1 张主图；合集 `comic-collection_c0ca2771eae92c88f3e5b5ccd6e8b96d` ready、35 个条目。
- 这 23 条的位置：`book_31ec09d8...nodes.comic_images.pageLedger.<videoId>.lastResolvedErrorCode`（全部是“已解决”字段，不是待处理字段）。
- 对应页（页标签）：漫画页 01–05、10–12、15–26、33–35，共 23 页；`lifetimeAttempts` 多为 1，21 号页为 3；`cost` 全部为 0。
- 全部在 `2026-11-07T18:09:41Z` 统一 resolved。
- 同一账本里另有 7 条 `BODY_SLOW`、2 条 `BODY_STALLED`，同样全部 resolved，未形成待补页。

判定：**背后有图有成品**。删改只会丢失“曾计费未取回、后来又成功”的审计轨迹，不影响页面、合集或发布状态。

### B. 卡在“生成中”的页（原 6 页）

- 现状：`status='generating'` 的页 = 0。
- 6 页身份（均已变成 `generation_failed`，`updated_at` 均为 `2026-11-11 11:21:04`）：
  - 书 `book_d0a73b5e66c7caeb9def33829d9d963c`（第二次选择，我谁也不要了）5 页：
    `video_01e3a753f07da74300fa19d5e63fe2b2`、`video_6013fba2532e1f857daa9223b915ae44`、`video_86d1b20a7354bca1ae4e11909753cc65`、`video_c431950d05a3e7d5366348438813e50f`、`video_b1891968a75056433bec731cf7255240`
  - 书 `book_fa42f92d40c64f2bf3fbb64b3743875b`（唯爱遥遥）1 页：`video_40061b41dde09bdbd2f8a7c51f304cc4`
- 创建时间 2026-10-30，最长等待到回收时约 12.0 天（与上一轮“11.7 天”同源，只是测量时点不同）。
- 证据：6 行的 `updated_at` 完全落在同一分钟内，且代码里存在 `ListStaleGeneratingComicPages` / `ExpireStaleGeneratingComicPage`（`backend/internal/batchfactoryv11/comic_pages_mysql.go:226-260`，把超时 `generating` 改为 `generation_failed`）。据此判断是被回收任务批量处理的（推断，但时间戳高度吻合）。

判定：**已自动回收**，不需要再做“回收成失败态”；这 6 页现在属于下面 C 类。

### C. 没有图的页（30 页）

按书汇总：

| 书 ID | 书名/批次 | 缺图页数 | 页状态 |
|---|---|---|---|
| `book_d0a73b5e66c7caeb9def33829d9d963c` | 第二次选择，我谁也不要了 / batch_1daf（漫画测试批量） | 22 | 全部 `generation_failed` |
| `book_fa42f92d40c64f2bf3fbb64b3743875b` | 唯爱遥遥 / batch_d419（测试2） | 2 | `generation_failed` |
| `book_4090786e93fcf091439f12369c9dd02b` | 小妾造反后 / batch_2f46 | 2 | 1 `generation_failed` + 1 `prompt_ready` |
| `book_0007bf04022e8b39cb9414276fe87eda` | 小妾造反后 / batch_d147（小说获取·9本） | 2 | 全部 `prompt_ready` |
| `book_6804eadae7472d163a52d172a17d8aa0` | 长姐落选那日… / batch_d147 | 1 | `prompt_ready` |
| `book_db4f7c53c56fb67b51471201a096a331` | 测试书 / batch_59965（V88??-1-25???-20261111，owner `ceshi_owner2`） | 2 | 1 `ready` + 1 `generation_failed` |

状态合计：`generation_failed` 25、`prompt_ready` 4、`ready` 1（全库另有 1 个 `generation_failed` 页是有图的，不在此清单）。

样本页标识（每书 2 条，完整 30 条可按同一条件导出）：
- d0a73b5e：`video_01e3a753f07da74300fa19d5e63fe2b2`、`video_f2c4c0829417539a6530608731018f18`
- fa42f92d：`video_e089daaa3656d94450ad02ddf34eefb1`、`video_40061b41dde09bdbd2f8a7c51f304cc4`
- 4090786e：`video_a9d8a50debba4c6714213e28c8a54b1f`、`video_ddb4837bca4f5b2daf2c4366bbd5a0c7`
- 0007bf04：`video_bd788c4dfd3fc15970d048d016f783de`、`video_ea97dbeea0ed67e78962736e8c1fb441`
- 6804eada：`video_b8c11996d63027a646904661bac1afbc`
- db4f7c53：`video_234dcdb19ebb4b679ab8507655c4ccd0`、`video_c37ddb9ae472b766dd8676de2cbabf21`

引用检查：这 30 页在 `comic_collection_items` 里**没有任何条目**（按 `comic_page_id` 左连接全部为 NULL）。也就是说现有合集没有引用缺图页，不会出现“合集指向空图”的直接坏链；但书级状态与页面状态存在缺口：`book_fa42f92d` 的 job 已是 `completed / uploaded`、合集 ready（35 条），却有 1 页始终缺图且未入合集。

### D. 没有任何图的资产（91 个）

按书分布（15 本，合计 91 个）：

| 书 ID | 批次 | 资产总数 | 无图资产 | 样本资产 |
|---|---|---|---|---|
| `book_0007bf04022e8b39cb9414276fe87eda` | d147 | 10 | 10 | `comic-asset_0a8381cf...`（月锦） |
| `book_6804eadae7472d163a52d172a17d8aa0` | d147 | 14 | 14 | `comic-asset_21864cfd...`（prop） |
| `book_269ea8a97ba959004acb92342fed3d45` | 2f46 | 14 | 8 | `comic-asset_03518a19...`（scene） |
| `book_31ec09d8158f60fc42690f637b296216` | 5f35 | 7 | 6 | `comic-asset_1a4f13bb...`（scene） |
| `book_44922afa8ab3d462156309d9f7e8fc9a` | d147 | 7 | 6 | `asset_1e6d3804...`（prop） |
| `book_03915b7f85675c993c1e11df9077c2b2` | 8beb | 5 | 5 | `comic-asset_1adefd49...`（赵家院落正午） |
| `book_0695c6edc3f8c3c8021b4cb09a420526` | 8192 | 7 | 7 | `asset_2c6c53f2...`（赵德发） |
| `book_0c55e48f6285330a95ebc9246ef2cc8f` | 27db | 7 | 7 | `asset_76cd3261...`（傅总） |
| `book_20cd2f38db2f7f43ba3daad30399942c` | 27db | 6 | 5 | `asset_0fb08201...` |
| `book_3d00ca6e44118ef7c4c19686cfa94851` | 27db | 5 | 5 | `asset_178c3ee5...` |
| `book_f4aa884211da8b780eaed659784549dd` | 27db | 6 | 6 | `asset_1916814e...` |
| `book_29114d2323d031c922506d0ff69dfee5` | 0a75 | 4 | 3 | `asset_0b73829c...` |
| `book_37232d6084fe24badf90d913316ca881` | 0888 | 3 | 3 | `comic-asset_86d3f198...` |
| `book_37d17ce1b6507e935a2f2a01a0f7d6ae` | 672 | 3 | 3 | `asset_1af2dfe3...` |
| `book_b1efd66743c7fb0eb374542c8f24ef94` | a74b | 4 | 3 | `comic-asset_019db6e9...` |

另：17 行“非主图”资产图分布在 4 本书（`269ea8a9`、`4090786e`、`e6dd48a4`、`fa42f92d`），它们所属的资产都另有主图，不构成“缺主图”。

引用检查：`comic_pages.reference_image_urls` 中引用资产 ID / `asset-image_` 存储名的页数均为 **0**。因此这批“无图资产”目前没有被任何页当作参考图引用。

判定：**有元数据（角色/场景/道具名、prompt、facts/appearance），无图像、无引用**。删掉资产行不会破坏已有页面或合集，但会丢掉后续补图/重跑所需的设定文本。

### E. zombie `book_stage_runs`（15 条 running，全部 ≤11-03）

| 书 ID | 阶段 | attempt | 创建时间 |
|---|---|---|---|
| `book_29114d2323d031c922506d0ff69dfee5` | assets | 2 | 2026-10-14 10:48:50 |
| 同上 | assets | 3 | 2026-10-14 10:51:03 |
| 同上 | assets | 4 | 2026-10-14 10:52:45 |
| 同上 | assets | 5 | 2026-10-14 10:55:34 |
| 同上 | assets | 6 | 2026-10-14 11:18:12 |
| `book_3d00ca6e44118ef7c4c19686cfa94851` | assets | 3 | 2026-10-16 18:26:26 |
| `book_0c55e48f6285330a95ebc9246ef2cc8f` | opening | 4 | 2026-10-17 05:09:41 |
| `book_f4aa884211da8b780eaed659784549dd` | director | 3 | 2026-10-17 05:27:30 |
| `book_20cd2f38db2f7f43ba3daad30399942c` | assets | 8 | 2026-10-17 06:40:31 |
| `book_4090786e93fcf091439f12369c9dd02b` | comic | 6 | 2026-10-28 16:06:12 |
| 同上 | comic | 7 | 2026-10-29 03:09:04 |
| `book_2cab90477677843bbfc77929626829fe` | assets | 1 | 2026-11-03 02:08:20 |
| `book_2494d8f355e35664e0ceb79f1c6ef1dc` | assets | 1 | 2026-11-03 02:18:54 |
| `book_c02d71fc8900f2446a8e8d5c3280992f` | assets | 1 | 2026-11-03 02:32:15 |
| 同上 | assets | 2 | 2026-11-03 02:33:55 |

对应批次 job 均已是终态（`completed` / `failed` / `stopped`）；其中 27db、0a75 的书后来都跑到了 `completed`，b3ad 的三本书（2cab/2494/c02d）也都 `completed / published`。只有 `book_4090786e` 的书级 job 停在 `failed / comic_images`。

判定：**纯账本残留**。`book_stage_runs` 没有外键，软删字段也不存在；`SummaryWithAvailability` 只把它读出来展示（`backend/internal/batchfactoryv11/stage_service.go:172-181`），新一次运行的 attempt 取“同阶段最大值 + 1”（同文件 184-192），不因 running 阻塞。

## 三、处置建议

| 类别 | 建议 | 判据 | 预估条数 | 是否需授权 |
|---|---|---|---|---|
| A. 23 条 billed | **保留不动** | 页/图/合集都正常，删了只丢审计；活跃未取回为 0 | 0 条清理 | 若坚持清理：需 UPDATE `job_json`，须授权 |
| B. 6 页原“生成中” | **无需操作**（已回收） | 现为 `generation_failed`，可直接手动重试 | 0 条 | 重试出图需额度授权 |
| C. 30 页缺图 | 分两档：① 25 页 `generation_failed` → 可按书重试补图；② 4 页 `prompt_ready` + 1 页 `ready` → 可直接出图 | 页无图、无合集引用；其中 `fa42f92d`（唯爱遥遥）已发布却缺 1 页，优先补 | 30 页（优先 1 页已发布书） | 出图消耗额度，需授权 |
| D. 91 个无图资产 | **保留，按需补图**；不建议删除 | 有设定文本、当前零引用；后续补页/重跑需要参考图 | 91 个（可按书分批） | 补图需额度授权；删资产需写库授权 |
| E. 15 条 zombie stage_runs | **标记为 failed（回收成失败态）** | 批次已终态，running 只是历史假象；仅影响阶段列表展示 | 15 条 | 需 UPDATE 授权（不可逆） |

不做但需知道：多本书存在“同一本书多条 ready 合集”（如 `fa42f92d` 5 条 ×35、`d82908ae` 3 条、`269ea8a9` 2 条 ×6、`4090786e` 1 条仅 2 个条目）。这不是本次 5 类之一，但任何“按书清理”动作都要避开这些合集，否则会让书的前后状态更不一致。

## 四、不可逆操作与授权要求

本次未执行任何写操作。以下动作一旦执行，均不可通过回滚恢复，需要明确授权：

1. **任何 UPDATE / DELETE / INSERT**：包括把 `comic_pages.status` 从 `generation_failed` 改回可生成、把 `book_stage_runs.status` 从 running 改 failed、删除缺图页/无图资产。
   - `comic_pages`、`book_assets`、`book_asset_images`、`book_stage_runs` 都没有 `deleted_at` 软删列，删除即物理删除。
   - 改 `book_stage_runs` 还会让后续 attempt 编号回退（`nextAttempt = 现存最大值 + 1`），阶段历史出现断档。
2. **改 `batch_automation_jobs.job_json`**（例如清掉 23 条 `lastResolvedErrorCode`）：会覆盖整份账本快照，建议先备份整行再动；页面与合集不受影响。
3. **触发补图 / 重试**：会产生上游出图费用。按代码里的单页成本估算 `automationPageImageUnitCost = 0.03` 元/次，30 页约 ¥0.9、91 个资产按 1 次/个约 ¥2.7（仅为估算，实际以渠道计费为准），必须先授权。
4. **删除图片行**：SQL 删 `comic_page_images` / `book_asset_images` 不会删存储文件，会留下孤儿文件；且合集 `image_id` 可能悬空（当前缺图页未入合集，暂时不会悬空）。

## 五、只读核对口径（可复现）

```sql
-- 页状态
SELECT status, COUNT(*) FROM batch_factory_v12_comic_pages GROUP BY status;
-- 缺图页
SELECT p.* FROM batch_factory_v12_comic_pages p
LEFT JOIN batch_factory_v12_comic_page_images i ON i.comic_page_id=p.id
WHERE i.id IS NULL;
-- 无图资产
SELECT a.* FROM batch_factory_v12_book_assets a
LEFT JOIN batch_factory_v12_book_asset_images i ON i.asset_id=a.id
WHERE i.id IS NULL;
-- 已计费未取回（历史/活跃）
SELECT id, state FROM batch_automation_jobs WHERE job_json LIKE '%IMAGE_BILLED_NOT_RETRIEVED%';
SELECT COUNT(*) FROM batch_automation_jobs WHERE job_json LIKE '%"billedNotRetrieved":true%'; -- 0
-- zombie 阶段
SELECT * FROM batch_factory_v12_book_stage_runs WHERE status='running';
```

结论：当前真正需要“人来做决定”的只有 **C 类 30 页是否补图** 和 **D 类 91 个资产是否补图**；A、B 不需要清，E 是可选的一致性修复。
