package novelcreation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"qiantie/backend/internal/batchfactoryv11"
	novelprompt "qiantie/backend/internal/prompt/novelcreation"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const studioMaxTurns = 20
const studioMaxRecordBytes = 2 * 1024 * 1024

func normalizeStudioThinking(value string) (string, error) {
	switch strings.TrimSpace(value) {
	case "", "standard", "medium":
		return "standard", nil
	case "fast", "low":
		return "fast", nil
	case "deep", "high":
		return "deep", nil
	default:
		return "", inputError("请选择快速、标准或深入的思考要求")
	}
}

func studioKey(key string) error {
	if strings.TrimSpace(key) == "" || utf8.RuneCountInString(key) > 128 {
		return inputError("请提供有效的请求编号")
	}
	return nil
}
func readStudio(ctx context.Context, st Store, owner, id string, lock bool) (r studioRecord, e error) {
	if owner == "" || id == "" {
		return r, studioMissing()
	}
	raw, e := st.LoadStudio(ctx, owner, id, lock)
	if e != nil {
		return r, e
	}
	e = json.Unmarshal(raw, &r)
	return r, persistenceError(e)
}
func writeStudio(ctx context.Context, st Store, owner string, r studioRecord, create bool) error {
	r.Session.UpdatedAt = time.Now().UTC()
	raw, e := json.Marshal(r)
	if e != nil {
		return persistenceError(e)
	}
	if len(raw) > studioMaxRecordBytes {
		return inputError("会话内容已达上限，请新建创作会话")
	}
	return st.SaveStudio(ctx, owner, r.Session.ID, raw, create)
}
func studioView(r studioRecord) StudioSnapshot {
	out := StudioSnapshot{Session: r.Session, Messages: append([]StudioMessage{}, r.Messages...), PendingAction: r.Action}
	if len(r.Turns) > 0 {
		v := r.Turns[len(r.Turns)-1].StudioTurn
		out.Turn = &v
	}
	return out
}

// The budget counts Unicode characters in message content, excluding protocol
// JSON and the fixed system instruction. Keep the newest user message whole,
// then add newest complete earlier user/assistant groups without cutting text.
func boundedStudioMessages(messages []StudioMessage, budget int) ([]StudioMessage, error) {
	if len(messages) == 0 {
		return []StudioMessage{}, nil
	}
	last := len(messages) - 1
	used := utf8.RuneCountInString(messages[last].Content)
	if used > budget {
		return nil, inputError("本条消息超过所选上下文字数，请提高上限或缩短消息")
	}
	start := last
	for start > 0 {
		previous := start - 1
		for previous > 0 && messages[previous].Role != "user" {
			previous--
		}
		cost := 0
		for i := previous; i < start; i++ {
			cost += utf8.RuneCountInString(messages[i].Content)
		}
		if used+cost > budget {
			break
		}
		used += cost
		start = previous
	}
	return append([]StudioMessage(nil), messages[start:]...), nil
}
func (s *Service) CreateStudioSession(ctx context.Context, owner string, in StudioCreateSessionInput) (out StudioSession, err error) {
	owner, err = requiredOwner(owner)
	if err != nil {
		return
	}
	in.IdempotencyKey = strings.TrimSpace(in.IdempotencyKey)
	in.Mode = strings.TrimSpace(in.Mode)
	in.Title = strings.TrimSpace(in.Title)
	if err = studioKey(in.IdempotencyKey); err != nil {
		return
	}
	if in.Mode != "long_original" {
		return out, inputError("此入口目前仅支持长篇原创会话")
	}
	if utf8.RuneCountInString(in.Title) > 255 {
		return out, inputError("会话名称不能超过255字")
	}
	if in.Title == "" {
		in.Title = "长篇原创讨论"
	}
	digest := fingerprint(in)
	err = s.store.AtomicCreate(ctx, owner, func(tx Store) error {
		now := time.Now().UTC()
		out = StudioSession{ID: novelID("studio"), Mode: in.Mode, Title: in.Title, CreatedAt: now, UpdatedAt: now}
		r := studioRecord{CreateKey: in.IdempotencyKey, CreateFingerprint: digest, Session: out, Messages: []StudioMessage{}, Turns: []studioTurnRecord{}}
		raw, _ := json.Marshal(r)
		saved, e := tx.ReserveStudioCreation(ctx, owner, in.IdempotencyKey, digest, raw)
		if e != nil {
			return e
		}
		if e = json.Unmarshal(saved, &r); e != nil {
			return persistenceError(e)
		}
		out = r.Session
		return nil
	})
	return
}
func (s *Service) ListStudioSessions(ctx context.Context, owner string) ([]StudioSession, error) {
	owner, e := requiredOwner(owner)
	if e != nil {
		return nil, e
	}
	return s.store.StudioSessions(ctx, owner)
}

