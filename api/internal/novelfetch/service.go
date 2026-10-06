package novelfetch

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

type Service struct {
	store      Store
	fetcher    Fetcher
	model      TextModel
	dispatcher Dispatcher
	batch      BatchFactoryBoundary
	publisher  PublishIntentBoundary
	now        func() time.Time
}

type Options struct {
	Store        Store
	Fetcher      Fetcher
	Model        TextModel
	Dispatcher   Dispatcher
	BatchFactory BatchFactoryBoundary
	Publisher    PublishIntentBoundary
	Now          func() time.Time
}

func NewService(options Options) *Service {
	now := options.Now
	if now == nil {
		now = time.Now
	}
	return &Service{
		store: options.Store, fetcher: options.Fetcher, model: options.Model,
		dispatcher: options.Dispatcher, batch: options.BatchFactory,
		publisher: options.Publisher, now: now,
	}
}

func (s *Service) CreateBatch(ctx context.Context, input CreateBatchInput) (Batch, []Book, error) {
	if s == nil || s.store == nil {
		return Batch{}, nil, ErrRuntimeUnavailable
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = "小说获取批次"
	}
	if len(input.Groups) == 0 {
		return Batch{}, nil, fmt.Errorf("%w: 至少添加一个书城", ErrInvalid)
	}
	seen := map[string]struct{}{}
	books := make([]Book, 0)
	for _, group := range input.Groups {
		source := strings.TrimSpace(group.Source)
		platformID := strings.TrimSpace(group.PlatformID)
		if source == "" || platformID == "" {
			return Batch{}, nil, fmt.Errorf("%w: 书城与 platformId 不能为空", ErrInvalid)
		}
		for _, item := range group.Books {
			bookID := strings.TrimSpace(item.BookID)
			if bookID == "" {
				return Batch{}, nil, fmt.Errorf("%w: Book ID 不能为空", ErrInvalid)
			}
			key := platformID + "_" + bookID
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			books = append(books, Book{
				Key: key, Source: source, PlatformID: platformID, ExternalBookID: bookID,
				Title: strings.TrimSpace(item.Title), Versions: map[string]string{}, Status: StatusPending,
			})
			if len(books) > 200 {
				return Batch{}, nil, fmt.Errorf("%w: 单批最多 200 本", ErrInvalid)
			}
		}
	}
	if len(books) == 0 {
		return Batch{}, nil, fmt.Errorf("%w: 至少添加一本小说", ErrInvalid)
	}
	now := s.now().UTC()
	batch := Batch{ID: newID("nf_batch"), Name: name, CreatedAt: now}
	for index := range books {
		books[index].BatchID = batch.ID
		batch.BookKeys = append(batch.BookKeys, books[index].Key)
	}
	stored, err := s.store.CreateBatch(ctx, batch)
	if err != nil {
		return Batch{}, nil, err
	}
	for _, book := range books {
		if err := s.store.PutBook(ctx, book); err != nil {
			return stored, nil, err
		}
	}
	return stored, books, nil
}

func (s *Service) StartRun(ctx context.Context, batchID string, runAt time.Time) (Run, error) {
	if s == nil || s.store == nil || s.dispatcher == nil {
		return Run{}, ErrRuntimeUnavailable
	}
	if _, err := s.store.GetBatch(ctx, strings.TrimSpace(batchID)); err != nil {
		return Run{}, err
	}
	config, err := s.store.GetConfig(ctx)
	if err != nil {
		return Run{}, err
	}
	config = normalizeConfig(config)
	now := s.now().UTC()
	when := runAt.UTC()
	status := StatusQueued
	if runAt.IsZero() {
		when = now
	} else if when.After(now) {
		status = StatusScheduled
	}
	run := Run{
		ID: newID("nf_run"), BatchID: batchID, Status: status, RunAt: when,
		ConfigSnapshot: config, CreatedAt: now, UpdatedAt: now,
	}
	run, err = s.store.CreateRun(ctx, run)
	if err != nil {
		return Run{}, err
	}
	if err := s.dispatcher.EnqueueRun(ctx, RunDispatch{RunID: run.ID, AvailableAt: when}); err != nil {
		run.Status = StatusBlocked
		run.UpdatedAt = s.now().UTC()
		_ = s.store.UpdateRun(ctx, run)
		return run, fmt.Errorf("%w: %v", ErrRuntimeUnavailable, err)
	}
	return run, nil
}

