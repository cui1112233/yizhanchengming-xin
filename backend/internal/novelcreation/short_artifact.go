package novelcreation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// Editable keys are existing logical artifacts, never filesystem paths. Waiting
// blueprints feed the next run; completed manuscript edits do not rerun models.
func shortArtifactEditable(story ShortStory, key string) bool {
	if story.Status == "awaiting_confirmation" {
		return story.Checkpoint == "blueprint" && key == "imitation/blueprint.md"
	}
	if story.Status != "completed" {
		return false
	}
	switch key {
	case "imitation/deconstruction.md", "imitation/directions.md", "imitation/selected-direction.md", "imitation/blueprint.md", "outline/story.json", "recommendation.txt":
		return true
	}
	for _, chapter := range story.Chapters {
		if key == fmt.Sprintf("chapters/%03d.txt", chapter.Number) {
			return true
		}
	}
	return false
}

func decorateShortArtifacts(story *ShortStory) {
	for i := range story.Artifacts {
		story.Artifacts[i].Editable = shortArtifactEditable(*story, story.Artifacts[i].Key)
	}
}

func (s *Service) EditShortStoryArtifact(ctx context.Context, owner, project string, version int, key, content string) (out ShortStory, err error) {
	owner, err = requiredOwner(owner)
	if err != nil {
		return
	}
	err = s.store.Atomic(ctx, owner, project, func(tx Store) error {
		child := *s
		child.store = tx
		story, e := child.GetShortStory(ctx, owner, project)
		if e != nil {
			return e
		}
		if story.Version != version {
			return conflict()
		}
		if _, e = writableSession(ctx, tx, owner, project, story.SessionID); e != nil {
			return e
		}
		run, e := tx.GetRun(ctx, owner, story.RunID)
		if e != nil {
			return e
		}
		if run.SessionID != story.SessionID || run.ProjectID != project {
			return conflict()
		}
		if !(story.Status == "completed" && run.Status == RunSucceeded) && !(story.Status == "awaiting_confirmation" && run.Status == RunAwaitingConfirmation) {
			return inputError("任务执行期间不能编辑文件")
		}
		index := -1
		for i, a := range story.Artifacts {
			if a.Key == key {
				index = i
				break
			}
		}
		if index < 0 || !shortArtifactEditable(story, key) {
			return inputError("此文件当前不可编辑")
		}
		if strings.TrimSpace(content) == "" || jsLength(content) > LongImitationMaxUTF16 {
			return inputError("文件内容为空或超出上限")
		}
		if story.Checkpoint == "blueprint" {
			var protocol struct {
				Version int `json:"short_imitation_version"`
			}
			if e = json.Unmarshal(run.PromptSnapshot, &protocol); e != nil {
				return e
			}
			if protocol.Version != 2 {
				return inputError("蓝图编辑协议无效")
			}
		}
		switch key {
		case "imitation/deconstruction.md", "imitation/directions.md", "imitation/selected-direction.md", "imitation/blueprint.md":
			if story.Imitation == nil {
				return inputError("仿写阶段资料缺失")
			}
			switch key {
			case "imitation/deconstruction.md":
				story.Imitation.Deconstruction = content
			case "imitation/directions.md":
				story.Imitation.Directions = content
			case "imitation/selected-direction.md":
				story.Imitation.SelectedDirection = content
			case "imitation/blueprint.md":
				// The Markdown editor cannot update structured scene budgets. Keep
				// those plan lines intact rather than sending conflicting versions
				// of the same plan to the writer after confirmation.
				for _, line := range strings.Split(story.Imitation.Blueprint, "\n") {
					for _, scene := range story.Imitation.Scenes {
						if strings.Contains(line, scene.Name) && !strings.Contains(content, strings.TrimSpace(line)) {
							return inputError("请保留蓝图原有分段规划行（名称和预计字数）；可编辑其他说明或补充内容")
						}
					}
				}
				story.Imitation.Blueprint = content
			}
			if e = validateSavedShortImitation(story.Imitation, story.Input.TargetWordCount); e != nil {
				return inputError("编辑内容须保留既定分段名称、字数与事件映射及有效报告结构：" + e.Error())
			}
		case "outline/story.json":
			var outline ShortOutline
			decoder := json.NewDecoder(bytes.NewBufferString(content))
			decoder.DisallowUnknownFields()
			if e = decoder.Decode(&outline); e != nil {
				return inputError("大纲JSON格式无效")
			}
			if e = decoder.Decode(new(any)); e != io.EOF {
				return inputError("大纲JSON之后不能附加其他内容")
			}
			if e = validateShortOutline(outline, story.Input.TargetChapters); e != nil {
				return e
			}
			story.Outline = &outline
			story.Title = outline.Title
			if _, e = tx.RenameProject(ctx, owner, project, outline.Title); e != nil {
				return e
			}
		case "recommendation.txt":
			if e = validShortText(shortText{content}, 8000); e != nil {
				return e
			}
			story.Recommendation = content
		default:
			for i, c := range story.Chapters {
				if key == fmt.Sprintf("chapters/%03d.txt", c.Number) {
					story.Chapters[i].Content = content
				}
			}
		}
		story.Artifacts[index].Content = content
		if story.FullDraft != nil && (key == "outline/story.json" || strings.HasPrefix(key, "chapters/")) {
			if e = refreshShortFullCurrentArtifact(&story); e != nil {
				return e
			}
		}
		if story.Status == "completed" && key != "recommendation.txt" {
			story.QualityStale = true
		}
		story.Version++
		story.UpdatedAt = time.Now().UTC()
		decorateShortArtifacts(&story)
		if e = putContent(ctx, tx, owner, project, shortStoryKind, "story", story.Status, time.Time{}, story); e != nil {
			return e
		}
		out = story
		return nil
	})
	return
}
