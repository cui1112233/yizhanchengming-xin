package novelcreation

import (
	"context"
	"encoding/json"
	"qiantie/backend/internal/batchfactoryv11"
	prompt "qiantie/backend/internal/prompt/novelcreation"
	"strings"
	"testing"
)

func writerConversationFixture(t *testing.T, replies []string) (*Pipeline, *queuedPipelineProvider, RunContext, ShortStory) {
	t.Helper()
	a := NormalizeAgentSettings(AgentSettings{Version: 1, PassScore: 80, MaxRevisionRounds: 2}, 2000)
	p, provider, in, story := controlledShortFixture(t, replies, a)
	_ = json.Unmarshal([]byte(`{"ShortWriterConversationVersion":1}`), &in)
	return p, provider, in, story
}

func TestShortWriterConversationUsesFrozenFollowupAndExistingModelRoutes(t *testing.T) {
	body := shortControlText(strings.Repeat("首稿。", 300))
	revised := shortControlText(strings.Repeat("修订。", 300))
	p, provider, in, story := writerConversationFixture(t, []string{body, `{"score":10,"issues":[]}`, revised, `{"score":95,"issues":[]}`, shortControlText("推荐")})
	in.ShortImitationVersion = 2
	in.Session.Mode = "short_imitation"
	in.SystemPrompts = &SystemPromptSnapshot{Version: 1, Entries: map[string]FrozenSystemPrompt{}}
	for _, definition := range prompt.AdminPrompts() {
		in.SystemPrompts.Entries[definition.ID] = FrozenSystemPrompt{ID: definition.ID, Version: 1, Revision: 7, Body: "冻结-" + definition.ID + "\n" + definition.Definition.SystemPrompt}
	}
	in.Agent.ModelRoutes = map[string]string{"chapter-draft": "own-writer", "chapter-revision": "own-reviser", "chapter-audit": "own-reviewer"}
	routes := []string{}
	p.resolveProvider = func(_ context.Context, _ string, model string) (batchfactoryv11.DirectorProvider, error) {
		routes = append(routes, model)
		return provider, nil
	}
	if _, err := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, true); err != nil {
		t.Fatal(err)
	}
	revision := provider.requests[2]
	if !strings.Contains(revision.SystemPrompt, "冻结-"+prompt.ShortBodyPrompt.ID) || !strings.Contains(revision.SystemPrompt, "冻结-"+prompt.ShortImitationWritingID) || strings.Contains(revision.SystemPrompt, "冻结-"+prompt.ShortBodyRevision.ID) {
		t.Fatal("writer system replaced with revision system")
	}
	var user map[string]json.RawMessage
	if err := json.Unmarshal([]byte(revision.UserPrompt), &user); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(user["revision_instruction"]), "冻结-"+prompt.ShortWriterRevisionFollowupID) || !strings.Contains(string(user["configured_revision_instruction"]), "冻结-"+prompt.ShortBodyRevision.ID) || len(user["context"]) == 0 {
		t.Fatal("frozen follow-up/configured requirements lost")
	}
	if routes[0] != "own-writer" || routes[1] != "own-reviewer" || routes[2] != "own-reviser" {
		t.Fatal("model routing changed", routes)
	}
}

