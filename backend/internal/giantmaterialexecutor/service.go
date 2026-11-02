package giantmaterialexecutor

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{store: store, now: now}
}

func (s *Service) CreatePairing(ctx context.Context, owner, platform string) (PairingSecret, error) {
	owner = strings.TrimSpace(owner)
	platform = strings.ToLower(strings.TrimSpace(platform))
	if owner == "" {
		return PairingSecret{}, ErrInvalidInput
	}
	if platform != PlatformGiantMaterial {
		return PairingSecret{}, ErrInvalidPlatform
	}
	code, err := randomCode()
	if err != nil {
		return PairingSecret{}, err
	}
	now := s.now().UTC()
	expires := now.Add(PairingTTL)
	if err := s.store.CreatePairing(ctx, PairingRecord{ID: randomID("gme_pair_"), OwnerUsername: owner, Platform: platform, CodeHash: hashPairingCode(code), ExpiresAt: expires, CreatedAt: now}); err != nil {
		return PairingSecret{}, err
	}
	return PairingSecret{Code: code, ExpiresAt: expires}, nil
}

func (s *Service) Pair(ctx context.Context, input PairInput) (PairResult, error) {
	if strings.ToLower(strings.TrimSpace(input.Platform)) != PlatformGiantMaterial || normalizePairingCode(input.Code) == "" {
		if strings.ToLower(strings.TrimSpace(input.Platform)) != PlatformGiantMaterial {
			return PairResult{}, ErrInvalidPlatform
		}
		return PairResult{}, ErrPairingInvalid
	}
	if err := validateDeviceFields(input.DeviceName, input.OS, input.Version); err != nil {
		return PairResult{}, err
	}
	osName := strings.ToLower(strings.TrimSpace(input.OS))
	if !validExecutorOS(osName) {
		return PairResult{}, ErrInvalidInput
	}
	token, err := randomToken()
	if err != nil {
		return PairResult{}, err
	}
	executor := ExecutorRecord{ID: randomID("gme_exec_"), Platform: PlatformGiantMaterial, TokenHash: hashSecret(token), DeviceName: strings.TrimSpace(input.DeviceName), OS: osName, Version: strings.TrimSpace(input.Version), CreatedAt: s.now().UTC(), UpdatedAt: s.now().UTC()}
	created, err := s.store.PairExecutor(ctx, hashPairingCode(input.Code), PlatformGiantMaterial, executor, s.now().UTC())
	if err != nil {
		return PairResult{}, err
	}
	return PairResult{ExecutorID: created.ID, Token: token, HeartbeatIntervalSeconds: HeartbeatIntervalSeconds}, nil
}

func (s *Service) Heartbeat(ctx context.Context, token string, input HeartbeatInput) error {
	executor, err := s.executorForToken(ctx, token)
	if err != nil {
		return err
	}
	if err := validateDeviceFields(input.DeviceName, input.OS, input.Version); err != nil {
		return err
	}
	return s.store.UpdateHeartbeat(ctx, executor.ID, input, s.now().UTC())
}

func (s *Service) ListExecutors(ctx context.Context, owner string) ([]ExecutorView, error) {
	owner = strings.TrimSpace(owner)
	if owner == "" {
		return nil, ErrInvalidInput
	}
	records, err := s.store.ListExecutors(ctx, owner)
	if err != nil {
		return nil, err
	}
	now := s.now().UTC()
	failures, err := s.store.RecentFailures(ctx, owner, now.Add(-FailureCooldown))
	if err != nil {
		return nil, err
	}
	views := make([]ExecutorView, 0, len(records))
	for _, record := range records {
		if value, ok := failures[record.ID]; ok {
			copyValue := value
			views = append(views, ExecutorView{
				ID: record.ID, Name: record.DeviceName, Platform: record.Platform, OS: record.OS,
				Version: record.Version, Online: record.LastSeenAt != nil && !record.LastSeenAt.Before(now.Add(-OnlineThreshold)),
				LastSeenAt: record.LastSeenAt, RecentFailureAt: &copyValue,
			})
			continue
		}
		views = append(views, ExecutorView{ID: record.ID, Name: record.DeviceName, Platform: record.Platform, OS: record.OS, Version: record.Version, Online: record.LastSeenAt != nil && !record.LastSeenAt.Before(now.Add(-OnlineThreshold)), LastSeenAt: record.LastSeenAt})
	}
	return views, nil
}

