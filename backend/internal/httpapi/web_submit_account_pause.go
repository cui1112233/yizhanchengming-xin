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
	if firstText(pause["session_fingerprint"]) != submitSessionFingerprint(session) {
		return saveWorkshopSubmitPause(ctx, store, owner, nil)
	}
	probeAt, parseErr := time.Parse(time.RFC3339Nano, firstText(pause["next_probe_at"]))
	if parseErr == nil && !time.Now().Before(probeAt) && target != nil {
		if err := target.CheckSession(ctx, videomanagementconnector.Session{Cookie: firstText(session["cookie"]), TargetUsername: firstText(session["targetUsername"])}); err == nil {
			return saveWorkshopSubmitPause(ctx, store, owner, nil)
		}
		pause["next_probe_at"] = time.Now().Add(30 * time.Minute).UTC().Format(time.RFC3339Nano)
		if err := saveWorkshopSubmitPause(ctx, store, owner, pause); err != nil {
			return err
		}
	}
	return &webSubmitRequestError{Status: http.StatusConflict, Message: workshopLoginRequiredMessage, Code: "UPLOAD_LOGIN_REQUIRED"}
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
	release, err := acquireWorkshopSubmit(probeCtx, queue, owner, "")
	if err != nil {
		return
	}
	defer release()
	if err := checkWorkshopSubmitPause(probeCtx, store, owner, readUploadTargetSession(usersDir, owner), target); err == nil {
		_ = resumeWorkshopUploadsAfterLogin(probeCtx, store, owner)
	}
}

// Called only after a successful connector login. Waiting uploads retain their
// body/version checkpoints; cancelled work and other owners are not touched.
func resumeWorkshopUploadsAfterLogin(ctx context.Context, store novelfetchworkshop.Store, owner string) error {
	if store == nil {
		return nil
	}
	if err := saveWorkshopSubmitPause(ctx, store, owner, nil); err != nil {
		return err
	}
	docs, err := store.ListDocuments(ctx, owner)
	if err != nil {
		return err
	}
	for _, doc := range docs {
		if firstText(doc.Meta["siteSubmitStatus"]) != "waiting_login" {
			continue
		}
		status := strings.ToLower(firstText(doc.Meta["status"], doc.Meta["queueState"]))
		if status == "stopped" || status == "cancelled" {
			continue
		}
		doc.Meta["siteSubmitStatus"] = "pending_upload"
		doc.Meta["status"], doc.Meta["queueState"] = "upload_pending", "upload_pending"
		delete(doc.Meta, "automationNextRetryAt")
		delete(doc.Meta, "automationSubmitError")
		delete(doc.Meta, "error")
		appendNovelFetchDocumentLog(&doc, "upload_login_restored", "登录已恢复，等待继续上传已有文案", time.Now())
		if err := store.PutDocument(ctx, owner, doc); err != nil {
			return err
		}
	}
	return nil
}
