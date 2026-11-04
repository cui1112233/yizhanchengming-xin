// Package petmigrate 提供桌宠默认值的一次性迁移能力。
//
// 已停止使用：用户已确认改为“只对未设置 / 未知宠物 id 在读取期回落到
// little-companion，已保存的 stacky / pixiu / CM / 旧对象格式一律保留原值”。
// 因此本包默认口径 PolicyPreserveExplicit 不产生任何写盘需求，查出的“待迁移”
// 恒为 0；PolicyRewriteAll 只是历史口径，仅保留审计与回滚依据。
// 仅在明确批准时才允许执行改写，且本机 CLI 已不再提供写盘入口，默认只读。
package petmigrate

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// DefaultPetID 与前端 petCatalog.js 的 DEFAULT_PET_ID、
// 后端 user_config.go 的 defaultPetSetting 保持一致。
const DefaultPetID = "little-companion"

// ConfigFileName 与 UserConfigStore.configPath 保持一致。
const ConfigFileName = "api-config.json"

const (
	KindMissing = "missing"
	KindString  = "string"
	KindObject  = "object"
	KindInvalid = "invalid"
	KindOther   = "other"
)

// Policy 是扫描口径。
type Policy string

const (
	// PolicyPreserveExplicit 是当前生效口径：已保存的明确选择一律保留，
	// 只有未设置 / 未知 id 在读取期回落到 little-companion，不需要改写任何文件。
	PolicyPreserveExplicit Policy = "preserve-explicit"
	// PolicyRewriteAll 是已作废的历史口径，仅供审计与回滚复核，不得用于生产。
	PolicyRewriteAll Policy = "rewrite-all"
)

// Record 描述单个用户配置文件的宠物状态。
type Record struct {
	Source     string
	Username   string
	Path       string
	BeforeKind string
	BeforeID   string
	AfterID    string
	Changed    bool
	Err        string
}

// Scan 按当前口径（PolicyPreserveExplicit）只读扫描指定用户目录。
func Scan(source, usersDir string) ([]Record, error) {
	return ScanWithPolicy(source, usersDir, PolicyPreserveExplicit)
}

// ScanWithPolicy 按指定口径只读扫描；不会写任何文件。
func ScanWithPolicy(source, usersDir string, policy Policy) ([]Record, error) {
	entries, err := os.ReadDir(usersDir)
	if err != nil {
		return nil, err
	}
	var records []Record
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		username := entry.Name()
		path := filepath.Join(usersDir, username, ConfigFileName)
		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			if os.IsNotExist(readErr) {
				continue
			}
			records = append(records, Record{Source: source, Username: username, Path: path, BeforeKind: KindInvalid, Err: readErr.Error()})
			continue
		}
		var config map[string]any
		if json.Unmarshal(raw, &config) != nil {
			records = append(records, Record{Source: source, Username: username, Path: path, BeforeKind: KindInvalid, Err: "JSON 解析失败"})
			continue
		}
		record := Record{Source: source, Username: username, Path: path}
		kind, id := classifyPet(config)
		record.BeforeKind = kind
		record.BeforeID = id
		if policy == PolicyRewriteAll {
			record.Changed = kind != KindMissing && kind != KindInvalid && id != DefaultPetID
			if !record.Changed && kind != KindMissing && kind != KindInvalid {
				record.AfterID = DefaultPetID
			}
		} else {
			// 当前口径：已保存的选择永不被改写；未设置/空值读取时回落默认。
			if kind == KindMissing || id == "" {
				record.AfterID = DefaultPetID
			} else {
				record.AfterID = id
			}
		}
		records = append(records, record)
	}
	sort.Slice(records, func(i, j int) bool {
		if records[i].Source != records[j].Source {
			return records[i].Source < records[j].Source
		}
		return records[i].Username < records[j].Username
	})
	return records, nil
}

func classifyPet(config map[string]any) (string, string) {
	value, ok := config["pet"]
	if !ok || value == nil {
		return KindMissing, ""
	}
	switch typed := value.(type) {
	case string:
		return KindString, typed
	case map[string]any:
		return KindObject, strings.TrimSpace(fmt.Sprint(typed["id"]))
	default:
		return KindOther, fmt.Sprint(value)
	}
}

// Migrate 按记录中的 Changed 备份并改写文件。已停止使用：只接受
// PolicyRewriteAll 口径产生记录，且仅限明确批准的历史回滚场景；CLI 不再暴露写盘入口。
func Migrate(records []Record, backupRoot, timestamp string, dryRun bool) (backupPath string, migrated int, err error) {
	if dryRun {
		for _, record := range records {
			if record.Changed {
				migrated++
			}
		}
		return "", migrated, nil
	}
	affected := 0
	for _, record := range records {
		if record.Changed {
			affected++
		}
	}
	if affected == 0 {
		return "", 0, nil
	}
	if strings.TrimSpace(backupRoot) == "" {
		return "", 0, fmt.Errorf("备份目录不能为空")
	}
	backupPath = filepath.Join(backupRoot, timestamp)
	if err := os.MkdirAll(backupPath, 0o700); err != nil {
		return "", 0, err
	}
	for index := range records {
		record := &records[index]
		if !record.Changed || record.Err != "" {
			continue
		}
		if err := migrateFile(record, backupPath); err != nil {
			record.Err = err.Error()
			continue
		}
		record.AfterID = DefaultPetID
		migrated++
	}
	return backupPath, migrated, nil
}

