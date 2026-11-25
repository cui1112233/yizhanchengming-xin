package novelcreation

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type stubPlayText struct {
	responses []string
	err       error
	calls     int
	user      []string
	system    []string
}

func (s *stubPlayText) CompletePlayText(_ context.Context, _, _, systemPrompt, userPrompt string, _ int, _ float64) (string, error) {
	s.calls++
	s.system = append(s.system, systemPrompt)
	s.user = append(s.user, userPrompt)
	if s.err != nil {
		return "", s.err
	}
	if len(s.responses) == 0 {
		return "", errors.New("no queued play response")
	}
	next := s.responses[0]
	s.responses = s.responses[1:]
	return next, nil
}

func playFixture(t *testing.T) (*Service, Project, *stubPlayText) {
	t.Helper()
	service := NewService(NewMemoryStore(), NewMemoryRuntime(), nil)
	project, err := service.CreateProject(context.Background(), "alice", CreateProjectInput{
		Title: "雾港来信", Genre: "悬疑", World: "常年海雾的港城", CoreConflict: "真相与沉默", ModelID: "text-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	text := &stubPlayText{}
	service.ConfigurePlayText(text)
	return service, project, text
}

const playOpeningJSON = `{"sceneText":"雾从码头漫上来，你在废弃仓库门口醒来。","suggestedActions":["看向灯塔","检查口袋"]}`

func createPlayWorld(t *testing.T, service *Service, project Project, text *stubPlayText) PlayWorld {
	t.Helper()
	text.responses = append(text.responses, playOpeningJSON)
	world, opening, err := service.CreatePlayWorld(context.Background(), "alice", project.ID, PlayWorldInput{
		Title: "雾港灯塔", Premise: "你在雾港醒来，只剩一张写着灯塔的纸条。", PlayerPersona: "失忆的来客",
	})
	if err != nil {
		t.Fatal(err)
	}
	if opening.Turn != 0 || strings.TrimSpace(opening.SceneText) == "" {
		t.Fatalf("opening=%+v", opening)
	}
	return world
}

func TestCreatePlayWorldRequiresPremise(t *testing.T) {
	service, project, text := playFixture(t)
	_, _, err := service.CreatePlayWorld(context.Background(), "alice", project.ID, PlayWorldInput{Title: "空世界"})
	if !IsCode(err, ErrInputRequired) || !strings.Contains(err.Error(), "premise") {
		t.Fatalf("err=%v", err)
	}
	if text.calls != 0 {
		t.Fatalf("model must not be called, calls=%d", text.calls)
	}
}

func TestCreatePlayWorldRequiresCompleter(t *testing.T) {
	service := NewService(NewMemoryStore(), NewMemoryRuntime(), nil)
	project, err := service.CreateProject(context.Background(), "alice", CreateProjectInput{Title: "未接模型", Genre: "悬疑", ModelID: "text-1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = service.CreatePlayWorld(context.Background(), "alice", project.ID, PlayWorldInput{Premise: "开场"}); err == nil {
		t.Fatal("missing completer must fail")
	}
}

func TestCreatePlayWorldSavesOpeningAndLists(t *testing.T) {
	service, project, text := playFixture(t)
	world := createPlayWorld(t, service, project, text)
	if world.ProjectID != project.ID || world.Turn != 0 || len(world.Turns) != 1 {
		t.Fatalf("world=%+v", world)
	}
	if world.Mode != playDefaultMode || world.ID == "" || world.Title != "雾港灯塔" {
		t.Fatalf("world meta=%+v", world)
	}
	actions := world.Turns[0].SuggestedActions
	if len(actions) != 2 || actions[0] != "看向灯塔" {
		t.Fatalf("actions=%v", actions)
	}
	if text.calls != 1 || !strings.Contains(text.user[0], "雾港灯塔") {
		t.Fatalf("calls=%d user=%v", text.calls, text.user)
	}
	read, err := service.GetPlayWorld(context.Background(), "alice", project.ID, world.ID)
	if err != nil || read.ID != world.ID || len(read.Turns) != 1 {
		t.Fatalf("read=%+v err=%v", read, err)
	}
	list, err := service.ListPlayWorlds(context.Background(), "alice", project.ID)
	if err != nil || len(list) != 1 || list[0].ID != world.ID {
		t.Fatalf("list=%+v err=%v", list, err)
	}
}

func TestCreatePlayWorldWithSameTitleKeepsBothWorlds(t *testing.T) {
	service, project, text := playFixture(t)
	first := createPlayWorld(t, service, project, text)
	second := createPlayWorld(t, service, project, text)
	if first.ID == second.ID {
		t.Fatalf("同名世界不能互相覆盖: %q", first.ID)
	}
	list, err := service.ListPlayWorlds(context.Background(), "alice", project.ID)
	if err != nil || len(list) != 2 {
		t.Fatalf("list=%d err=%v", len(list), err)
	}
}

func TestStepPlayWorldMergesSimulationAndState(t *testing.T) {
	service, project, text := playFixture(t)
	world := createPlayWorld(t, service, project, text)
	text.responses = append(text.responses,
		`{"actionKind":"move","intent":"去灯塔","outcome":"你走到灯塔下，门虚掩着。","location":"灯塔","clock":"午夜","blocked":false,"entityChanges":[{"name":"守塔人","type":"npc","status":"沉睡"}],"relationChanges":[{"from":"我","to":"守塔人","type":"警惕"}],"newFacts":["灯塔的门虚掩着"],"goals":["进入灯塔"]}`,
		"```json\n"+`{"sceneText":"灯塔的门在风里晃动。","suggestedActions":["推门进去","敲门"]}`+"\n```",
		`{"location":"灯塔门口","clock":"午夜过后","entities":[{"name":"守塔人","status":"醒来","location":"塔内"}],"relations":[{"from":"我","to":"守塔人","type":"警惕"}],"facts":["灯塔的门虚掩着","塔内有灯"],"goals":["找到守塔人"]}`,
	)
	before := text.calls
	stepped, turn, err := service.StepPlayWorld(context.Background(), "alice", project.ID, world.ID, "走向灯塔")
	if err != nil {
		t.Fatal(err)
	}
	if text.calls-before != 3 {
		t.Fatalf("step must call model 3 times, got %d", text.calls-before)
	}
	if turn.Turn != 1 || turn.ActionKind != "move" || turn.SceneText != "灯塔的门在风里晃动。" {
		t.Fatalf("turn=%+v", turn)
	}
	if stepped.Turn != 1 || len(stepped.Turns) != 2 {
		t.Fatalf("world turn=%d turns=%d", stepped.Turn, len(stepped.Turns))
	}
	if stepped.Location != "灯塔门口" || stepped.Clock != "午夜过后" {
		t.Fatalf("location=%q clock=%q", stepped.Location, stepped.Clock)
	}
	if len(stepped.Entities) != 1 {
		t.Fatalf("entities=%+v", stepped.Entities)
	}
	entity := stepped.Entities[0]
	if entity.Name != "守塔人" || entity.Type != "npc" || entity.Status != "醒来" || entity.Location != "塔内" {
		t.Fatalf("entity=%+v", entity)
	}
	if len(stepped.Relations) != 1 || stepped.Relations[0].Type != "警惕" {
		t.Fatalf("relations=%+v", stepped.Relations)
	}
	if len(stepped.Facts) != 2 || len(stepped.Goals) != 2 {
		t.Fatalf("facts=%v goals=%v", stepped.Facts, stepped.Goals)
	}
	if len(turn.SuggestedActions) != 2 {
		t.Fatalf("actions=%v", turn.SuggestedActions)
	}
}

func TestStepPlayWorldRequiresInputAndExistingWorld(t *testing.T) {
	service, project, text := playFixture(t)
	world := createPlayWorld(t, service, project, text)
	before := text.calls
	if _, _, err := service.StepPlayWorld(context.Background(), "alice", project.ID, world.ID, "   "); !IsCode(err, ErrInputRequired) {
		t.Fatalf("err=%v", err)
	}
	if _, _, err := service.StepPlayWorld(context.Background(), "alice", project.ID, "missing-world", "前进"); !IsCode(err, ErrProjectNotFound) {
		t.Fatalf("err=%v", err)
	}
	if text.calls != before {
		t.Fatalf("model must not be called, delta=%d", text.calls-before)
	}
}

func TestEditPlayWorldDoesNotCallModel(t *testing.T) {
	service, project, text := playFixture(t)
	world := createPlayWorld(t, service, project, text)
	before := text.calls
	updated, err := service.EditPlayWorld(context.Background(), "alice", project.ID, world.ID, PlayEditPatch{
		Title: "雾港灯塔·改", Premise: "新的前提", Note: "备注不落库",
	})
	if err != nil {
		t.Fatal(err)
	}
	if text.calls != before {
		t.Fatalf("edit must not call model, delta=%d", text.calls-before)
	}
	if updated.Title != "雾港灯塔·改" || updated.Premise != "新的前提" || len(updated.Turns) != 1 {
		t.Fatalf("updated=%+v", updated)
	}
}

func TestPlayWorldIsOwnerScoped(t *testing.T) {
	service, project, text := playFixture(t)
	world := createPlayWorld(t, service, project, text)
	if _, err := service.GetPlayWorld(context.Background(), "bob", project.ID, world.ID); !IsCode(err, ErrProjectNotFound) {
		t.Fatalf("err=%v", err)
	}
}

func TestDecodePlayJSONToleratesFenceAndNoise(t *testing.T) {
	var out playNarrationOutput
	raw := "说明文字\n```json\n{\"sceneText\":\"雨停了。\",\"suggestedActions\":[\"出门\"]}\n```\n结束"
	if err := decodePlayJSON(raw, &out); err != nil {
		t.Fatal(err)
	}
	if out.SceneText != "雨停了。" || len(out.SuggestedActions) != 1 {
		t.Fatalf("out=%+v", out)
	}
}
