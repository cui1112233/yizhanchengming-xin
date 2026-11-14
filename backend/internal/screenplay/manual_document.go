package screenplay

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// SaveManualDocument 把用户手工修改保存成不可变的新版本，同时复制当前版本的
// 场次、人物和场景，避免“改了正文，结构化数据全丢了”。
func SaveManualDocument(ctx context.Context, store Store, owner, projectID, content string) (Document, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return Document{}, fmt.Errorf("%w: 剧本内容不能为空", ErrEmptyContent)
	}
	active, err := store.ActiveDocument(ctx, owner, projectID)
	if errors.Is(err, ErrNoActiveScript) {
		if _, err = store.GetProject(ctx, owner, projectID); err != nil {
			return Document{}, err
		}
		parsed := ParseScreenplayMarkdown(content)
		documentID := newScreenplayID("doc")
		for i := range parsed.Shots {
			parsed.Shots[i].ID = newScreenplayID("shot")
			parsed.Shots[i].DocumentID, parsed.Shots[i].ProjectID, parsed.Shots[i].OwnerUsername = documentID, projectID, owner
		}
		for i := range parsed.Characters {
			parsed.Characters[i].ID = newScreenplayID("character")
			parsed.Characters[i].DocumentID, parsed.Characters[i].ProjectID, parsed.Characters[i].OwnerUsername = documentID, projectID, owner
		}
		for i := range parsed.Scenes {
			parsed.Scenes[i].ID = newScreenplayID("scene")
			parsed.Scenes[i].DocumentID, parsed.Scenes[i].ProjectID, parsed.Scenes[i].OwnerUsername = documentID, projectID, owner
		}
		return store.SaveScreenplay(ctx, Document{ID: documentID, ProjectID: projectID, OwnerUsername: owner, Source: "manual", Content: content}, parsed.Shots, parsed.Characters, parsed.Scenes)
	}
	if err != nil {
		return Document{}, err
	}
	return cloneDocumentVersion(ctx, store, owner, projectID, active, content, "manual")
}

// RestoreDocument 不覆盖旧记录，而是把指定历史版本复制成一个新的活动版本。
// 因此重复恢复也能在版本历史中追踪，不会悄悄改写旧数据。
func RestoreDocument(ctx context.Context, store Store, owner, projectID, documentID string) (Document, error) {
	documentID = strings.TrimSpace(documentID)
	if documentID == "" {
		return Document{}, fmt.Errorf("%w: 剧本版本不能为空", ErrInvalidArgument)
	}
	source, err := store.GetDocument(ctx, owner, projectID, documentID)
	if err != nil {
		return Document{}, err
	}
	return cloneDocumentVersion(ctx, store, owner, projectID, source, source.Content, "restore")
}

func cloneDocumentVersion(ctx context.Context, store Store, owner, projectID string, source Document, content, sourceKind string) (Document, error) {
	shots, err := store.ListShots(ctx, owner, projectID, source.ID)
	if err != nil {
		return Document{}, err
	}
	characters, err := store.ListCharacters(ctx, owner, projectID, source.ID)
	if err != nil {
		return Document{}, err
	}
	scenes, err := store.ListScenes(ctx, owner, projectID, source.ID)
	if err != nil {
		return Document{}, err
	}

	documentID := newScreenplayID("doc")
	for index := range shots {
		shots[index].ID = newScreenplayID("shot")
		shots[index].DocumentID = documentID
		shots[index].ProjectID = projectID
		shots[index].OwnerUsername = owner
	}
	for index := range characters {
		characters[index].ID = newScreenplayID("character")
		characters[index].DocumentID = documentID
		characters[index].ProjectID = projectID
		characters[index].OwnerUsername = owner
	}
	for index := range scenes {
		scenes[index].ID = newScreenplayID("scene")
		scenes[index].DocumentID = documentID
		scenes[index].ProjectID = projectID
		scenes[index].OwnerUsername = owner
	}

	return store.SaveScreenplay(ctx, Document{
		ID: documentID, ProjectID: projectID, OwnerUsername: owner, Source: sourceKind,
		Content: content, ChapterCount: source.ChapterCount, SegmentCount: source.SegmentCount,
	}, shots, characters, scenes)
}