func (s *Service) ExecuteRun(ctx context.Context, runID string) (Run, error) {
	if s == nil || s.store == nil || s.fetcher == nil {
		return Run{}, ErrRuntimeUnavailable
	}
	run, err := s.store.GetRun(ctx, strings.TrimSpace(runID))
	if err != nil {
		return Run{}, err
	}
	now := s.now().UTC()
	if run.RunAt.After(now) {
		return run, ErrNotDue
	}
	run.Status = StatusRunning
	run.UpdatedAt = now
	if err := s.store.UpdateRun(ctx, run); err != nil {
		return run, err
	}
	books, err := s.store.ListBooks(ctx, run.BatchID)
	if err != nil {
		return run, err
	}
	for _, book := range books {
		_ = s.executeBook(ctx, run, book)
	}
	books, err = s.store.ListBooks(ctx, run.BatchID)
	if err != nil {
		return run, err
	}
	run.Status = aggregateRunStatus(books)
	run.UpdatedAt = s.now().UTC()
	if err := s.store.UpdateRun(ctx, run); err != nil {
		return run, err
	}
	return run, nil
}

func (s *Service) RetryBook(ctx context.Context, runID, bookKey string) (Book, error) {
	run, err := s.store.GetRun(ctx, strings.TrimSpace(runID))
	if err != nil {
		return Book{}, err
	}
	book, err := s.store.GetBook(ctx, run.BatchID, strings.TrimSpace(bookKey))
	if err != nil {
		return Book{}, err
	}
	if book.Status != StatusRetryableFailed && book.Status != StatusPartialFailed && book.Status != StatusFailed {
		return book, fmt.Errorf("%w: 当前书籍没有失败阶段", ErrInvalid)
	}
	if err := s.executeBook(ctx, run, book); err != nil {
		latest, getErr := s.store.GetBook(ctx, run.BatchID, book.Key)
		if getErr == nil {
			return latest, err
		}
		return book, err
	}
	return s.store.GetBook(ctx, run.BatchID, book.Key)
}

func (s *Service) executeBook(ctx context.Context, run Run, book Book) error {
	book.CurrentStage = "fetch"
	book.Status = StatusRunning
	book.Error = ""
	if err := s.store.UpdateBook(ctx, book); err != nil {
		return err
	}
	if strings.TrimSpace(book.OriginalRaw) == "" {
		started := s.now().UTC()
		result, err := s.fetcher.Fetch(ctx, FetchRequest{
			Source: book.Source, PlatformID: book.PlatformID, BookID: book.ExternalBookID,
		})
		if err != nil {
			book.Status = StatusRetryableFailed
			book.Error = safeError(err)
			book.CurrentStage = "fetch"
			_ = s.store.UpdateBook(ctx, book)
			s.record(ctx, run.ID, book.Key, "fetch", StatusFailed, book.Error, started)
			return err
		}
		book.OriginalRaw = result.OriginalRaw
		book.OriginalChars = utf8.RuneCountInString(result.OriginalRaw)
		if title := strings.TrimSpace(result.Title); title != "" {
			book.Title = title
		}
		book.Metadata = Metadata{
			Category: strings.TrimSpace(result.Category), Genre: strings.TrimSpace(result.Genre),
			Gender: strings.TrimSpace(result.Gender), Style: strings.TrimSpace(result.Style),
		}
		s.record(ctx, run.ID, book.Key, "fetch", StatusSucceeded, "", started)
	}
	started := s.now().UTC()
	book.CurrentStage = "rules"
	book.ProcessedText = applyRules(book.OriginalRaw, run.ConfigSnapshot)
	book.ProcessedChars = utf8.RuneCountInString(book.ProcessedText)
	if book.Versions == nil {
		book.Versions = map[string]string{}
	}
	if contains(run.ConfigSnapshot.TargetVersions, "original") {
		book.Versions["original"] = book.ProcessedText
	}
	s.record(ctx, run.ID, book.Key, "rules", StatusSucceeded, "", started)

	knowledge, err := s.store.ListKnowledge(ctx, "")
	if err != nil {
		book.Status = StatusRetryableFailed
		book.Error = safeError(err)
		book.CurrentStage = "knowledge"
		_ = s.store.UpdateBook(ctx, book)
		s.record(ctx, run.ID, book.Key, "knowledge", StatusFailed, book.Error, s.now().UTC())
		return err
	}
	var firstErr error
	for _, version := range run.ConfigSnapshot.TargetVersions {
		if version == "original" || strings.TrimSpace(book.Versions[version]) != "" {
			continue
		}
		started := s.now().UTC()
		stage := "rewrite:" + version
		book.CurrentStage = stage
		if s.model == nil || strings.TrimSpace(run.ConfigSnapshot.TextModelID) == "" {
			err := ErrModelUnavailable
			if firstErr == nil {
				firstErr = err
			}
			s.record(ctx, run.ID, book.Key, stage, StatusFailed, err.Error(), started)
			continue
		}
		text, err := s.model.Rewrite(ctx, RewriteRequest{
			Book: book, Version: version, Profile: run.ConfigSnapshot.RewriteProfiles[version],
			TextModelID: run.ConfigSnapshot.TextModelID, Knowledge: enabledKnowledge(knowledge),
		})
		if err != nil || strings.TrimSpace(text) == "" {
			if err == nil {
				err = fmt.Errorf("模型返回空文案")
			}
			if firstErr == nil {
				firstErr = err
			}
			s.record(ctx, run.ID, book.Key, stage, StatusFailed, safeError(err), started)
			continue
		}
		book.Versions[version] = text
		s.record(ctx, run.ID, book.Key, stage, StatusSucceeded, "", started)
	}
	if firstErr != nil {
		book.Status = StatusRetryableFailed
		if successfulVersionCount(book, run.ConfigSnapshot.TargetVersions) > 0 {
			book.Status = StatusPartialFailed
		}
		book.Error = safeError(firstErr)
		_ = s.store.UpdateBook(ctx, book)
		return firstErr
	}
	book.Status = StatusSucceeded
	book.CurrentStage = "completed"
	book.Error = ""
	return s.store.UpdateBook(ctx, book)
}