func (s *Service) ConsolidateExecutors(ctx context.Context) (int, error) {
	return s.store.ConsolidateDuplicateExecutors(ctx, s.now().UTC())
}

func (s *Service) DeleteExecutor(ctx context.Context, owner, id string) error {
	owner = strings.TrimSpace(owner)
	id = strings.TrimSpace(id)
	if owner == "" || id == "" {
		return ErrInvalidInput
	}
	return s.store.DeleteExecutor(ctx, owner, id, s.now().UTC())
}

// EnsureIdentityUniqueIndex 包装 store 同名方法（0a260d61 引入；Service 层需补此方法供 app.go 调用）
func (s *Service) EnsureIdentityUniqueIndex(ctx context.Context) error {
	return s.store.EnsureIdentityUniqueIndex(ctx)
}

func (s *Service) GetPreference(ctx context.Context, owner string) (PreferenceView, error) {
	owner = strings.TrimSpace(owner)
	if owner == "" {
		return PreferenceView{}, ErrInvalidInput
	}
	record, err := s.store.GetPreference(ctx, owner)
	if errors.Is(err, ErrPreferenceNotFound) {
		return PreferenceView{PreferredOS: DefaultPreferredOS}, nil
	}
	if err != nil {
		return PreferenceView{}, err
	}
	return PreferenceView{PreferredOS: record.PreferredOS, UpdatedAt: record.UpdatedAt}, nil
}

func (s *Service) SavePreference(ctx context.Context, owner, preferredOS string) (PreferenceView, error) {
	owner = strings.TrimSpace(owner)
	preferredOS = strings.ToLower(strings.TrimSpace(preferredOS))
	if owner == "" {
		return PreferenceView{}, ErrInvalidInput
	}
	if !validExecutorOS(preferredOS) {
		return PreferenceView{}, ErrInvalidInput
	}
	now := s.now().UTC()
	record := PreferenceRecord{OwnerUsername: owner, PreferredOS: preferredOS, UpdatedAt: now}
	if err := s.store.SavePreference(ctx, record); err != nil {
		return PreferenceView{}, err
	}
	return PreferenceView{PreferredOS: preferredOS, UpdatedAt: now}, nil
}

func validExecutorOS(value string) bool {
	return value == "windows" || value == "darwin"
}

func (s *Service) CreateJob(ctx context.Context, owner string, input CreateJobInput) (JobView, error) {
	owner = strings.TrimSpace(owner)
	if owner == "" || strings.TrimSpace(input.MaterialID) == "" || strings.TrimSpace(input.PlatformBookID) == "" || strings.TrimSpace(input.Title) == "" || strings.TrimSpace(input.ModelVersion) == "" || input.DurationSeconds <= 0 || input.DurationSeconds > 1800 {
		return JobView{}, ErrInvalidInput
	}
	if strings.ToLower(strings.TrimSpace(input.Platform)) != PlatformGiantMaterial {
		return JobView{}, ErrInvalidPlatform
	}
	if err := validateVideoURL(input.VideoURL); err != nil {
		return JobView{}, err
	}
	if input.VideoExpiresAt != nil && !input.VideoExpiresAt.After(s.now().UTC()) {
		return JobView{}, ErrInvalidInput
	}
	now := s.now().UTC()
	// 新任务钉选：用户显式设置过平台偏好时，优先该平台里最新、未冷却的
	// 设备，偏好平台没人能干活才兜底其他平台；从未设置过偏好（只有系统
	// 默认值）时沿用"钉给最新在线设备"的老行为；全都冷却时留空走公共池。
	preferredOS := ""
	if preferenceRecord, preferenceErr := s.store.GetPreference(ctx, owner); preferenceErr == nil {
		preferredOS = preferenceRecord.PreferredOS
	} else if !errors.Is(preferenceErr, ErrPreferenceNotFound) {
		return JobView{}, preferenceErr
	}
	targetExecutorID, err := s.pickTargetExecutor(ctx, owner, preferredOS, now)
	if err != nil {
		return JobView{}, err
	}
	record := JobRecord{ID: randomID("gme_job_"), OwnerUsername: owner, Platform: PlatformGiantMaterial, MaterialID: strings.TrimSpace(input.MaterialID), PlatformBookID: strings.TrimSpace(input.PlatformBookID), Title: bounded(input.Title, 191), VideoURL: strings.TrimSpace(input.VideoURL), VideoExpiresAt: input.VideoExpiresAt, DurationSeconds: input.DurationSeconds, ModelVersion: bounded(input.ModelVersion, 64), ContentRangeLines: bounded(input.ContentRangeLines, 64), TargetExecutorID: targetExecutorID, State: JobQueued, CreatedAt: now, UpdatedAt: now}
	key := jobKey(record)
	if existing, err := s.store.FindJobByKey(ctx, owner, key); err == nil {
		// 同 key 的失败/取消任务允许重试：用新的视频地址重置为排队状态。
		if existing.State == JobFailed || existing.State == JobCancelled {
			if requeued, retryErr := s.store.RequeueJob(ctx, existing.ID, record, s.now().UTC()); retryErr == nil {
				return s.jobViewForOwner(ctx, owner, requeued)
			} else if !errors.Is(retryErr, ErrJobNotFound) {
				return JobView{}, retryErr
			}
		}
		return s.jobViewForOwner(ctx, owner, existing)
	} else if err != ErrJobNotFound {
		return JobView{}, err
	}
	if err := s.store.CreateJob(ctx, record); err != nil {
		if err == ErrJobConflict {
			if existing, lookupErr := s.store.FindJobByKey(ctx, owner, key); lookupErr == nil {
				return s.jobViewForOwner(ctx, owner, existing)
			}
		}
		return JobView{}, err
	}
	return s.jobViewForOwner(ctx, owner, record)
}

