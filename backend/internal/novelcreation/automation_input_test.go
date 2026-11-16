package novelcreation

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

func TestAutomationInputFourModesPreserved(t *testing.T) {
	for mode, field := range map[string]string{"long_imitation": "referenceText", "short_imitation": "referenceText", "fanfiction": "existingSetting", "continuation": "previousText"} {
		for _, kind := range []string{"schedule", "batch"} {
			t.Run(mode+"/"+kind, func(t *testing.T) {
				f := newServiceFixture(t, nil)
				ctx := context.Background()
				session, err := f.service.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: f.project.ID, Mode: mode})
				if err != nil {
					t.Fatal(err)
				}
				data, _ := json.Marshal(map[string]any{"sessionId": session.ID, "intervalMinutes": 60, "enabled": true, "totalChapters": 1, "prompt": "开始创作", "modelId": f.project.ModelID, "modeInputs": map[string]string{field: "参考片段"}})
				var run Run
				if kind == "schedule" {
					var in ScheduleInput
					json.Unmarshal(data, &in)
					plan, e := f.service.CreateSchedule(ctx, "alice", f.project.ID, in)
					if e != nil {
						t.Fatal(e)
					}
					run, err = f.service.RunSchedule(ctx, "alice", f.project.ID, plan.ID)
				} else {
					var in BatchInput
					json.Unmarshal(data, &in)
					_, e := f.service.CreateBatch(ctx, "alice", f.project.ID, in)
					if e != nil {
						t.Fatal(e)
					}
					dispatched, e := f.service.AdvanceBatches(ctx)
					err = e
					if len(dispatched) > 0 {
						run = dispatched[0].Run
					}
				}
				if err != nil {
					t.Fatal(err)
				}
				if run.ID == "" {
					t.Fatal("automation failed to create run with complete input")
				}
				var snap struct {
					ModeInputs map[string]string `json:"mode_inputs"`
				}
				json.Unmarshal(run.PromptSnapshot, &snap)
				if snap.ModeInputs[field] != "参考片段" {
					t.Fatal("automation dropped mode input")
				}
				if _, err = f.service.Run(ctx, "alice", run.ID); err != nil {
					t.Fatal(err)
				}
				if f.runner.lastCall().ModeInputs[field] != "参考片段" {
					t.Fatal("pipeline lost mode input")
				}

			})
		}
	}
}

func TestAutomationInputValidationAndOwnership(t *testing.T) {
	ctx := context.Background()
	for _, mode := range []string{"long_imitation", "short_imitation", "fanfiction", "continuation"} {
		t.Run(mode, func(t *testing.T) {
			f := newServiceFixture(t, nil)
			session, _ := f.service.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: f.project.ID, Mode: mode})
			if _, err := f.service.CreateSchedule(ctx, "alice", f.project.ID, ScheduleInput{SessionID: session.ID, Prompt: "创作", Enabled: true}); !IsCode(err, ErrInputRequired) {
				t.Fatalf("schedule accepted missing mode input: %v", err)
			}
			if _, err := f.service.CreateBatch(ctx, "alice", f.project.ID, BatchInput{SessionID: session.ID, TotalChapters: 1}); !IsCode(err, ErrInputRequired) {
				t.Fatalf("batch accepted missing mode input: %v", err)
			}
			plans, _ := f.service.ListSchedules(ctx, "alice", f.project.ID)
			batches, _ := f.service.ListBatches(ctx, "alice", f.project.ID)
			if len(plans)+len(batches) != 0 {
				t.Fatal("invalid automation was saved")
			}
		})
	}
	f := newServiceFixture(t, nil)
	foreign, _ := f.service.CreateProject(ctx, "bob", CreateProjectInput{Title: "其他账号", Genre: "都市", ModelID: "text-1"})
	material, _ := f.store.CreateMaterial(ctx, "bob", foreign.ID, CreateMaterialInput{Content: "其他资料"})
	note, _ := f.store.CreateNote(ctx, "bob", foreign.ID, ProjectNoteInput{Content: "其他备注"})
	style, _ := f.service.SaveStyleProfile(ctx, "bob", StyleProfile{Content: "其他文风"})
	for _, input := range []AutomationInput{{KnowledgeIDs: []string{material.ID}}, {NoteIDs: []string{note.ID}}, {StyleProfileID: style.ID}} {
		if _, err := f.service.prepareGenerationInput(ctx, "alice", f.project.ID, f.session.ID, input); !IsCode(err, ErrInputRequired) {
			t.Fatalf("foreign selection accepted: %+v %v", input, err)
		}
	}
	if _, err := f.service.SaveStyleProfile(ctx, "alice", StyleProfile{ID: style.ID, Content: "不能修改"}); !IsCode(err, ErrInputRequired) {
		t.Fatalf("foreign style update: %v", err)
	}
	if _, err := f.service.CreateSchedule(ctx, "bob", f.project.ID, ScheduleInput{SessionID: f.session.ID, Prompt: "创作"}); !IsCode(err, ErrProjectNotFound) {
		t.Fatalf("foreign project: %v", err)
	}
}