func (s *Service) PreviewRules(ctx context.Context, text string, config Config) string {
	_ = ctx
	return applyRules(text, normalizeConfig(config))
}

func (s *Service) GetConfig(ctx context.Context) (Config, error) {
	if s == nil || s.store == nil {
		return Config{}, ErrRuntimeUnavailable
	}
	cfg, err := s.store.GetConfig(ctx)
	return normalizeConfig(cfg), err
}

func (s *Service) SaveConfig(ctx context.Context, config Config) (Config, error) {
	if s == nil || s.store == nil {
		return Config{}, ErrRuntimeUnavailable
	}
	return s.store.SaveConfig(ctx, normalizeConfig(config))
}

func (s *Service) ListKnowledge(ctx context.Context, kind string) ([]KnowledgeEntry, error) {
	if s == nil || s.store == nil {
		return nil, ErrRuntimeUnavailable
	}
	return s.store.ListKnowledge(ctx, strings.TrimSpace(kind))
}

func (s *Service) UpsertKnowledge(ctx context.Context, entry KnowledgeEntry) (KnowledgeEntry, error) {
	if s == nil || s.store == nil {
		return KnowledgeEntry{}, ErrRuntimeUnavailable
	}
	entry.ID = strings.TrimSpace(entry.ID)
	if entry.ID == "" {
		entry.ID = newID("nf_knowledge")
	}
	entry.Kind = strings.TrimSpace(entry.Kind)
	entry.Title = strings.TrimSpace(entry.Title)
	entry.Content = strings.TrimSpace(entry.Content)
	if entry.Kind == "" || entry.Content == "" {
		return KnowledgeEntry{}, fmt.Errorf("%w: 知识库类型和内容不能为空", ErrInvalid)
	}
	return s.store.UpsertKnowledge(ctx, entry)
}

func (s *Service) DeleteKnowledge(ctx context.Context, kind, id string) error {
	if s == nil || s.store == nil {
		return ErrRuntimeUnavailable
	}
	return s.store.DeleteKnowledge(ctx, strings.TrimSpace(kind), strings.TrimSpace(id))
}

func (s *Service) Records(ctx context.Context, runID string) ([]Record, error) {
	if s == nil || s.store == nil {
		return nil, ErrRuntimeUnavailable
	}
	return s.store.ListRecords(ctx, strings.TrimSpace(runID))
}

func (s *Service) HandoffToBatchFactory(ctx context.Context, runID string) (HandoffResult, error) {
	if s == nil || s.batch == nil {
		return HandoffResult{}, ErrHandoffUnavailable
	}
	run, err := s.store.GetRun(ctx, runID)
	if err != nil {
		return HandoffResult{}, err
	}
	batch, err := s.store.GetBatch(ctx, run.BatchID)
	if err != nil {
		return HandoffResult{}, err
	}
	books, err := s.store.ListBooks(ctx, run.BatchID)
	if err != nil {
		return HandoffResult{}, err
	}
	return s.batch.CreateFromNovelFetch(ctx, HandoffRequest{Batch: batch, Run: run, Books: books})
}