func migrateFile(record *Record, backupPath string) error {
	raw, err := os.ReadFile(record.Path)
	if err != nil {
		return err
	}
	backupFile := filepath.Join(backupPath, record.Username, ConfigFileName)
	if err := os.MkdirAll(filepath.Dir(backupFile), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(backupFile, raw, 0o600); err != nil {
		return err
	}
	var config map[string]any
	if err := json.Unmarshal(raw, &config); err != nil {
		return err
	}
	config["pet"] = DefaultPetID
	encoded, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	tmp := record.Path + ".tmp"
	if err := os.WriteFile(tmp, encoded, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, record.Path)
}

// Counts 返回迁移前的分布：键为宠物 id，未设置为“未设置”。
func Counts(records []Record) map[string]int {
	counts := map[string]int{}
	for _, record := range records {
		switch record.BeforeKind {
		case KindMissing:
			counts["未设置"]++
		case KindInvalid:
			counts["读取失败"]++
		default:
			id := record.BeforeID
			if id == "" {
				id = "空值"
			}
			counts[id]++
		}
	}
	return counts
}

// CSV 生成逐用户清单。
func CSV(records []Record) string {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)
	_ = writer.Write([]string{"source", "username", "path", "before_kind", "before_pet_id", "after_pet_id", "action", "error"})
	for _, record := range records {
		action := "保持"
		if record.Changed {
			action = "迁移"
		}
		if record.Err != "" {
			action = "失败"
		}
		after := record.AfterID
		before := record.BeforeID
		if before == "" {
			before = "未设置"
		}
		_ = writer.Write([]string{record.Source, record.Username, record.Path, record.BeforeKind, before, after, action, record.Err})
	}
	writer.Flush()
	return builder.String()
}

// Markdown 生成带汇总与回滚说明的迁移清单。
func Markdown(records []Record, backupPath string, dryRun bool, migrated int) string {
	mode := "真实迁移"
	if dryRun {
		mode = "dry-run（只读，不写文件）"
	}
	var builder strings.Builder
	builder.WriteString("# 2026-11-04 桌宠默认值迁移清单\n\n")
	builder.WriteString("> 已作废：用户决定仅对未设置用户生效，本次迁移已回滚，生产未执行\n\n")
	builder.WriteString("- 模式：" + mode + "\n")
	builder.WriteString("- 当前口径：已保存的 stacky / pixiu / CM / 旧对象格式一律保留原值；仅未设置 / 未知 id 在读取期回落 `" + DefaultPetID + "`\n")
	builder.WriteString("- 目标默认宠物：" + DefaultPetID + "（小家伙）\n")
	builder.WriteString("- 历史说明：早期版本曾一次性改写 `api-config.json` 中的 `pet`，该做法已停止使用并回滚\n\n")
	builder.WriteString("## 1. 汇总\n\n")
	builder.WriteString("| 用户目录 | 用户数 | 受影响（需迁移） | 保留原值 | 未设置 | 读取失败 |\n")
	builder.WriteString("| --- | --- | --- | --- | --- | --- |\n")
	sources := map[string][]Record{}
	for _, record := range records {
		sources[record.Source] = append(sources[record.Source], record)
	}
	sourceNames := make([]string, 0, len(sources))
	for name := range sources {
		sourceNames = append(sourceNames, name)
	}
	sort.Strings(sourceNames)
	totalAffected := 0
	for _, name := range sourceNames {
		items := sources[name]
		affected, already, missing, invalid := 0, 0, 0, 0
		for _, item := range items {
			switch {
			case item.Err != "" || item.BeforeKind == KindInvalid:
				invalid++
			case item.BeforeKind == KindMissing:
				missing++
			case item.Changed:
				affected++
			default:
				already++
			}
		}
		totalAffected += affected
		builder.WriteString("| " + name + " | " + strconv.Itoa(len(items)) + " | " + strconv.Itoa(affected) + " | " +
			strconv.Itoa(already) + " | " + strconv.Itoa(missing) + " | " + strconv.Itoa(invalid) + " |\n")
	}
	builder.WriteString("\n本次实际迁移条数：" + strconv.Itoa(migrated) + "（待迁移合计 " + strconv.Itoa(totalAffected) + "）\n")

	builder.WriteString("\n## 2. 当前宠物分布\n\n| 宠物值 | 人数 |\n| --- | --- |\n")
	counts := Counts(records)
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		builder.WriteString("| " + key + " | " + strconv.Itoa(counts[key]) + " |\n")
	}

	builder.WriteString("\n## 3. 读取后宠物分布\n\n")
	if dryRun {
		builder.WriteString("dry-run 未写文件；按当前口径，已保存选择保持原值，未设置/未知 id 读取时回落 `" + DefaultPetID + "`。\n")
	} else {
		after := map[string]int{}
		for _, record := range records {
			if record.Err != "" {
				after["失败"]++
				continue
			}
			id := record.AfterID
			if id == "" {
				id = DefaultPetID
			}
			after[id]++
		}
		afterKeys := make([]string, 0, len(after))
		for key := range after {
			afterKeys = append(afterKeys, key)
		}
		sort.Strings(afterKeys)
		builder.WriteString("| 宠物值 | 人数 |\n| --- | --- |\n")
		for _, key := range afterKeys {
			builder.WriteString("| " + key + " | " + strconv.Itoa(after[key]) + " |\n")
		}
	}

	builder.WriteString("\n## 4. 备份与回滚\n\n")
	if backupPath == "" {
		builder.WriteString("- 备份：dry-run 未产生备份\n")
	} else {
		builder.WriteString("- 备份目录：" + backupPath + "\n")
		builder.WriteString("- 回滚：把备份目录下 `<用户名>/api-config.json` 覆盖回原路径即可恢复迁移前状态\n")
	}
	builder.WriteString("- 幂等：当前口径下扫描结果恒为“待迁移 0 条”，不会产生任何写盘\n")
	builder.WriteString("- 用户经 `POST /api/config` 改成 `stacky`/`pixiu` 会正常保存，刷新后仍是其选择\n")

	return builder.String()
}