func TestAutomationInputSnapshotsSurviveEditsDeletionAndRetry(t *testing.T) {
	ctx := context.Background()
	f := newServiceFixture(t, nil)
	material, _ := f.store.CreateMaterial(ctx, "alice", f.project.ID, CreateMaterialInput{Content: "原始资料"})
	note, _ := f.store.CreateNote(ctx, "alice", f.project.ID, ProjectNoteInput{Content: "原始备注"})
	style, _ := f.service.SaveStyleProfile(ctx, "alice", StyleProfile{Content: "原始文风"})
	session, _ := f.service.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: f.project.ID, Mode: "short_imitation"})
	batch, err := f.service.CreateBatch(ctx, "alice", f.project.ID, BatchInput{SessionID: session.ID, TotalChapters: 2, KnowledgeIDs: []string{material.ID}, NoteIDs: []string{note.ID}, StyleProfileID: style.ID})
	if err != nil {
		t.Fatal(err)
	}
	dispatched, err := f.service.AdvanceBatches(ctx)
	if err != nil || len(dispatched) != 1 {
		t.Fatalf("dispatch=%v %v", dispatched, err)
	}
	run := dispatched[0].Run
	// Editing/deleting the originals cannot mutate an accepted run.
	f.service.SaveStyleProfile(ctx, "alice", StyleProfile{ID: style.ID, Content: "新文风"})
	f.store.UpdateNote(ctx, "alice", f.project.ID, note.ID, ProjectNoteInput{Content: "新备注"})
	f.store.DeleteMaterial(ctx, "alice", f.project.ID, material.ID)
	f.store.DeleteNote(ctx, "alice", f.project.ID, note.ID)
	if _, err = f.service.Run(ctx, "alice", run.ID); err != nil {
		t.Fatal(err)
	}
	got := f.runner.lastCall()
	if got.StyleProfile != "原始文风" || len(got.Notes) != 1 || got.Notes[0].Content != "原始备注" || len(got.KnowledgeMaterials) != 1 || got.KnowledgeMaterials[0].Content != "原始资料" {
		t.Fatalf("snapshot mutated: %+v", got)
	}
	// A retry copies the original content, even after selected sources disappear.
	f.store.UpdateRun(ctx, "alice", RunUpdate{ID: run.ID, Status: RunSucceededWithWarnings})
	retry, err := f.service.RetryRun(ctx, "alice", run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.service.Run(ctx, "alice", retry.ID); err != nil {
		t.Fatal(err)
	}
	got = f.runner.lastCall()
	if got.StyleProfile != "原始文风" || got.Notes[0].Content != "原始备注" || got.KnowledgeMaterials[0].Content != "原始资料" {
		t.Fatal("retry reloaded mutable materials")
	}
	// A new batch chapter must revalidate current selections, then stop once.
	f.service.AdvanceBatches(ctx)
	after, _ := f.store.GetBatch(ctx, "alice", batch.ID)
	if after.Status != BatchFailed || after.LastError == "" {
		t.Fatalf("invalid next chapter: %+v", after)
	}
	if _, err = f.service.ResumeBatch(ctx, "alice", f.project.ID, batch.ID); !IsCode(err, ErrInputRequired) {
		t.Fatalf("resume accepted deleted references: %v", err)
	}
	if f.runner.callCount() != 2 {
		t.Fatal("deleted references reached pipeline")
	}
}