func (s *Service) RequestNetworkSubmit(ctx context.Context, request SubmitIntentRequest) (SubmitIntentResult, error) {
	if s == nil || s.publisher == nil {
		return SubmitIntentResult{}, ErrPublishUnavailable
	}
	run, err := s.store.GetRun(ctx, request.RunID)
	if err != nil {
		return SubmitIntentResult{}, err
	}
	book, err := s.store.GetBook(ctx, run.BatchID, request.BookKey)
	if err != nil {
		return SubmitIntentResult{}, err
	}
	version := strings.TrimSpace(request.Version)
	if version == "" || strings.TrimSpace(book.Versions[version]) == "" {
		return SubmitIntentResult{}, fmt.Errorf("%w: 提交版本尚未完成", ErrInvalid)
	}
	request.BatchID = run.BatchID
	return s.publisher.CreateIntent(ctx, request)
}

func (s *Service) record(ctx context.Context, runID, bookKey, stage string, status Status, errText string, started time.Time) {
	if s.store == nil {
		return
	}
	records, _ := s.store.ListRecords(ctx, runID)
	attempt := 1
	for _, item := range records {
		if item.BookKey == bookKey && item.Stage == stage && item.Attempt >= attempt {
			attempt = item.Attempt + 1
		}
	}
	_ = s.store.AppendRecord(ctx, Record{
		RunID: runID, BookKey: bookKey, Stage: stage, Attempt: attempt,
		Status: status, Error: safeErrorString(errText), StartedAt: started, FinishedAt: s.now().UTC(),
	})
}

func normalizeConfig(config Config) Config {
	if config.MaxText <= 0 {
		config.MaxText = 4000
	}
	if config.MaxText > 100000 {
		config.MaxText = 100000
	}
	config.TextModelID = strings.TrimSpace(config.TextModelID)
	if len(config.TargetVersions) == 0 {
		config.TargetVersions = []string{"original", "ai1"}
	}
	seen := map[string]struct{}{}
	versions := make([]string, 0, len(config.TargetVersions))
	for _, version := range config.TargetVersions {
		version = strings.ToLower(strings.TrimSpace(version))
		if version != "original" && !strings.HasPrefix(version, "ai") {
			continue
		}
		if _, ok := seen[version]; ok {
			continue
		}
		seen[version] = struct{}{}
		versions = append(versions, version)
	}
	config.TargetVersions = versions
	if config.RewriteProfiles == nil {
		config.RewriteProfiles = map[string]string{}
	}
	if config.SensitiveReplacements == nil {
		config.SensitiveReplacements = map[string]string{}
	}
	return config
}

func applyRules(raw string, config Config) string {
	text := truncateRunes(strings.ReplaceAll(raw, "
", "
"), config.MaxText)
	lines := strings.Split(text, "
")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if config.TrimLines {
			line = strings.TrimSpace(line)
		}
		remove := false
		for _, prefix := range config.ChapterRemovePrefixes {
			if p := strings.TrimSpace(prefix); p != "" && strings.HasPrefix(strings.TrimSpace(line), p) {
				remove = true
				break
			}
		}
		if remove {
			continue
		}
		for from, to := range config.SensitiveReplacements {
			if from != "" {
				line = strings.ReplaceAll(line, from, to)
			}
		}
		if config.DropBlankLines && strings.TrimSpace(line) == "" {
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "
")
}

func truncateRunes(value string, max int) string {
	if max <= 0 {
		return value
	}
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}

func aggregateRunStatus(books []Book) Status {
	if len(books) == 0 {
		return StatusFailed
	}
	ok, failed := 0, 0
	for _, book := range books {
		switch book.Status {
		case StatusSucceeded:
			ok++
		case StatusFailed, StatusRetryableFailed, StatusPartialFailed, StatusBlocked:
			failed++
		default:
			return StatusRunning
		}
	}
	if ok == len(books) {
		return StatusSucceeded
	}
	if failed == len(books) {
		return StatusFailed
	}
	return StatusPartialFailed
}

func successfulVersionCount(book Book, targets []string) int {
	count := 0
	for _, version := range targets {
		if strings.TrimSpace(book.Versions[version]) != "" {
			count++
		}
	}
	return count
}

func enabledKnowledge(entries []KnowledgeEntry) []KnowledgeEntry {
	out := make([]KnowledgeEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.Enabled {
			out = append(out, entry)
		}
	}
	return out
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func safeError(err error) string {
	if err == nil {
		return ""
	}
	return safeErrorString(err.Error())
}

func safeErrorString(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	lower := strings.ToLower(value)
	for _, marker := range []string{"password", "token", "authorization", "bearer", "cookie", "secret", "api_key", "apikey", "mysql://", "dsn"} {
		if strings.Contains(lower, marker) {
			return "敏感错误详情已脱敏"
		}
	}
	runes := []rune(value)
	if len(runes) > 512 {
		return string(runes[:512]) + "…"
	}
	return value
}

func newID(prefix string) string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	}
	return prefix + "_" + hex.EncodeToString(buf)
}