// pickTargetExecutor 选择新任务的钉选目标。preferredOS 非空时先在该平台
// 的在线、未冷却设备里选最新；偏好平台没有能干活的设备才兜底其他平台；
// 全都不可用时返回空串，任务留在公共队列，待执行器恢复后自动领取。
// preferredOS 为空表示不按偏好过滤（显式 RetryJob 沿用"选最新在线设备"语义）。
func (s *Service) pickTargetExecutor(ctx context.Context, owner, preferredOS string, now time.Time) (string, error) {
	executors, err := s.store.ListExecutors(ctx, owner)
	if err != nil {
		return "", err
	}
	pick := func(wantOS string) string {
		var chosen *ExecutorRecord
		for index := range executors {
			candidate := &executors[index]
			if candidate.Platform != PlatformGiantMaterial || candidate.LastSeenAt == nil || candidate.LastSeenAt.Before(now.Add(-OnlineThreshold)) {
				continue
			}
			if wantOS != "" && candidate.OS != wantOS {
				continue
			}

			if failure, failureErr := s.store.LatestPlatformFailure(ctx, owner, candidate.OS, now.Add(-FailureCooldown)); failureErr != nil || failure != nil {
				continue
			}
			if chosen == nil || candidate.LastSeenAt.After(*chosen.LastSeenAt) || (candidate.LastSeenAt.Equal(*chosen.LastSeenAt) && candidate.ID > chosen.ID) {
				chosen = candidate
			}
		}
		if chosen == nil {
			return ""
		}
		return chosen.ID
	}
	if preferredOS != "" {
		if id := pick(preferredOS); id != "" {
			return id, nil
		}
	}
	return pick(""), nil
}

func (s *Service) GetJob(ctx context.Context, owner, id string) (JobView, error) {
	owner = strings.TrimSpace(owner)
	record, err := s.store.JobForOwner(ctx, owner, strings.TrimSpace(id))
	if err != nil {
		return JobView{}, err
	}
	return s.jobViewForOwner(ctx, owner, record)
}

func (s *Service) jobViewForOwner(ctx context.Context, owner string, record JobRecord) (JobView, error) {
	view := jobView(record)
	if record.State != JobQueued {
		return view, nil
	}
	availableExecutorID, err := s.pickTargetExecutor(ctx, owner, "", s.now().UTC())
	if err != nil {
		return JobView{}, err
	}
	if availableExecutorID != "" {
		return view, nil
	}
	view.WaitingForExecutor = true
	return view, nil
}

func (s *Service) CancelJob(ctx context.Context, owner, id string) (JobView, error) {
	record, err := s.store.CancelJob(ctx, strings.TrimSpace(owner), strings.TrimSpace(id), s.now().UTC())
	if err != nil {
		return JobView{}, err
	}
	return jobView(record), nil
}

