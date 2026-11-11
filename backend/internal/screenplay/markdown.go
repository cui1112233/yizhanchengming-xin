package screenplay

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// ParsedScreenplay 是 screenplay.md 解析结果。
type ParsedScreenplay struct {
	Shots      []Shot
	Characters []Character
	Scenes     []Scene
	Episodes   []int
}

var (
	episodeHeadingRe = regexp.MustCompile(`^#\s*第\s*([0-9]+)\s*集\s*[:：]?\s*(.*)$`)
	sceneHeadingRe   = regexp.MustCompile(`^##\s*([0-9]+)\s*-\s*([0-9]+)\s*[:：]\s*(.*)$`)
	castLineRe       = regexp.MustCompile(`^人物\s*[:：]\s*(.*)$`)
	characterCardRe  = regexp.MustCompile(`^[（(]([^：:）)]{1,40})[：:](.*)[）)]\s*$`)
	dialogueRe       = regexp.MustCompile(`^([^\s（(：:【】△▲]{1,24})(\s*[（(][^）)]*[）)])?\s*[:：]\s*(.+)$`)
	actionRe         = regexp.MustCompile(`^[△▲]\s*`)
)

// ParseScreenplayMarkdown 解析 B 档格式（规格 §3.3 §5）：
//
//	# 第N集：标题
//	## N-N：时段 内/外 地点
//	人物：A B
//	（姓名：性别/年龄/外貌/性格）
//	△ 动作
//	角色（情绪）：台词
//
// 解析是容错的：缺集标题时按第 1 集处理，缺场头时该段不产出分场。
// TODO(推断): 原版 ensure_screenplay_db_from_md 的行级规则未在规格中逐条给出，
// 这里按提示词声明的格式实现；对照测试时需用真实产物补齐差异。
func ParseScreenplayMarkdown(content string) ParsedScreenplay {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	result := ParsedScreenplay{}
	episodeSet := map[int]bool{}
	characterIndex := map[string]int{}
	sceneIndex := map[string]int{}
	cardSeen := map[string]bool{}

	currentEpisode := 1
	var currentShot *Shot
	var currentBlock []string
	sortOrder := 0

	flush := func() {
		if currentShot != nil {
			currentShot.ScriptText = strings.TrimSpace(strings.Join(currentBlock, "\n"))
			result.Shots = append(result.Shots, *currentShot)
		}
		currentShot = nil
		currentBlock = nil
	}

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			if currentShot != nil {
				currentBlock = append(currentBlock, "")
			}
			continue
		}
		if match := episodeHeadingRe.FindStringSubmatch(line); match != nil {
			flush()
			episode, _ := strconv.Atoi(match[1])
			if episode > 0 {
				currentEpisode = episode
			}
			if !episodeSet[currentEpisode] {
				episodeSet[currentEpisode] = true
				result.Episodes = append(result.Episodes, currentEpisode)
			}
			continue
		}
		if match := sceneHeadingRe.FindStringSubmatch(line); match != nil {
			flush()
			episode, _ := strconv.Atoi(match[1])
			if episode > 0 {
				currentEpisode = episode
			}
			if !episodeSet[currentEpisode] {
				episodeSet[currentEpisode] = true
				result.Episodes = append(result.Episodes, currentEpisode)
			}
			location := extractLocation(match[3])
			sortOrder++
			currentShot = &Shot{
				EpisodeNo:         currentEpisode,
				SortOrder:         sortOrder,
				Heading:           strings.TrimSpace(match[3]),
				SceneName:         location,
				LocationReference: location,
			}
			if location != "" && sceneIndex[location] == 0 {
				sceneIndex[location] = len(result.Scenes) + 1
				result.Scenes = append(result.Scenes, Scene{
					Name:        location,
					Description: strings.TrimSpace(match[3]),
					SortOrder:   len(result.Scenes) + 1,
				})
			}
			continue
		}
		if match := castLineRe.FindStringSubmatch(line); match != nil {
			if currentShot != nil {
				currentBlock = append(currentBlock, line)
			}
			for _, name := range splitNames(match[1]) {
				addCharacter(&result, characterIndex, name, "")
				if currentShot != nil && !containsString(currentShot.CharacterNames, name) {
					currentShot.CharacterNames = append(currentShot.CharacterNames, name)
				}
			}
			continue
		}
		if match := characterCardRe.FindStringSubmatch(line); match != nil {
			name := strings.TrimSpace(match[1])
			desc := strings.TrimSpace(match[2])
			if !cardSeen[name] {
				cardSeen[name] = true
				addCharacter(&result, characterIndex, name, desc)
			}
			if currentShot != nil {
				currentBlock = append(currentBlock, line)
				if !containsString(currentShot.CharacterNames, name) {
					currentShot.CharacterNames = append(currentShot.CharacterNames, name)
				}
			}
			continue
		}
		if currentShot == nil {
			continue
		}
		currentBlock = append(currentBlock, line)
		if actionRe.MatchString(line) {
			continue
		}
		if match := dialogueRe.FindStringSubmatch(line); match != nil {
			speaker := strings.TrimSpace(match[1])
			if speaker != "" && !strings.ContainsAny(speaker, "，。！？") {
				addCharacter(&result, characterIndex, speaker, "")
				if !containsString(currentShot.CharacterNames, speaker) {
					currentShot.CharacterNames = append(currentShot.CharacterNames, speaker)
				}
			}
		}
	}
	flush()

	sort.Ints(result.Episodes)
	for i := range result.Characters {
		result.Characters[i].SortOrder = i + 1
	}
	for i := range result.Scenes {
		result.Scenes[i].SortOrder = i + 1
	}
	return result
}

func addCharacter(result *ParsedScreenplay, index map[string]int, name, description string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	if position, ok := index[name]; ok {
		if result.Characters[position-1].Description == "" && description != "" {
			result.Characters[position-1].Description = description
		}
		return
	}
	index[name] = len(result.Characters) + 1
	result.Characters = append(result.Characters, Character{Name: name, Description: description})
}

func splitNames(raw string) []string {
	names := make([]string, 0, 4)
	for _, field := range strings.Fields(raw) {
		for _, part := range strings.FieldsFunc(field, func(r rune) bool {
			switch r {
			case '，', ',', '、', '；', ';', '/', '|':
				return true
			}
			return false
		}) {
			name := strings.TrimSpace(part)
			if name != "" {
				names = append(names, name)
			}
		}
	}
	return names
}

// extractLocation 从「日 内 蓝雨传媒28楼·经纪人办公室」中取地点部分。
func extractLocation(heading string) string {
	heading = strings.TrimSpace(heading)
	if heading == "" {
		return ""
	}
	fields := strings.Fields(heading)
	if len(fields) <= 2 {
		return heading
	}
	return strings.Join(fields[2:], " ")
}
