package httpapi

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/screenplay"
	"strings"
	"time"
)

// persistExport copies the completed renderer output to the same durable object
// store used by project media. It never starts a render or a provider request.
func (s *ScreenplayMediaWorkbench) persistExport(ctx context.Context, p screenplay.Project, job aicut.ExportJob, retry bool) (screenplay.MediaItem, error) {
	var out screenplay.MediaItem
	if s.Objects == nil || s.Cut == nil || s.Media == nil {
		return out, fmt.Errorf("成片持久存储未配置")
	}
	if job.ProjectID != p.ID || job.Owner != p.OwnerUsername || job.State != aicut.ExportSucceeded {
		return out, fmt.Errorf("成片尚未完成或不属于此作品")
	}
	token := newAgentID("export-persist")
	lock := "screenplay-export-persist:" + p.OwnerUsername + ":" + p.ID + ":" + job.ID
	if e := s.Projects.AcquireLock(ctx, lock, token, 5*time.Minute); e != nil {
		return out, e
	}
	defer s.Projects.ReleaseLock(context.WithoutCancel(ctx), lock, token)
	st, e := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if e != nil {
		return out, e
	}
	for _, old := range st.Items {
		if old.ID == job.ID && old.Kind == "export" {
			if old.Status == "succeeded" && old.AssetID != "" {
				return old, nil
			}
			if !retry {
				return old, fmt.Errorf("%s", old.Error)
			}
		}
	}
	hash := sha256.Sum256([]byte(p.OwnerUsername + ":" + p.ID + ":" + job.ID))
	id := fmt.Sprintf("sp-final-%x", hash[:20])
	out = screenplay.MediaItem{ID: job.ID, Kind: "export", Name: "MP4成片", Status: "failed", ProviderID: job.ID, CreatedAt: job.CreatedAt}
	asset, assetErr := s.Cut.Asset(ctx, p.OwnerUsername, p.ID, id)
	if assetErr == nil {
		body, openErr := s.Objects.Open(ctx, aicutObjectAsset(asset))
		if openErr == nil {
			body.Close()
			out.AssetID = id
			out.Status = "succeeded"
			out.DurationMS = asset.DurationMS
		}
	}
	if out.Status != "succeeded" {
		e = s.ensureCut(ctx, p)
		if e == nil {
			var file *os.File
			file, e = os.Open(job.OutputPath)
			if e == nil {
				defer file.Close()
				var info os.FileInfo
				info, e = file.Stat()
				if e == nil && (info.Size() <= 0 || info.Size() > 512<<20) {
					e = fmt.Errorf("成片文件为空或超过512MB")
				}
				if e == nil {
					key, keyErr := agentworkspace.AgentAssetObjectKey(p.OwnerUsername, p.ID, id, "mp4")
					e = keyErr
					if e == nil {
						var duration int64
						for _, track := range job.TimelineSnapshot.Tracks {
							for _, clip := range track.Clips {
								if end := clip.StartMS + clip.DurationMS; end > duration {
									duration = end
								}
							}
						}
						asset = aicut.Asset{ID: id, ProjectID: p.ID, Kind: "video", Name: "成片.mp4", ContentType: "video/mp4", ObjectKey: strings.Replace(key, "agent-assets/", "aicut-assets/", 1), ByteSize: info.Size(), DurationMS: duration, CreatedAt: time.Now().UTC()}
						e = s.Objects.Put(ctx, aicutObjectAsset(asset), file)
						if e == nil && assetErr != nil {
							e = s.Cut.CreateAsset(ctx, p.OwnerUsername, asset)
						}
						if e == nil {
							out.AssetID = id
							out.DurationMS = duration
							out.Status = "succeeded"
						}
					}
				}
			}
		}
	}
	if e != nil {
		out.Error = "成片持久保存失败：" + e.Error() + "；可重试导入原成片，不会重新调用模型"
	}
	_, saveErr := s.Media.MutateMedia(context.WithoutCancel(ctx), p.OwnerUsername, p.ID, func(v *screenplay.MediaState) error {
		for i := range v.Items {
			if v.Items[i].ID == out.ID {
				v.Items[i] = out
				return nil
			}
		}
		v.Items = append(v.Items, out)
		return nil
	})
	if saveErr != nil {
		return out, saveErr
	}
	if out.Status != "succeeded" {
		return out, fmt.Errorf("%s", out.Error)
	}
	return out, nil
}