// Only local interpretation is replayed after a saved reply. Never resend here.
func settleStudio(r *studioRecord) bool {
	if len(r.Turns) == 0 {
		return false
	}
	t := &r.Turns[len(r.Turns)-1]
	if t.State == "prepared" && time.Since(t.SentAt) > 2*time.Minute {
		t.State = "rejected"
		t.ErrorMessage = "上次请求在发送前中断，未调用模型。请重新发送。"
		return true
	}
	if t.State == "may_have_sent" && time.Since(t.SentAt) > 2*time.Minute {
		t.State = "unknown"
		t.ErrorMessage = unknownResult().Error()
		return true
	}
	if t.State != "received" {
		return false
	}
	var result struct {
		Message string `json:"message"`
		Action  *struct {
			Type string     `json:"type"`
			Book StudioBook `json:"book"`
		} `json:"action"`
	}
	d := json.NewDecoder(strings.NewReader(t.RawReply))
	d.DisallowUnknownFields()
	e := d.Decode(&result)
	var extra any
	if e == nil && d.Decode(&extra) != io.EOF {
		e = errors.New("extra JSON")
	}
	result.Message = strings.TrimSpace(result.Message)
	if e == nil && (result.Message == "" || utf8.RuneCountInString(result.Message) > 12000) {
		e = errors.New("invalid message")
	}
	if e == nil && result.Action != nil {
		b := &result.Action.Book
		b.Title = strings.TrimSpace(b.Title)
		b.Genre = strings.TrimSpace(b.Genre)
		if result.Action.Type != "create_book" || b.Title == "" || b.Genre == "" || utf8.RuneCountInString(b.Title) > 255 || utf8.RuneCountInString(b.Genre) > 96 || utf8.RuneCountInString(b.World) > 12000 || utf8.RuneCountInString(b.Protagonist) > 4000 || utf8.RuneCountInString(b.CoreConflict) > 4000 || b.TargetChapterWords < 800 || b.TargetChapterWords > 8000 {
			e = errors.New("invalid action")
		}
	}
	if e != nil {
		t.State = "invalid_output"
		t.ErrorMessage = "模型回复格式不正确，已保留原始结果，未创建作品。"
		return true
	}
	r.Messages = append(r.Messages, StudioMessage{ID: novelID("studio-msg"), Role: "assistant", Content: result.Message, Status: "completed"})
	if result.Action != nil {
		r.Action = &StudioAction{ID: novelID("action"), Type: "create_book", Status: "pending", Book: result.Action.Book}
	} else if r.Action != nil {
		r.Action.Status = "superseded"
	}
	t.State = "completed"
	t.ErrorMessage = ""
	return true
}
func (s *Service) GetStudioSnapshot(ctx context.Context, owner, id string) (out StudioSnapshot, err error) {
	owner, err = requiredOwner(owner)
	if err != nil {
		return
	}
	err = s.store.AtomicCreate(ctx, owner, func(tx Store) error {
		r, e := readStudio(ctx, tx, owner, strings.TrimSpace(id), true)
		if e != nil {
			return e
		}
		if settleStudio(&r) {
			if e = writeStudio(ctx, tx, owner, r, false); e != nil {
				return e
			}
		}
		out = studioView(r)
		return nil
	})
	return
}