// RetryJob returns a failed or cancelled job to the queue without requiring the
// client to resend the temporary media URL. It also selects the newest online
// executor again, so an expired lease cannot keep a retry pinned to an older
// machine.
func (s *Service) RetryJob(ctx context.Context, owner, id string) (JobView, error) {
	owner = strings.TrimSpace(owner)
	if owner == "" || strings.TrimSpace(id) == "" {
		return JobView{}, ErrInvalidInput
	}
	record, err := s.store.JobForOwner(ctx, owner, strings.TrimSpace(id))
	if err != nil {
		return JobView{}, err
	}
	if record.State != JobFailed && record.State != JobCancelled {
		return JobView{}, ErrInvalidJobState
	}
	// 显式重试不按平台偏好过滤（用户手动重试通常就想交给最新上线的设备），
	// 但仍绕开正在失败冷却中的设备；全都冷却时留空走公共池。
	record.TargetExecutorID, err = s.pickTargetExecutor(ctx, owner, "", s.now().UTC())
	if err != nil {
		return JobView{}, err
	}
	requeued, err := s.store.RequeueJob(ctx, record.ID, record, s.now().UTC())
	if err != nil {
		return JobView{}, err
	}
	return s.jobViewForOwner(ctx, owner, requeued)
}

func (s *Service) Claim(ctx context.Context, token string) (ClaimResult, error) {
	executor, err := s.executorForToken(ctx, token)
	if err != nil {
		return ClaimResult{}, err
	}
	now := s.now().UTC()
	leaseToken, err := randomToken()
	if err != nil {
		return ClaimResult{}, err
	}
	expires := now.Add(JobLeaseTTL)
	// 第一步：明确点名给本执行器的定向重试任务优先领取。平台偏好门控
	// 不拦截显式定向（任务指定谁就是谁），但失败冷却这道安全门仍然生效。
	callerFailure, err := s.store.LatestPlatformFailure(ctx, executor.OwnerUsername, executor.OS, now.Add(-FailureCooldown))
	if err != nil {
		return ClaimResult{}, err
	}
	if callerFailure == nil {
		record, claimErr := s.store.ClaimJob(ctx, executor, hashSecret(leaseToken), expires, now, true)
		if claimErr == nil {
			return ClaimResult{Job: executorJobView(record), LeaseToken: leaseToken, LeaseGeneration: record.LeaseGeneration, LeaseExpiresAt: expires}, nil
		}
		if !errors.Is(claimErr, ErrNoClaimableJob) {
			return ClaimResult{}, claimErr
		}
	}
	// 第二步：没有点名任务时，才走平台偏好门控领取公共池任务。
	allowed, err := s.canClaimPlatform(ctx, executor, now)
	if err != nil {
		return ClaimResult{}, err
	}
	if !allowed {
		return ClaimResult{}, ErrNoClaimableJob
	}
	record, err := s.store.ClaimJob(ctx, executor, hashSecret(leaseToken), expires, now, false)
	if err != nil {
		return ClaimResult{}, err
	}
	return ClaimResult{Job: executorJobView(record), LeaseToken: leaseToken, LeaseGeneration: record.LeaseGeneration, LeaseExpiresAt: expires}, nil
}

// ClaimLongPollWait 是没活时 claim 请求在服务器侧最多等待的时间。执行器
// 用的是无超时 HTTP 客户端，Nginx/Node 转发层余量都大于该值，老版本执行器
// 无需升级也会被自动“按住”，空转敲门从每秒一次降到每个等待窗口一次。
const ClaimLongPollWait = 25 * time.Second

// claimLongPollInterval 是等待期间复查数据库的节奏。新任务出现后最坏延迟
// 一个 interval（OCR 任务本身耗时约十分钟，2 秒领取延迟可忽略）。
var claimLongPollInterval = 2 * time.Second

// ClaimWhenAvailable 先立刻试领；没活时在服务器侧等待最多 wait，期间定期
// 复查，一有任务立刻返回；到点仍无任务才返回 ErrNoClaimableJob。客户端
// 断开（ctx 取消）时立即退出，不继续翻数据库。
func (s *Service) ClaimWhenAvailable(ctx context.Context, token string, wait time.Duration) (ClaimResult, error) {
	result, err := s.Claim(ctx, token)
	if err == nil || !errors.Is(err, ErrNoClaimableJob) || wait <= 0 {
		return result, err
	}
	remaining := wait
	for remaining > 0 {
		step := claimLongPollInterval
		if step > remaining {
			step = remaining
		}
		select {
		case <-ctx.Done():
			return ClaimResult{}, ctx.Err()
		case <-time.After(step):
		}
		remaining -= step
		result, err = s.Claim(ctx, token)
		if err == nil || !errors.Is(err, ErrNoClaimableJob) {
			return result, err
		}
	}
	return ClaimResult{}, ErrNoClaimableJob
}