func TestShortWriterConversationRetryRebuildsHistoryWithoutRedrafting(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	base := imitationStageReplies()
	replies := append([]string{}, base[:6]...)
	replies = append(replies, `{"score":10,"issues":["需修订"]}`, "", shortControlText(strings.Repeat("修订。", 300)), `{"score":95,"issues":[]}`, base[7], base[8])
	provider := &queuedPipelineProvider{responses: replies, errors: make([]error, 8)}
	provider.errors[7] = &batchfactoryv11.DirectorHTTPError{StatusCode: 429, Body: "rate limited"}
	service := NewService(store, NewMemoryRuntime(), pipelineWithProvider(provider))
	input := shortInput("writer-history-retry")
	input.TargetWordCount = 1000
	started, err := service.StartShortStory(ctx, "alice", input)
	if err != nil {
		t.Fatal(err)
	}
	freezeLegacyShortFullDraft(t, service, "alice", started.Run.ID)
	var snapshot map[string]any
	_ = json.Unmarshal(started.Run.PromptSnapshot, &snapshot)
	if snapshot["short_writer_conversation_version"] != float64(1) {
		t.Fatal("new imitation did not freeze conversation protocol")
	}
	first, err := service.Run(ctx, "alice", started.Run.ID)
	if !IsCode(err, ErrModelRateLimited) || first.Status != RunFailed {
		t.Fatal(first.Status, err)
	}
	next, err := service.RetryRun(ctx, "alice", first.ID)
	if err != nil {
		t.Fatal(err)
	}
	service = NewService(store, NewMemoryRuntime(), pipelineWithProvider(provider))
	last, err := service.Run(ctx, "alice", next.ID)
	if err != nil || last.Status != RunSucceeded || len(provider.requests) != 12 {
		t.Fatal("retry repeated completed writer calls", last.Status, err, len(provider.requests))
	}
	if len(provider.requests[8].Messages) != 4 || provider.requests[8].Messages[2].Content != base[5] {
		t.Fatal("retry lost original draft")
	}
	execution, ok, err := store.GetExecution(ctx, "alice", next.ID)
	if err != nil || !ok {
		t.Fatal(err)
	}
	draft := execution.Calls[prompt.ShortBodyPrompt.ID+":0:0"]
	revision := execution.Calls[prompt.ShortBodyRevision.ID+":0:0:controlled:1"]
	if draft.Metadata == nil || len(draft.Metadata.Messages) != 2 || revision.Metadata == nil || len(revision.Metadata.Messages) != 4 || revision.Metadata.Messages[2].Content != base[5] {
		t.Fatal("journal metadata does not preserve actual conversation")
	}
}

func TestShortWriterConversationKeepsOriginalDraftAcrossRevisionRounds(t *testing.T) {
	first := shortControlText(strings.Repeat("首稿。", 300))
	second := shortControlText(strings.Repeat("第二稿。", 250))
	third := shortControlText(strings.Repeat("第三稿。", 250))
	p, provider, in, story := writerConversationFixture(t, []string{first, `{"score":10,"issues":["首轮意见"]}`, second, `{"score":20,"issues":["第二轮意见"]}`, third, `{"score":95,"issues":[]}`, shortControlText("推荐")})
	out, err := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, true)
	if err != nil || out.Status != "completed" {
		t.Fatal(out.Status, err)
	}
	original := provider.requests[0]
	if len(original.CompletionMessages()) != 2 {
		t.Fatal("draft should remain two messages")
	}
	for _, index := range []int{2, 4} {
		request := provider.requests[index]
		messages := request.CompletionMessages()
		if len(messages) != 4 {
			t.Fatalf("revision %d dropped original writer conversation: got %d messages", index, len(messages))
		}
		if messages[0].Role != "system" || messages[0].Content != original.SystemPrompt || messages[1].Role != "user" || messages[1].Content != original.UserPrompt || messages[2].Role != "assistant" || messages[2].Content != first || messages[3].Role != "user" {
			t.Fatal("original draft context changed", index)
		}
		if request.SystemPrompt != messages[0].Content || request.UserPrompt != messages[3].Content {
			t.Fatal("request summary does not match actual conversation")
		}
	}
	if !strings.Contains(provider.requests[4].UserPrompt, "第二轮意见") || !strings.Contains(provider.requests[4].UserPrompt, "第二稿") {
		t.Fatal("latest feedback/body missing from follow-up")
	}
}

func TestShortWriterConversationRepairRetainsHistoryAndEffectiveDraft(t *testing.T) {
	first := shortControlText(strings.Repeat("首稿。", 300))
	revised := shortControlText(strings.Repeat("修订。", 300))
	for _, draftRepair := range []bool{false, true} {
		t.Run(map[bool]string{false: "revision_repair", true: "draft_repair"}[draftRepair], func(t *testing.T) {
			replies := []string{first, `{"score":10,"issues":[]}`, `{}`, revised, `{"score":95,"issues":[]}`, shortControlText("推荐")}
			revisionIndex, repairIndex := 2, 3
			if draftRepair {
				replies = []string{`{}`, first, `{"score":10,"issues":[]}`, revised, `{"score":95,"issues":[]}`, shortControlText("推荐")}
				revisionIndex = 3
				repairIndex = 1
			}
			p, provider, in, story := writerConversationFixture(t, replies)
			if _, err := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, true); err != nil {
				t.Fatal(err)
			}
			messages := provider.requests[revisionIndex].CompletionMessages()
			if len(messages) != 4 || messages[2].Content != first {
				t.Fatal("repair did not preserve effective original draft", len(messages))
			}
			repair := provider.requests[repairIndex].CompletionMessages()
			want := 6
			if draftRepair {
				want = 4
			}
			if len(repair) != want || repair[0].Content != provider.requests[0].SystemPrompt || repair[len(repair)-2].Role != "assistant" || repair[len(repair)-2].Content != `{}` || repair[len(repair)-1].Role != "user" {
				t.Fatal("format repair discarded conversation", len(repair), want)
			}
		})
	}
}