func (s *Service) studioResolver() (studioProviderResolver, error) {
	p, ok := s.pipeline.(studioProviderResolver)
	if !ok {
		return nil, inputError("小说聊天模型尚未配置")
	}
	return p, nil
}
func (s *Service) SendStudioMessage(ctx context.Context, owner, id string, in StudioSendInput) (out StudioSnapshot, err error) {
	owner, err = requiredOwner(owner)
	if err != nil {
		return
	}
	id = strings.TrimSpace(id)
	in.Content = strings.TrimSpace(in.Content)
	in.ModelID = strings.TrimSpace(in.ModelID)
	in.IdempotencyKey = strings.TrimSpace(in.IdempotencyKey)
	in.ThinkingLevel, err = normalizeStudioThinking(in.ThinkingLevel)
	if err != nil {
		return
	}
	if in.ContextBudget == 0 {
		in.ContextBudget = 64000
	}
	if err = studioKey(in.IdempotencyKey); err != nil {
		return
	}
	if in.Content == "" || utf8.RuneCountInString(in.Content) > 12000 || in.ModelID == "" || len(in.ModelID) > 96 || len(in.ThinkingLevel) > 32 || in.ContextBudget < 0 || in.ContextBudget > 64000 {
		return out, inputError("请检查消息、模型和上下文长度")
	}
	if utf8.RuneCountInString(in.Content) > in.ContextBudget {
		return out, inputError("本条消息超过所选上下文字数，请提高上限或缩短消息")
	}
	digest := fingerprint(in)
	var request batchfactoryv11.TextCompletionRequest
	var turnID string
	fresh := false
	err = s.store.AtomicCreate(ctx, owner, func(tx Store) error {
		// An account-wide reservation prevents a lost reply from being resent
		// under the same key to another studio. It intentionally has no project
		// or completed-result receipt: those do not exist before confirmation.
		binding := fingerprint(struct {
			SessionID string
			Input     StudioSendInput
		}{id, in})
		receipt, e := tx.ReserveSubmission(ctx, owner, "studio_turn", in.IdempotencyKey, binding)
		if e != nil {
			return e
		}
		if receipt.Fingerprint != binding {
			return conflict()
		}
		r, e := readStudio(ctx, tx, owner, id, true)
		if e != nil {
			return e
		}
		if settleStudio(&r) {
			if e = writeStudio(ctx, tx, owner, r, false); e != nil {
				return e
			}
		}
		for _, t := range r.Turns {
			if t.IdempotencyKey == in.IdempotencyKey {
				if t.Fingerprint != digest {
					return conflict()
				}
				out = studioView(r)
				return nil
			}
		}
		if r.Session.ProjectID != "" {
			return inputError("此会话已创建作品，请进入作品会话继续")
		}
		if len(r.Turns) >= studioMaxTurns {
			return inputError("会话轮数已达上限，请新建会话")
		}
		if len(r.Turns) > 0 {
			last := r.Turns[len(r.Turns)-1]
			if last.State == "may_have_sent" || last.State == "prepared" || last.State == "unknown" {
				return unknownResult()
			}
		}
		// A pending proposal is superseded as soon as a new user instruction is accepted.
		var priorProposal *StudioAction
		remainingBudget := in.ContextBudget
		if r.Action != nil {
			copy := *r.Action
			priorProposal = &copy
			b := copy.Book
			// Proposed field content counts toward the same Unicode-character
			// budget as message text; JSON keys/protocol syntax do not count.
			proposalChars := utf8.RuneCountInString(b.Title + b.Genre + b.World + b.Protagonist + b.CoreConflict + strconv.Itoa(b.TargetChapterWords))
			remainingBudget -= proposalChars
			if remainingBudget < utf8.RuneCountInString(in.Content) {
				return inputError("已有建议的作品设定加本条消息超过所选上下文字数，请提高上限后再修改")
			}
		}
		if r.Action != nil && r.Action.Status == "pending" {
			r.Action.Status = "superseded"
		}
		r.Messages = append(r.Messages, StudioMessage{ID: novelID("studio-msg"), Role: "user", Content: in.Content, Status: "accepted"})
		contextMessages, e := boundedStudioMessages(r.Messages, remainingBudget)
		if e != nil {
			return e
		}
		payload, _ := json.Marshal(map[string]any{"messages": contextMessages, "previousProposal": priorProposal, "mode": r.Session.Mode, "thinkingLevel": in.ThinkingLevel, "contextBudget": in.ContextBudget})
		systemPrompts, lookupErr := s.snapshotSystemPrompts()
		if lookupErr != nil {
			return lookupErr
		}
		planningSystem, lookupErr := frozenSystemText(RunContext{SystemPrompts: systemPrompts}, novelprompt.StudioPlanningPresetID, novelprompt.StudioPlanningSystem)
		if lookupErr != nil {
			return lookupErr
		}
		request = batchfactoryv11.TextCompletionRequest{SystemPrompt: planningSystem, UserPrompt: string(payload), MaxTokens: 6000}
		turnID = novelID("studio-turn")
		r.Turns = append(r.Turns, studioTurnRecord{StudioTurn: StudioTurn{ID: turnID, IdempotencyKey: in.IdempotencyKey, State: "prepared", ModelID: in.ModelID}, Fingerprint: digest, Request: request, SentAt: time.Now().UTC()})
		// Reserve space for the raw reply and its interpreted message/action.
		preparedRaw, _ := json.Marshal(r)
		if len(preparedRaw) > studioMaxRecordBytes-768*1024 {
			return inputError("会话内容已达上限，请新建创作会话")
		}
		if e = writeStudio(ctx, tx, owner, r, false); e != nil {
			return e
		}
		out = studioView(r)
		fresh = true
		return nil
	})
	if err != nil || !fresh {
		return
	}
	resolver, e := s.studioResolver()
	var provider batchfactoryv11.DirectorProvider
	if e == nil {
		provider, e = resolver.ResolveStudioProvider(ctx, owner, in.ModelID)
	}
	if e != nil || provider == nil {
		if e == nil {
			e = inputError("模型尚未配置")
		}
		return s.finishStudioTurn(owner, id, turnID, "rejected", "", e.Error())
	}
	// Persist the send marker in a separate transaction before any provider I/O.
	err = s.store.AtomicCreate(ctx, owner, func(tx Store) error {
		r, e := readStudio(ctx, tx, owner, id, true)
		if e != nil {
			return e
		}
		t := &r.Turns[len(r.Turns)-1]
		if t.ID != turnID || t.State != "prepared" {
			return conflict()
		}
		t.State = "may_have_sent"
		t.SentAt = time.Now().UTC()
		return writeStudio(ctx, tx, owner, r, false)
	})
	if err != nil {
		return
	}
	callCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	reply, callErr := provider.Complete(batchfactoryv11.WithSingleAttempt(callCtx), request)
	if callErr != nil {
		state := "unknown"
		message := unknownResult().Error()
		var httpErr *batchfactoryv11.DirectorHTTPError
		if errors.As(callErr, &httpErr) && (httpErr.StatusCode == 400 || httpErr.StatusCode == 401 || httpErr.StatusCode == 403 || httpErr.StatusCode == 404 || httpErr.StatusCode == 422 || httpErr.StatusCode == 429) {
			state = "rejected"
			message = mapModelError(callErr).Error()
		}
		return s.finishStudioTurn(owner, id, turnID, state, "", message)
	}
	// Store raw bytes durably first; parsing is a recoverable local operation.
	replyJSON, _ := json.Marshal(reply)
	if len(replyJSON) > 256*1024 {
		preview := reply
		if len(preview) > 64*1024 {
			preview = preview[:64*1024]
			for !utf8.ValidString(preview) {
				preview = preview[:len(preview)-1]
			}
		}
		sum := sha256.Sum256([]byte(reply))
		return s.finishStudioTurn(owner, id, turnID, "invalid_output", preview, "模型回复超出保存上限，已保存前64KB预览及原文校验信息，未创建作品，不会自动重试。", studioReplyOverflow{SHA256: hex.EncodeToString(sum[:]), Bytes: len(reply)})
	}
	out, err = s.finishStudioTurn(owner, id, turnID, "received", reply, "")
	if err != nil {
		return
	}
	return s.GetStudioSnapshot(context.Background(), owner, id)
}

