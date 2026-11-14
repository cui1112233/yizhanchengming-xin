package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/videomanagementconnector"
)

const workshopSubmitPauseKey = "novel_fetch_submit_account_pause"
const workshopLoginRequiredMessage = "视频管理系统登录已失效，此账号上传已暂停；重新登录成功后将继续上传已有文案"
const workshopContractChangedMessage = "视频管理系统接口结构已变化，上传已暂停；10分钟后只读探测，已有提交先回读确认"

func submitSessionFingerprint(session map[string]any) string {
	digest := sha256.Sum256([]byte(firstText(session["targetUsername"]) + "\n" + firstText(session["baseUrl"]) + "\n" + firstText(session["cookie"])))
	return hex.EncodeToString(digest[:])
}

func saveWorkshopSubmitPause(ctx context.Context, store novelfetchworkshop.Store, owner string, pause map[string]any) error {
	for attempt := 0; attempt < 3; attempt++ {
		cfg, err := store.GetConfig(ctx, owner)
		if err != nil {
			return err
		}
		if pause == nil {
			delete(cfg, workshopSubmitPauseKey)
		} else {
			cfg[workshopSubmitPauseKey] = pause
		}
		err = store.PutConfig(ctx, owner, cfg)
		if !errors.Is(err, novelfetchworkshop.ErrConfigConflict) {
			return err
		}
	}
	return fmt.Errorf("上传暂停状态保存冲突，请稍后重试")
}

func checkWorkshopSubmitPause(ctx context.Context, store novelfetchworkshop.Store, owner string, session map[string]any, target videomanagementconnector.Service) error {
	cfg, err := store.GetConfig(ctx, owner)
	if err != nil {
		return err
	}
	pause := mapOrEmpty(cfg[workshopSubmitPauseKey])
	if firstText(pause["session_fingerprint"]) == "" {
		return nil
	}
	contractChanged := firstText(pause["reason"]) == "site_contract_changed"
	if !contractChanged && firstText(pause["session_fingerprint"]) != submitSessionFingerprint(session) {
		return saveWorkshopSubmitPause(ctx, store, owner, nil)
	}
	probeAt, parseErr := time.Parse(time.RFC3339Nano, firstText(pause["next_probe_at"]))
	if parseErr == nil && !time.Now().Before(probeAt) && target != nil {
		probeSession := videomanagementconnector.Session{Cookie: firstText(session["cookie"]), TargetUsername: firstText(session["targetUsername"])}
		probeErr := target.CheckSession(ctx, probeSession)
		if contractChanged {
			probeErr = videomanagementconnector.ErrSiteContractChanged
			if probe, ok := target.(interface {
				ProbeSubmissionContract(context.Context, videomanagementconnector.Session) error
			}); ok {
				probeErr = probe.ProbeSubmissionContract(ctx, probeSession)
			}
		}
		if probeErr == nil {
			return resumeWorkshopPausedUploads(ctx, store, owner, contractChanged)
		}
		delay := 30 * time.Minute
		if contractChanged {
			delay = 10 * time.Minute
		}
		pause["next_probe_at"] = time.Now().Add(delay).UTC().Format(time.RFC3339Nano)
		if err := saveWorkshopSubmitPause(ctx, store, owner, pause); err != nil {
			return err
		}
	}
	if contractChanged {
		return workshopContractBlockedError()
	}
	return &webSubmitRequestError{Status: http.StatusConflict, Message: workshopLoginRequiredMessage, Code: "UPLOAD_LOGIN_REQUIRED"}
}

func workshopContractBlockedError() error {
	return &webSubmitRequestError{Status: http.StatusConflict, Message: workshopContractChangedMessage, Code: "UPLOAD_SITE_CONTRACT_CHANGED"}
}

func workshopSubmitBlockedStatus(err error) string {
	var blocked *webSubmitRequestError
	if errors.As(err, &blocked) {
		switch blocked.Code {
		case "UPLOAD_LOGIN_REQUIRED":
			return "waiting_login"
		case "UPLOAD_SITE_CONTRACT_CHANGED":
			return "waiting_configuration"
		}
	}
	return ""
}

func pauseWorkshopSubmitContract(ctx context.Context, store novelfetchworkshop.Store, owner string, session map[string]any) error {
	return saveWorkshopSubmitPause(ctx, store, owner, map[string]any{"session_fingerprint": submitSessionFingerprint(session), "reason": "site_contract_changed", "opened_at": time.Now().UTC().Format(time.RFC3339Nano), "next_probe_at": time.Now().Add(10 * time.Minute).UTC().Format(time.RFC3339Nano)})
}