func (s *Service) canClaimPlatform(ctx context.Context, executor ExecutorRecord, now time.Time) (bool, error) {
	// 自己刚干砸：先冷却，任何平台都一样
	callerFailure, err := s.store.LatestPlatformFailure(ctx, executor.OwnerUsername, executor.OS, now.Add(-FailureCooldown))
	if err != nil {
		return false, err
	}
	if callerFailure != nil {
		return false, nil
	}
	preference, err := s.GetPreference(ctx, executor.OwnerUsername)
	if err != nil {
		return false, err
	}
	preferred := preference.PreferredOS
	if executor.OS == preferred {
		// 它就是偏好平台且没在冷却：恢复优先
		return true, nil
	}
	preferredFailure, err := s.store.LatestPlatformFailure(ctx, executor.OwnerUsername, preferred, now.Add(-FailureCooldown))
	if err != nil {
		return false, err
	}
	if preferredFailure != nil {
		// 偏好平台在冷却：非偏好平台兜底
		return true, nil
	}
	executors, err := s.store.ListExecutors(ctx, executor.OwnerUsername)
	if err != nil {
		return false, err
	}
	for _, item := range executors {
		if item.OS == preferred && item.LastSeenAt != nil && !item.LastSeenAt.Before(now.Add(-OnlineThreshold)) {
			// 偏好平台在岗：单子留给它
			return false, nil
		}
	}
	// 偏好平台不在线：兜底
	return true, nil
}

func (s *Service) Renew(ctx context.Context, token, jobID string, lease LeaseCredential) (LeaseView, error) {
	executor, err := s.executorForToken(ctx, token)
	if err != nil {
		return LeaseView{}, err
	}
	if strings.TrimSpace(lease.Token) == "" || lease.Generation < 1 {
		return LeaseView{}, ErrStaleLease
	}
	expires := s.now().UTC().Add(JobLeaseTTL)
	record, err := s.store.RenewJob(ctx, executor.ID, strings.TrimSpace(jobID), hashSecret(lease.Token), lease.Generation, expires, s.now().UTC())
	if err != nil {
		return LeaseView{}, err
	}
	return LeaseView{LeaseGeneration: record.LeaseGeneration, LeaseExpiresAt: expires}, nil
}

func (s *Service) Progress(ctx context.Context, token, jobID string, lease LeaseCredential, next JobState, progress ProgressInput) error {
	executor, err := s.executorForToken(ctx, token)
	if err != nil {
		return err
	}
	if !validProgress(next) || progress.Completed < 0 || progress.Total < 0 || progress.Completed > progress.Total || progress.Percent < 0 || progress.Percent > 100 {
		return ErrInvalidInput
	}
	_, err = s.store.SetProgress(ctx, executor.ID, strings.TrimSpace(jobID), hashSecret(lease.Token), lease.Generation, next, progress, s.now().UTC())
	return err
}

func (s *Service) Complete(ctx context.Context, token, jobID string, lease LeaseCredential, input ResultInput) (JobView, error) {
	executor, err := s.executorForToken(ctx, token)
	if err != nil {
		return JobView{}, err
	}
	if strings.TrimSpace(input.Text) == "" {
		return JobView{}, ErrInvalidInput
	}
	if len([]byte(input.Text)) > MaxResultBytes {
		return JobView{}, ErrResultTooLarge
	}
	record, err := s.store.CompleteJob(ctx, executor.ID, strings.TrimSpace(jobID), hashSecret(lease.Token), lease.Generation, input, s.now().UTC())
	if err != nil {
		return JobView{}, err
	}
	return jobView(record), nil
}

func (s *Service) Fail(ctx context.Context, token, jobID string, lease LeaseCredential, input FailureInput) error {
	executor, err := s.executorForToken(ctx, token)
	if err != nil {
		return err
	}
	_, err = s.store.FailJob(ctx, executor.ID, strings.TrimSpace(jobID), hashSecret(lease.Token), lease.Generation, FailureInput{Code: bounded(input.Code, 96), Message: bounded(input.Message, 512)}, s.now().UTC())
	return err
}

func (s *Service) executorForToken(ctx context.Context, token string) (ExecutorRecord, error) {
	if strings.TrimSpace(token) == "" {
		return ExecutorRecord{}, ErrExecutorUnauthorized
	}
	record, err := s.store.ExecutorByTokenHash(ctx, hashSecret(strings.TrimSpace(token)))
	if err != nil {
		return ExecutorRecord{}, ErrExecutorUnauthorized
	}
	return record, nil
}