type studioReplyOverflow struct {
	SHA256 string
	Bytes  int
}

func (s *Service) finishStudioTurn(owner, id, turnID, state, reply, message string, overflow ...studioReplyOverflow) (out StudioSnapshot, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = s.store.AtomicCreate(ctx, owner, func(tx Store) error {
		r, e := readStudio(ctx, tx, owner, id, true)
		if e != nil {
			return e
		}
		if len(r.Turns) == 0 {
			return conflict()
		}
		t := &r.Turns[len(r.Turns)-1]
		if t.ID != turnID {
			return conflict()
		}
		if t.State != "prepared" && t.State != "may_have_sent" && t.State != "unknown" {
			out = studioView(r)
			return nil
		}
		t.State = state
		t.RawReply = reply
		if len(overflow) > 0 {
			t.RawReplySHA256 = overflow[0].SHA256
			t.RawReplyBytes = overflow[0].Bytes
			t.RawReplyTruncated = true
		}
		t.ErrorMessage = message
		if e = writeStudio(ctx, tx, owner, r, false); e != nil {
			return e
		}
		out = studioView(r)
		return nil
	})
	return
}

func (s *Service) CancelStudioAction(ctx context.Context, owner, id string, in StudioActionInput) (out StudioSnapshot, err error) {
	owner, err = requiredOwner(owner)
	if err != nil {
		return
	}
	err = s.store.AtomicCreate(ctx, owner, func(tx Store) error {
		r, e := readStudio(ctx, tx, owner, strings.TrimSpace(id), true)
		if e != nil {
			return e
		}
		if r.Action == nil || r.Action.ID != strings.TrimSpace(in.ActionID) {
			return conflict()
		}
		if r.Action.Status == "cancelled" {
			out = studioView(r)
			return nil
		}
		if r.Action.Status != "pending" {
			return conflict()
		}
		r.Action.Status = "cancelled"
		if e = writeStudio(ctx, tx, owner, r, false); e != nil {
			return e
		}
		out = studioView(r)
		return nil
	})
	return
}

