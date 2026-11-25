package novelcreation

import (
	"context"
	"strings"
	"time"
)

// 洛水 novels.pipeline.cover_runner 的等价实现：
// 先生成封面提示词，再调用应用统一的绘图接口生成封面图。
// CM 由 HTTP 层注入绘图能力，这里只负责提示词、落库与读取，
// 因此不会把图像渠道配置复制进小说模块。
type ProjectCover struct {
	ProjectID string    `json:"projectId"`
	Title     string    `json:"title"`
	Prompt    string    `json:"prompt"`
	Caption   string    `json:"caption"`
	ImageURL  string    `json:"imageUrl"`
	ModelID   string    `json:"modelId,omitempty"`
	Source    string    `json:"promptSource"`
	CreatedAt time.Time `json:"createdAt"`
}

const coverRecordKind = "cover"

// CoverImageGenerator 由应用层注入统一的绘图能力。
// 小说模块只传提示词与账号，不复制图像渠道配置。
type CoverImageGenerator interface {
	GenerateCoverImage(ctx context.Context, owner, prompt, imageModelID string) (string, error)
}

// ConfigureCoverGenerator 注入绘图实现；未注入时封面生成返回明确错误。
func (s *Service) ConfigureCoverGenerator(generator CoverImageGenerator) {
	s.covers = generator
}

// buildCoverPrompt 按书名、题材与核心冲突拼装绘图提示词。
// 不调用模型，便于重试与审计；styleHint 可由用户覆盖默认风格。
func buildCoverPrompt(project Project, styleHint string) string {
	title := strings.TrimSpace(project.Title)
	genre := strings.TrimSpace(project.Genre)
	style := strings.TrimSpace(styleHint)
	if style == "" {
		style = "网文封面，电影海报感，构图干净，标题留白，高清"
	}
	parts := []string{}
	if title != "" {
		parts = append(parts, "书名《"+title+"》")
	}
	if genre != "" {
		parts = append(parts, "题材："+genre)
	}
	if world := strings.TrimSpace(project.World); world != "" {
		parts = append(parts, "世界设定："+world)
	}
	if conflict := strings.TrimSpace(project.CoreConflict); conflict != "" {
		parts = append(parts, "核心冲突："+conflict)
	}
	parts = append(parts, "风格："+style)
	parts = append(parts, "不含文字水印，不出现真实人物肖像")
	return strings.Join(parts, "；")
}

func coverCaption(project Project) string {
	title := strings.TrimSpace(project.Title)
	if genre := strings.TrimSpace(project.Genre); genre != "" {
		return "《" + title + "》· " + genre + "封面"
	}
	return "《" + title + "》封面"
}

// PrepareProjectCover 生成封面提示词并确认作品存在。
// 返回的提示词用于调用应用统一的绘图接口。
func (s *Service) PrepareProjectCover(ctx context.Context, owner, projectID, styleHint string) (prompt, caption string, err error) {
	if s == nil || s.store == nil {
		return "", "", &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	project, err := s.store.GetProject(ctx, strings.TrimSpace(owner), strings.TrimSpace(projectID))
	if err != nil {
		return "", "", err
	}
	return buildCoverPrompt(project, styleHint), coverCaption(project), nil
}

// SaveProjectCover 保存已生成的封面；失败不覆盖旧封面。
func (s *Service) SaveProjectCover(ctx context.Context, owner, projectID string, cover ProjectCover) (ProjectCover, error) {
	if s == nil || s.store == nil {
		return ProjectCover{}, &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	owner, projectID = strings.TrimSpace(owner), strings.TrimSpace(projectID)
	if strings.TrimSpace(cover.ImageURL) == "" {
		return ProjectCover{}, &DomainError{Code: ErrInvalidOutput, Message: "绘图接口未返回图片地址"}
	}
	cover.ProjectID = projectID
	if cover.CreatedAt.IsZero() {
		cover.CreatedAt = time.Now().UTC()
	}
	if cover.Source == "" {
		cover.Source = "template"
	}
	if err := putContent(ctx, s.store, owner, projectID, coverRecordKind, projectID, "ready", time.Time{}, cover); err != nil {
		return cover, err
	}
	return cover, nil
}

// GenerateProjectCover 生成并保存封面：先准备提示词，再调用注入的绘图接口。
// 绘图失败时不写入记录，旧封面保持不变。
func (s *Service) GenerateProjectCover(ctx context.Context, owner, projectID, imageModelID, styleHint string) (ProjectCover, error) {
	if s.covers == nil {
		return ProjectCover{}, inputError("未配置绘图接口，无法生成封面")
	}
	prompt, caption, err := s.PrepareProjectCover(ctx, owner, projectID, styleHint)
	if err != nil {
		return ProjectCover{}, err
	}
	imageURL, err := s.covers.GenerateCoverImage(ctx, strings.TrimSpace(owner), prompt, strings.TrimSpace(imageModelID))
	if err != nil {
		return ProjectCover{}, err
	}
	return s.SaveProjectCover(ctx, owner, projectID, ProjectCover{
		Title:    caption,
		Prompt:   prompt,
		Caption:  caption,
		ImageURL: imageURL,
		ModelID:  strings.TrimSpace(imageModelID),
	})
}

// GetProjectCover 读取已保存封面；没有封面时返回 found=false。
func (s *Service) GetProjectCover(ctx context.Context, owner, projectID string) (ProjectCover, bool, error) {
	if s == nil || s.store == nil {
		return ProjectCover{}, false, &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	owner, projectID = strings.TrimSpace(owner), strings.TrimSpace(projectID)
	if _, err := s.store.GetProject(ctx, owner, projectID); err != nil {
		return ProjectCover{}, false, err
	}
	var cover ProjectCover
	found, err := readContent(ctx, s.store, owner, projectID, coverRecordKind, projectID, &cover)
	if err != nil || !found {
		return ProjectCover{}, false, err
	}
	return cover, true, nil
}
