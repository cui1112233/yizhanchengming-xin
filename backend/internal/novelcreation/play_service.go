package novelcreation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func (s *Service) callPlayJSON(ctx context.Context, owner, modelID, systemPrompt, userPrompt string, maxTokens int, temperature float64, target any) error {
	if s.playText == nil {
		return inputError("未配置互动世界模型，无法推进剧情")
	}
	raw, err := s.playText.CompletePlayText(ctx, owner, modelID, systemPrompt, strings.TrimSpace(userPrompt), maxTokens, temperature)
	if err != nil {
		return err
	}
	if err = decodePlayJSON(raw, target); err != nil {
		return &DomainError{Code: ErrInvalidOutput, Message: "互动世界返回格式无效", Cause: err}
	}
	return nil
}

func playDefinitionPrompt(systemPrompt, outputContract string) string {
	return strings.TrimSpace(systemPrompt) + "\n\n输出合同（必须严格遵守）：\n" + strings.TrimSpace(outputContract)
}

func (s *Service) playProject(ctx context.Context, owner, projectID string) (Project, error) {
	if s == nil || s.store == nil {
		return Project{}, &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	return s.store.GetProject(ctx, strings.TrimSpace(owner), strings.TrimSpace(projectID))
}

// CreatePlayWorld 建立互动世界并生成开场。
func (s *Service) CreatePlayWorld(ctx context.Context, owner, projectID string, input PlayWorldInput) (out PlayWorld, opening PlayTurn, err error) {
	project, err := s.playProject(ctx, owner, projectID)
	if err != nil {
		return out, opening, err
	}
	premise := strings.TrimSpace(input.Premise)
	if premise == "" {
		return out, opening, inputError("互动世界 premise 不能为空")
	}
	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = "未命名世界"
	}
	mode := strings.TrimSpace(input.Mode)
	if mode != "guided" {
		mode = playDefaultMode
	}
	world := PlayWorld{
		ProjectID: strings.TrimSpace(projectID), ID: playID(title), Title: title, Premise: premise, Mode: mode,
		PlayerPersona: strings.TrimSpace(input.PlayerPersona), WorldContract: strings.TrimSpace(input.WorldContract),
		Entities: []PlayEntity{}, Relations: []PlayRelation{}, Facts: []string{}, Goals: []string{}, Turns: []PlayTurn{},
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	var openingOut playOpeningOutput
	user := fmt.Sprintf("世界：%s\n设定：%s\n模式：%s", world.Title, world.Premise, world.Mode)
	if project.Genre != "" {
		user += "\n题材：" + project.Genre
	}
	if err = s.callPlayJSON(ctx, owner, project.ModelID, playDefinitionPrompt(PlayOpening.SystemPrompt, PlayOpening.OutputContract), user, 2500, 0.85, &openingOut); err != nil {
		return out, opening, err
	}
	scene := strings.TrimSpace(openingOut.SceneText)
	if scene == "" {
		return out, opening, &DomainError{Code: ErrInvalidOutput, Message: "开场正文为空"}
	}
	opening = PlayTurn{Turn: 0, Input: "(opening)", SceneText: scene, SuggestedActions: normalizePlayActions(openingOut.SuggestedActions), CreatedAt: time.Now().UTC()}
	world.Turns = append(world.Turns, opening)
	if err = s.savePlayWorld(ctx, owner, world); err != nil {
		return out, opening, err
	}
	return world, opening, nil
}

// StepPlayWorld 推进一幕：推演 → 渲染 → 同步状态。
func (s *Service) StepPlayWorld(ctx context.Context, owner, projectID, worldID, playerInput string) (world PlayWorld, turn PlayTurn, err error) {
	project, err := s.playProject(ctx, owner, projectID)
	if err != nil {
		return world, turn, err
	}
	world, err = s.GetPlayWorld(ctx, owner, projectID, worldID)
	if err != nil {
		return world, turn, err
	}
	input := strings.TrimSpace(playerInput)
	if input == "" {
		return world, turn, inputError("行动输入不能为空")
	}
	nextTurn := world.Turn + 1
	summary := playStateSummary(world)
	previous := playSceneExcerpt(world)

	simulateUser := fmt.Sprintf("世界设定：%s\n当前状态：\n%s\n\n上一幕摘录：\n%s\n\n玩家行动：%s\n\n这是第 %d 幕。",
		world.Premise, summary, previous, input, nextTurn)
	var simulation playSimulationOutput
	if err = s.callPlayJSON(ctx, owner, project.ModelID, playDefinitionPrompt(PlaySimulate.SystemPrompt, PlaySimulate.OutputContract), simulateUser, 4096, 0.25, &simulation); err != nil {
		return world, turn, err
	}
	outcome := strings.TrimSpace(simulation.Outcome)
	if simulation.Blocked {
		if reason := strings.TrimSpace(simulation.BlockedReason); reason != "" {
			outcome = "行动受阻：" + reason
		}
	}
	if outcome == "" {
		outcome = "行动已发生，结果尚未落定。"
	}

	narrateUser := fmt.Sprintf("世界设定：%s\n当前状态：\n%s\n\n上一幕摘录：\n%s\n\n玩家行动：%s\n\n推演结果：%s",
		world.Premise, summary, previous, input, truncatePlayText(outcome, 1200))
	var narration playNarrationOutput
	if err = s.callPlayJSON(ctx, owner, project.ModelID, playDefinitionPrompt(PlayNarrate.SystemPrompt, PlayNarrate.OutputContract), narrateUser, 4096, 0.85, &narration); err != nil {
		return world, turn, err
	}
	scene := strings.TrimSpace(narration.SceneText)
	if scene == "" {
		scene = "（这一拍悬着，没有落定。）"
	}

	stateUser := fmt.Sprintf("这是第 %d 幕。\n\n当前状态：\n%s\n\n场景正文：\n%s\n\n推演结果：%s",
		nextTurn, summary, truncatePlayText(scene, 3500), truncatePlayText(outcome, 800))
	var state playStateOutput
	if err = s.callPlayJSON(ctx, owner, project.ModelID, playDefinitionPrompt(PlayStateUpdate.SystemPrompt, PlayStateUpdate.OutputContract), stateUser, 2048, 0.1, &state); err != nil {
		return world, turn, err
	}

	turn = PlayTurn{
		Turn: nextTurn, Input: input, ActionKind: strings.TrimSpace(simulation.ActionKind), Outcome: outcome,
		SceneText: scene, SuggestedActions: normalizePlayActions(narration.SuggestedActions), CreatedAt: time.Now().UTC(),
	}
	world.Turn = nextTurn
	if location := strings.TrimSpace(state.Location); location != "" {
		world.Location = location
	} else if location := strings.TrimSpace(simulation.Location); location != "" {
		world.Location = location
	}
	if clock := strings.TrimSpace(state.Clock); clock != "" {
		world.Clock = clock
	} else if clock := strings.TrimSpace(simulation.Clock); clock != "" {
		world.Clock = clock
	}
	world.Entities = mergePlayEntities(world.Entities, append(append([]PlayEntity{}, simulation.EntityChanges...), state.Entities...))
	world.Relations = mergePlayRelations(world.Relations, append(append([]PlayRelation{}, simulation.RelationChanges...), state.Relations...))
	world.Facts = mergePlayStrings(world.Facts, append(append([]string{}, simulation.NewFacts...), state.Facts...))
	world.Goals = mergePlayStrings(world.Goals, append(append([]string{}, simulation.Goals...), state.Goals...))
	world.Turns = append(world.Turns, turn)
	if len(world.Turns) > playTurnLimit {
		world.Turns = world.Turns[len(world.Turns)-playTurnLimit:]
	}
	world.UpdatedAt = time.Now().UTC()
	if err = s.savePlayWorld(ctx, owner, world); err != nil {
		return world, turn, err
	}
	return world, turn, nil
}

// EditPlayWorld 直接修改世界设定，不调用模型。
func (s *Service) EditPlayWorld(ctx context.Context, owner, projectID, worldID string, patch PlayEditPatch) (PlayWorld, error) {
	var world PlayWorld
	if _, err := s.playProject(ctx, owner, projectID); err != nil {
		return world, err
	}
	var err error
	world, err = s.GetPlayWorld(ctx, owner, projectID, worldID)
	if err != nil {
		return world, err
	}
	if title := strings.TrimSpace(patch.Title); title != "" {
		world.Title = title
	}
	if premise := strings.TrimSpace(patch.Premise); premise != "" {
		world.Premise = premise
	}
	if persona := strings.TrimSpace(patch.PlayerPersona); persona != "" {
		world.PlayerPersona = persona
	}
	if contract := strings.TrimSpace(patch.WorldContract); contract != "" {
		world.WorldContract = contract
	}
	if visual := strings.TrimSpace(patch.VisualContract); visual != "" {
		world.VisualContract = visual
	}
	world.UpdatedAt = time.Now().UTC()
	if err = s.savePlayWorld(ctx, owner, world); err != nil {
		return world, err
	}
	return world, nil
}

func (s *Service) savePlayWorld(ctx context.Context, owner string, world PlayWorld) error {
	return putContent(ctx, s.store, strings.TrimSpace(owner), world.ProjectID, playWorldKind, world.ID, "ready", time.Time{}, world)
}

// GetPlayWorld 读取世界与全部已保存回合。
func (s *Service) GetPlayWorld(ctx context.Context, owner, projectID, worldID string) (PlayWorld, error) {
	var world PlayWorld
	if _, err := s.playProject(ctx, owner, projectID); err != nil {
		return world, err
	}
	found, err := readContent(ctx, s.store, strings.TrimSpace(owner), strings.TrimSpace(projectID), playWorldKind, strings.TrimSpace(worldID), &world)
	if err != nil {
		return world, err
	}
	if !found {
		return world, &DomainError{Code: ErrProjectNotFound, Message: "互动世界不存在: " + strings.TrimSpace(worldID)}
	}
	if world.Turns == nil {
		world.Turns = []PlayTurn{}
	}
	return world, nil
}

// ListPlayWorlds 按更新时间倒序返回项目下的互动世界。
func (s *Service) ListPlayWorlds(ctx context.Context, owner, projectID string) ([]PlayWorld, error) {
	if _, err := s.playProject(ctx, owner, projectID); err != nil {
		return nil, err
	}
	rows, err := s.store.ListContentRecords(ctx, strings.TrimSpace(owner), strings.TrimSpace(projectID), playWorldKind)
	if err != nil {
		return nil, err
	}
	out := []PlayWorld{}
	for _, row := range rows {
		var world PlayWorld
		if err = json.Unmarshal(row.Data, &world); err != nil {
			return nil, persistenceError(err)
		}
		out = append(out, world)
	}
	return out, nil
}