func (s *Service) ConfirmStudioAction(ctx context.Context, owner, id string, in StudioActionInput) (out StudioTransition, err error) {
	owner, err = requiredOwner(owner)
	if err != nil {
		return
	}
	id = strings.TrimSpace(id)
	in.ActionID = strings.TrimSpace(in.ActionID)
	in.ModelID = strings.TrimSpace(in.ModelID)
	// Completed confirmation is an immutable receipt, even if model config changed.
	r, e := readStudio(ctx, s.store, owner, id, false)
	if e != nil {
		return out, e
	}
	if r.Action == nil || r.Action.ID != in.ActionID {
		return out, conflict()
	}
	if r.Action.Status == "confirmed" && r.Transition != nil {
		return *r.Transition, nil
	}
	if r.Action.Status != "pending" {
		return out, conflict()
	}
	if strings.TrimSpace(in.ThinkingLevel) == "" && len(r.Turns) > 0 {
		var lastInput struct {
			ThinkingLevel string `json:"thinkingLevel"`
		}
		if e = json.Unmarshal([]byte(r.Turns[len(r.Turns)-1].Request.UserPrompt), &lastInput); e != nil {
			return out, persistenceError(e)
		}
		in.ThinkingLevel = lastInput.ThinkingLevel
	}
	in.ThinkingLevel, e = normalizeStudioThinking(in.ThinkingLevel)
	if e != nil {
		return out, e
	}
	if in.ModelID == "" || len(in.ModelID) > 96 {
		return out, inputError("请选择已配置的文本模型")
	}
	resolver, e := s.studioResolver()
	if e != nil {
		return out, e
	}
	provider, e := resolver.ResolveStudioProvider(ctx, owner, in.ModelID)
	if e != nil {
		return out, e
	}
	if provider == nil {
		return out, inputError("模型尚未配置")
	}
	err = s.store.AtomicCreate(ctx, owner, func(tx Store) error {
		r, e := readStudio(ctx, tx, owner, id, true)
		if e != nil {
			return e
		}
		if r.Action == nil || r.Action.ID != in.ActionID {
			return conflict()
		}
		if r.Action.Status == "confirmed" && r.Transition != nil {
			out = *r.Transition
			return nil
		}
		if r.Action.Status != "pending" {
			return conflict()
		}
		b := r.Action.Book
		project, e := tx.CreateProject(ctx, owner, CreateProjectInput{Title: b.Title, Genre: b.Genre, World: b.World, Protagonist: b.Protagonist, CoreConflict: b.CoreConflict, TargetChapterWords: b.TargetChapterWords, ModelID: in.ModelID})
		if e != nil {
			return e
		}
		session, e := tx.CreateSession(ctx, owner, CreateSessionInput{ProjectID: project.ID, Mode: "long_original", Title: b.Title})
		if e != nil {
			return e
		}
		child := NewService(tx, quietRuntime{s.runtime}, s.pipeline)
		child.promptCatalog = s.promptCatalog
		// Preserve the actual bounded planning input, not the complete durable
		// transcript. Reinjecting excluded history here would defeat the chosen
		// context limit. The studio ID traces the full original conversation.
		if len(r.Turns) == 0 {
			return inputError("缺少已保存的讨论输入，无法创建生产任务")
		}
		boundedDiscussion := json.RawMessage(r.Turns[len(r.Turns)-1].Request.UserPrompt)
		planningContext, e := json.Marshal(map[string]any{"studioSessionId": r.Session.ID, "confirmedBook": b, "discussion": boundedDiscussion})
		if e != nil {
			return persistenceError(e)
		}
		run, e := child.SendMessage(ctx, owner, SendMessageInput{ProjectID: project.ID, SessionID: session.ID, Content: novelprompt.StudioProductionInstruction + "\n" + string(planningContext), ModelID: in.ModelID, ThinkingLevel: in.ThinkingLevel, IdempotencyKey: "studio-confirm:" + r.Action.ID})
		if e != nil {
			return e
		}
		r.Session.ProjectID = project.ID
		r.Session.ProductionSessionID = session.ID
		r.Session.ProductionRunID = run.ID
		r.Session.UpdatedAt = time.Now().UTC()
		r.Action.Status = "confirmed"
		out = StudioTransition{Session: r.Session, Project: project, ProductionSession: session, Run: run}
		r.Transition = &out
		return writeStudio(ctx, tx, owner, r, false)
	})
	if err == nil {
		s.Notify(ctx, owner, out.Run.ID)
	}
	return
}