func TestAutomationInputStyleOnlyAndDefaultSnapshot(t *testing.T) {
	for _, mode := range []string{"long_imitation", "short_imitation"} {
		t.Run(mode, func(t *testing.T) {
			ctx := context.Background()
			f := newServiceFixture(t, nil)
			style, _ := f.service.SaveStyleProfile(ctx, "alice", StyleProfile{Title: "文风", Content: "短句"})
			session, _ := f.service.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: f.project.ID, Mode: mode})
			plan, err := f.service.CreateSchedule(ctx, "alice", f.project.ID, ScheduleInput{SessionID: session.ID, Prompt: "创作", StyleProfileID: style.ID})
			if err != nil {
				t.Fatal(err)
			}
			run, err := f.service.RunSchedule(ctx, "alice", f.project.ID, plan.ID)
			if err != nil {
				t.Fatal(err)
			}
			f.service.SaveStyleProfile(ctx, "alice", StyleProfile{ID: style.ID, Content: "长句"})
			if _, err = f.service.Run(ctx, "alice", run.ID); err != nil {
				t.Fatal(err)
			}
			if f.runner.lastCall().StyleProfile != "短句" {
				t.Fatal("selected style was ignored")
			}
		})
	}
	ctx := context.Background()
	f := newServiceFixture(t, nil)
	f.store.UpdateProjectSettings(ctx, "alice", f.project.ID, ProjectSettings{StyleProfile: "旧作品文风"})
	run, err := f.service.SendMessage(ctx, "alice", f.sendInput("default-style"))
	if err != nil {
		t.Fatal(err)
	}
	f.store.UpdateProjectSettings(ctx, "alice", f.project.ID, ProjectSettings{StyleProfile: "更改文风"})
	if _, err = f.service.Run(ctx, "alice", run.ID); err != nil {
		t.Fatal(err)
	}
	if f.runner.lastCall().StyleProfile != "旧作品文风" {
		t.Fatal("default project style was not snapshotted")
	}
}

func TestAutomationInputLegacyScheduleStopsAndRepairValidates(t *testing.T) {
	ctx := context.Background()
	f := newServiceFixture(t, nil)
	session, _ := f.service.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: f.project.ID, Mode: "continuation"})
	plan, _ := f.store.CreateSchedule(ctx, "alice", Schedule{ProjectID: f.project.ID, SessionID: session.ID, Prompt: "续写", ModelID: "text-1", Enabled: true, IntervalMinutes: 60})
	if _, err := f.service.DispatchDueSchedules(ctx, plan.NextRunAt); err == nil {
		t.Fatal("legacy plan accepted missing input")
	}
	after, _ := f.store.GetSchedule(ctx, "alice", plan.ID)
	if after.Enabled || after.ConfigurationError == "" {
		t.Fatalf("plan was not paused: %+v", after)
	}
	dispatched, err := f.service.DispatchDueSchedules(ctx, after.NextRunAt)
	if err != nil || len(dispatched) != 0 {
		t.Fatal("broken plan repeatedly dispatched")
	}
	if _, err = f.service.UpdateSchedule(ctx, "alice", f.project.ID, plan.ID, ScheduleInput{SessionID: session.ID, Prompt: "续写", Enabled: true}); !IsCode(err, ErrInputRequired) {
		t.Fatal("invalid repair accepted")
	}
	fixed, err := f.service.UpdateSchedule(ctx, "alice", f.project.ID, plan.ID, ScheduleInput{SessionID: session.ID, Prompt: "续写", Enabled: true, ModeInputs: map[string]string{"previousText": "前文"}})
	if err != nil {
		t.Fatal(err)
	}
	if !fixed.Enabled || fixed.ConfigurationError != "" {
		t.Fatal("valid repair did not clear warning")
	}
}