func TestShortWriterConversationVersionZeroKeepsLegacyRequests(t *testing.T) {
	body := shortControlText(strings.Repeat("正文。", 300))
	p, provider, in, story := writerConversationFixture(t, []string{body, `{"score":10,"issues":[]}`, `{}`, body, `{"score":95,"issues":[]}`, shortControlText("推荐")})
	_ = json.Unmarshal([]byte(`{"ShortWriterConversationVersion":0}`), &in)
	if _, err := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, true); err != nil {
		t.Fatal(err)
	}
	for _, request := range provider.requests {
		if len(request.Messages) != 0 || len(request.CompletionMessages()) != 2 {
			t.Fatal("old protocol fingerprint changed")
		}
	}
}

func TestShortWriterConversationSeparatesPiecesAndHonorsCancellation(t *testing.T) {
	first := shortControlText(strings.Repeat("第一段。", 250))
	second := shortControlText(strings.Repeat("第二段。", 250))
	bad := `{"score":10,"issues":[]}`
	pass := `{"score":95,"issues":[]}`
	p, provider, in, story := writerConversationFixture(t, []string{first, bad, first, pass, shortControlText("记忆"), second, bad, second, pass, shortControlText("推荐")})
	story.Input.TargetWordCount = 2000
	story.Input.TargetChapters = 2
	story.Outline.Chapters = append(story.Outline.Chapters, ShortOutlineChapter{Number: 2, Title: "终点", Summary: "结清冲突"})
	if _, err := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, true); err != nil {
		t.Fatal(err)
	}
	if provider.requests[2].Messages[2].Content != first || provider.requests[7].Messages[2].Content != second || provider.requests[2].Messages[1].Content == provider.requests[7].Messages[1].Content {
		t.Fatal("pieces shared original draft history")
	}
	runtime := NewMemoryRuntime()
	p, provider, in, story = writerConversationFixture(t, []string{first, bad, first, pass, shortControlText("推荐")})
	in.Runtime = runtime
	provider.afterCall = func(n int) {
		if n == 1 {
			_ = runtime.Cancel(context.Background(), in.Run.ID)
		}
	}
	if _, err := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, true); err == nil || len(provider.requests) != 2 {
		t.Fatal("cancelled task sent revision", err, len(provider.requests))
	}
}

func TestShortWriterConversationMissingDraftStopsBeforeDispatch(t *testing.T) {
	p, provider, in, _ := writerConversationFixture(t, nil)
	var revised shortText
	if err := p.shortCall(context.Background(), in, &recordingEventSink{}, prompt.ShortBodyRevision, ":missing", RunWriting, map[string]any{"review": "需要修改"}, &revised, func() error { return validShortText(revised, 12000) }, 4000); err == nil || len(provider.requests) != 0 {
		t.Fatal("missing history silently downgraded to two messages", err)
	}
}

func TestShortWriterConversationSnapshotIsLimitedAndRejectsUnknownVersion(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	provider := &queuedPipelineProvider{}
	service := NewService(store, NewMemoryRuntime(), pipelineWithProvider(provider))
	input := shortInput("original-without-writer-conversation")
	input.Mode = "short_original"
	input.Requirements = "寻找线索"
	original, err := service.StartShortStory(ctx, "alice", input)
	if err != nil {
		t.Fatal(err)
	}
	var snapshot map[string]any
	_ = json.Unmarshal(original.Run.PromptSnapshot, &snapshot)
	if _, exists := snapshot["short_writer_conversation_version"]; exists {
		t.Fatal("conversation protocol unexpectedly added to original mode")
	}
	input = shortInput("invalid-writer-conversation")
	started, err := service.StartShortStory(ctx, "alice", input)
	if err != nil {
		t.Fatal(err)
	}
	_ = json.Unmarshal(started.Run.PromptSnapshot, &snapshot)
	snapshot["short_writer_conversation_version"] = 2
	raw, _ := json.Marshal(snapshot)
	run := store.runs[memoryKey("alice", started.Run.ID)]
	run.PromptSnapshot = raw
	store.runs[memoryKey("alice", started.Run.ID)] = run
	failed, err := service.Run(ctx, "alice", started.Run.ID)
	if err == nil || failed.Status != RunFailed || len(provider.requests) != 0 {
		t.Fatal("unknown protocol dispatched calls", failed.Status, err)
	}
}