func workshopSubmitRequiresLogin(err error) bool {
	var blocked *webSubmitRequestError
	return errors.As(err, &blocked) && blocked.Code == "UPLOAD_LOGIN_REQUIRED"
}

func pauseWorkshopSubmitAccount(ctx context.Context, store novelfetchworkshop.Store, owner string, session map[string]any) error {
	return saveWorkshopSubmitPause(ctx, store, owner, map[string]any{"session_fingerprint": submitSessionFingerprint(session), "reason": "auth_expired", "opened_at": time.Now().UTC().Format(time.RFC3339Nano), "next_probe_at": time.Now().Add(30 * time.Minute).UTC().Format(time.RFC3339Nano)})
}

func recoverWorkshopSubmitAccount(ctx context.Context, store novelfetchworkshop.Store, queue novelfetchworkshop.ProcessQueue, owner, usersDir string, target videomanagementconnector.Service) {
	if target == nil {
		return
	}
	cfg, err := store.GetConfig(ctx, owner)
	if err != nil {
		return
	}
	pause := mapOrEmpty(cfg[workshopSubmitPauseKey])
	if firstText(pause["session_fingerprint"]) == "" {
		return
	}
	when, err := time.Parse(time.RFC3339Nano, firstText(pause["next_probe_at"]))
	if err != nil || time.Now().Before(when) {
		return
	}
	probeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	session := readUploadTargetSession(usersDir, owner)
	release, err := acquireWorkshopSubmit(probeCtx, queue, workshopExternalSubmitScope(session), "")
	if err != nil {
		return
	}
	defer release()
	if err := checkWorkshopSubmitPause(probeCtx, store, owner, session, target); err == nil {
		_ = resumeWorkshopPausedUploads(probeCtx, store, owner, firstText(pause["reason"]) == "site_contract_changed")
	}
}

// Called only after a successful connector login. Waiting uploads retain their
// body/version checkpoints; cancelled work and other owners are not touched.
func resumeWorkshopUploadsAfterLogin(ctx context.Context, store novelfetchworkshop.Store, owner string) error {
	// Authentication alone does not prove a changed upload API has recovered.
	if store != nil {
		cfg, err := store.GetConfig(ctx, owner)
		if err != nil {
			return err
		}
		if firstText(mapOrEmpty(cfg[workshopSubmitPauseKey])["reason"]) == "site_contract_changed" {
			return nil
		}
	}
	return resumeWorkshopPausedUploads(ctx, store, owner, false)
}

func resumeWorkshopPausedUploads(ctx context.Context, store novelfetchworkshop.Store, owner string, contract bool) error {
	if store == nil {
		return nil
	}
	docs, err := store.ListDocuments(ctx, owner)
	if err != nil {
		return err
	}
	for _, doc := range docs {
		waiting := firstText(doc.Meta["siteSubmitStatus"]) == "waiting_login"
		if contract {
			waiting = firstText(doc.Meta["siteSubmitPauseReason"]) == "site_contract_changed"
		}
		if !waiting {
			continue
		}
		status := strings.ToLower(firstText(doc.Meta["status"], doc.Meta["queueState"]))
		if status == "stopped" || status == "cancelled" {
			continue
		}
		doc.Meta["siteSubmitStatus"] = "pending_upload"
		if len(mapOrEmpty(doc.Meta["siteSubmitPendingConfirmation"])) > 0 {
			doc.Meta["siteSubmitStatus"] = "pending_confirmation"
		}
		delete(doc.Meta, "siteSubmitPauseReason")
		doc.Meta["status"], doc.Meta["queueState"] = "upload_pending", "upload_pending"
		delete(doc.Meta, "automationNextRetryAt")
		delete(doc.Meta, "automationSubmitError")
		delete(doc.Meta, "error")
		if contract {
			appendNovelFetchDocumentLog(&doc, "upload_contract_restored", "接口只读探测已恢复，等待核对和继续上传已有文案", time.Now())
		} else {
			appendNovelFetchDocumentLog(&doc, "upload_login_restored", "登录已恢复，等待继续上传已有文案", time.Now())
		}
		if err := store.PutDocument(ctx, owner, doc); err != nil {
			return err
		}
	}
	return saveWorkshopSubmitPause(ctx, store, owner, nil)
}