func TestAutomationInputMemoryStoreDeepCopies(t *testing.T) {
	ctx := context.Background()
	f := newServiceFixture(t, nil)
	original := map[string]string{"referenceText": "原始"}
	ids := []string{"资料"}
	notes := []string{"备注"}
	plan, _ := f.store.CreateSchedule(ctx, "alice", Schedule{ProjectID: f.project.ID, SessionID: f.session.ID, ModeInputs: original, KnowledgeIDs: ids, NoteIDs: notes})
	batch, _ := f.store.CreateBatch(ctx, "alice", Batch{ProjectID: f.project.ID, SessionID: f.session.ID, ModeInputs: original, KnowledgeIDs: ids, NoteIDs: notes, Status: BatchRunning})
	original["referenceText"] = "外部修改"
	ids[0] = "外部修改"
	notes[0] = "外部修改"
	plan.ModeInputs["referenceText"] = "返回值修改"
	batch.KnowledgeIDs[0] = "返回值修改"
	plans, _ := f.store.ListSchedules(ctx, "alice", f.project.ID)
	batches, _ := f.store.ListBatches(ctx, "alice", f.project.ID)
	if plans[0].ModeInputs["referenceText"] != "原始" || batches[0].KnowledgeIDs[0] != "资料" || batches[0].NoteIDs[0] != "备注" {
		t.Fatal("stored contract aliases caller")
	}
	plans[0].ModeInputs["referenceText"] = "列表修改"
	batches[0].NoteIDs[0] = "列表修改"
	plan, _ = f.store.GetSchedule(ctx, "alice", plan.ID)
	batch, _ = f.store.GetBatch(ctx, "alice", batch.ID)
	if plan.ModeInputs["referenceText"] != "原始" || batch.NoteIDs[0] != "备注" {
		t.Fatal("read contract aliases store")
	}
}

func TestAutomationInputContinueAndBatchResumeKeepSelections(t *testing.T) {
	ctx := context.Background()
	f := newServiceFixture(t, nil)
	session, _ := f.service.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: f.project.ID, Mode: "long_imitation"})
	note, _ := f.store.CreateNote(ctx, "alice", f.project.ID, ProjectNoteInput{Content: "初始备注"})
	style, _ := f.service.SaveStyleProfile(ctx, "alice", StyleProfile{Content: "初始文风"})
	batch, err := f.service.CreateBatch(ctx, "alice", f.project.ID, BatchInput{SessionID: session.ID, TotalChapters: 2, ModeInputs: map[string]string{"referenceText": "参考片段"}, NoteIDs: []string{note.ID}, StyleProfileID: style.ID})
	if err != nil {
		t.Fatal(err)
	}
	paused, err := f.service.PauseBatch(ctx, "alice", f.project.ID, batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	resumed, err := f.service.ResumeBatch(ctx, "alice", f.project.ID, batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(batchInput(paused), batchInput(resumed)) {
		t.Fatal("pause/resume changed selections")
	}
	dispatched, err := f.service.AdvanceBatches(ctx)
	if err != nil || len(dispatched) != 1 {
		t.Fatalf("dispatch=%v %v", dispatched, err)
	}
	source := dispatched[0].Run
	// Simulate the durable blueprint checkpoint; confirmation must reuse the
	// original input instead of resolving edited materials again.
	if _, err = f.store.UpdateRun(ctx, "alice", RunUpdate{ID: source.ID, Status: RunAwaitingConfirmation, RuntimeSnapshot: json.RawMessage(`{"foundationPlan":{"title":"蓝图"}}`)}); err != nil {
		t.Fatal(err)
	}
	f.store.UpdateNote(ctx, "alice", f.project.ID, note.ID, ProjectNoteInput{Content: "后来备注"})
	f.service.SaveStyleProfile(ctx, "alice", StyleProfile{ID: style.ID, Content: "后来文风"})
	confirmed, err := f.service.ConfirmRun(ctx, "alice", source.ID, json.RawMessage(foundationJSON))
	if err != nil {
		t.Fatal(err)
	}
	var before, after map[string]json.RawMessage
	json.Unmarshal(source.PromptSnapshot, &before)
	json.Unmarshal(confirmed.PromptSnapshot, &after)
	for _, key := range []string{"mode_inputs", "knowledge_ids", "knowledge_materials", "note_ids", "notes", "style_profile_id", "style_profile"} {
		var left, right any
		json.Unmarshal(before[key], &left)
		json.Unmarshal(after[key], &right)
		if !reflect.DeepEqual(left, right) {
			t.Fatalf("confirmation changed %s", key)
		}
	}
}