func TestShortWriterConversationOwnedMySQLRetry(t *testing.T) {
	db := studioOwnedDB(t)
	ctx := context.Background()
	for _, version := range []int{0, 1} {
		t.Run(map[int]string{0: "legacy", 1: "conversation"}[version], func(t *testing.T) {
			owner := novelID("writer-sql")
			base := imitationStageReplies()
			replies := append([]string{}, base[:6]...)
			replies = append(replies, `{"score":10,"issues":["需修订"]}`, "", shortControlText(strings.Repeat("修订。", 300)), `{"score":95,"issues":[]}`, base[7], base[8])
			provider := &queuedPipelineProvider{responses: replies, errors: make([]error, 8)}
			provider.errors[7] = &batchfactoryv11.DirectorHTTPError{StatusCode: 429, Body: "rate limited"}
			makeService := func() *Service {
				return NewService(NewMySQLStore(db), NewMemoryRuntime(), pipelineWithProvider(provider))
			}
			service := makeService()
			settings := NormalizeAgentSettings(AgentSettings{Version: 1, PassScore: 80, MaxRevisionRounds: 1}, 2000)
			pause := false
			settings.PauseAfterBlueprint = &pause
			if _, err := service.UpdateAgentDefaults(ctx, owner, settings); err != nil {
				t.Fatal(err)
			}
			input := shortInput(novelID("writer-retry"))
			input.TargetWordCount = 1000
			started, err := service.StartShortStory(ctx, owner, input)
			if err != nil {
				t.Fatal(err)
			}
			freezeLegacyShortFullDraft(t, service, owner, started.Run.ID)
			if version == 0 {
				if _, err = db.Exec(`UPDATE novel_runs SET prompt_snapshot_json=JSON_SET(prompt_snapshot_json,'$.short_writer_conversation_version',0) WHERE owner_username=? AND run_id=?`, owner, started.Run.ID); err != nil {
					t.Fatal(err)
				}
			}
			first, err := service.Run(ctx, owner, started.Run.ID)
			if !IsCode(err, ErrModelRateLimited) || first.Status != RunFailed || len(provider.requests) != 8 {
				t.Fatal(first.Status, err, len(provider.requests))
			}
			service = makeService()
			next, err := service.RetryRun(ctx, owner, first.ID)
			if err != nil {
				t.Fatal(err)
			}
			service = makeService()
			last, err := service.Run(ctx, owner, next.ID)
			if err != nil || last.Status != RunSucceeded || len(provider.requests) != 12 {
				t.Fatal("SQL retry repeated original draft", last.Status, err, len(provider.requests))
			}
			request := provider.requests[8]
			if version == 1 {
				if len(request.Messages) != 4 || request.Messages[2].Content != base[5] || request.Messages[0].Content != provider.requests[5].SystemPrompt || request.Messages[1].Content != provider.requests[5].UserPrompt {
					t.Fatal("SQL replay lost original writer history")
				}
			} else if len(request.Messages) != 0 || len(request.CompletionMessages()) != 2 {
				t.Fatal("legacy SQL request changed")
			}
			execution, ok, err := NewMySQLStore(db).GetExecution(ctx, owner, next.ID)
			if err != nil || !ok {
				t.Fatal(err)
			}
			draft := execution.Calls[prompt.ShortBodyPrompt.ID+":0:0"]
			revision := execution.Calls[prompt.ShortBodyRevision.ID+":0:0:controlled:1"]
			if draft.Metadata == nil || revision.Metadata == nil {
				t.Fatal("SQL metadata missing")
			}
			if version == 1 {
				if len(draft.Metadata.Messages) != 2 || len(revision.Metadata.Messages) != 4 || revision.Metadata.Messages[2].Content != base[5] {
					t.Fatal("SQL metadata conversation missing after reopening")
				}
			} else if len(draft.Metadata.Messages) != 0 || len(revision.Metadata.Messages) != 0 {
				t.Fatal("legacy metadata changed")
			}
			t.Logf("version=%d: writer draft sent once; revision429 resumed through a reopened service; stored messages=%d; 12 fake calls, zero paid requests", version, len(revision.Metadata.Messages))
		})
	}
}