func jobView(record JobRecord) JobView {
	return JobView{ID: record.ID, Platform: record.Platform, MaterialID: record.MaterialID, PlatformBookID: record.PlatformBookID, Title: record.Title, ModelVersion: record.ModelVersion, ContentRangeLines: record.ContentRangeLines, TargetExecutorID: record.TargetExecutorID, State: record.State, CancelRequested: record.CancelRequested, LeaseExecutorID: record.LeaseExecutorID, LeaseGeneration: record.LeaseGeneration, LeaseExpiresAt: record.LeaseExpiresAt, Progress: record.Progress, Result: record.Result, ErrorCode: record.ErrorCode, ErrorMessage: record.ErrorMessage, CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt}
}

func executorJobView(record JobRecord) ExecutorJobView {
	return ExecutorJobView{ID: record.ID, Platform: record.Platform, MaterialID: record.MaterialID, PlatformBookID: record.PlatformBookID, Title: record.Title, VideoURL: record.VideoURL, VideoExpiresAt: record.VideoExpiresAt, DurationSeconds: record.DurationSeconds, ModelVersion: record.ModelVersion, ContentRangeLines: record.ContentRangeLines, State: record.State}
}

func normalizePairingCode(code string) string {
	return strings.ToUpper(strings.NewReplacer("-", "", " ", "", "_", "").Replace(strings.TrimSpace(code)))
}

func validateDeviceFields(name, osName, version string) error {
	if len(strings.TrimSpace(name)) > 191 || len(strings.TrimSpace(osName)) > 32 || len(strings.TrimSpace(version)) > 64 {
		return ErrInvalidInput
	}
	return nil
}

func validateVideoURL(raw string) error {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme != "https" || !strings.HasSuffix(parsed.Hostname(), "material.hnqingyuwen.top") || parsed.Path == "" {
		return ErrInvalidInput
	}
	return nil
}

func validProgress(state JobState) bool {
	switch state {
	case JobRunning, JobCleaning, JobUploading:
		return true
	default:
		return false
	}
}

func validJobTransition(from, to JobState) error {
	if from == to {
		return nil
	}
	switch {
	case from == JobLeased && to == JobRunning:
	case from == JobRunning && to == JobCleaning:
	case from == JobCleaning && to == JobUploading:
	default:
		return ErrInvalidJobState
	}
	return nil
}

const pairingAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func randomCode() (string, error) {
	bytes := make([]byte, 10)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	for i := range bytes {
		bytes[i] = pairingAlphabet[int(bytes[i])%len(pairingAlphabet)]
	}
	return string(bytes[:5]) + "-" + string(bytes[5:]), nil
}

func randomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func randomID(prefix string) string {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%s%x", prefix, time.Now().UnixNano())
	}
	return prefix + base64.RawURLEncoding.EncodeToString(bytes)
}

type executorCandidate struct {
	id        string
	version   string
	lastSeen  *time.Time
	updatedAt time.Time
}

func pickKeeping(list []executorCandidate) string {
	target := list[0]
	for _, item := range list[1:] {
		cmp := compareVersion(item.version, target.version)
		if cmp > 0 ||
			(cmp == 0 && compareSeen(item.lastSeen, target.lastSeen) > 0) ||
			(cmp == 0 && compareSeen(item.lastSeen, target.lastSeen) == 0 && item.updatedAt.After(target.updatedAt)) {
			target = item
		}
	}
	return target.id
}

func compareVersion(a, b string) int {
	pa := strings.Split(strings.TrimSpace(a), ".")
	pb := strings.Split(strings.TrimSpace(b), ".")
	n := len(pa)
	if len(pb) > n {
		n = len(pb)
	}
	for i := 0; i < n; i++ {
		var x, y int
		if i < len(pa) {
			x = atoiOrZero(pa[i])
		}
		if i < len(pb) {
			y = atoiOrZero(pb[i])
		}
		if x != y {
			if x > y {
				return 1
			}
			return -1
		}
	}
	return 0
}

func atoiOrZero(value string) int {
	value = strings.TrimSpace(value)
	out := 0
	for _, r := range value {
		if r < '0' || r > '9' {
			return 0
		}
		out = out*10 + int(r-'0')
	}
	return out
}

func compareSeen(a, b *time.Time) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}
	if a.After(*b) {
		return 1
	}
	if b.After(*a) {
		return -1
	}
	return 0
}
