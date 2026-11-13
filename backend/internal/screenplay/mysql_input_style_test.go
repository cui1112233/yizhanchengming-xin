package screenplay

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"strings"
	"testing"
)

func TestMySQLCreateProjectPersistsSourceInSameInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectExec(`INSERT INTO screenplay_projects\s+\(id, owner_username, title, script_mode, aspect_ratio, art_style, style_desc, selected_style, synopsis, source_text, model_id, prompt_id, asset_degree, permission_mode, revision, created_at, updated_at\)`).WithArgs("p", "alice", "标题", "", "", "名称", "后端画风", "style-id", "", "完整原文", "", "", "", string(PermissionAuto), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	project, err := NewMySQLStore(db).CreateProject(context.Background(), Project{ID: "p", OwnerUsername: "alice", Title: "标题", ArtStyle: "名称", StyleDesc: "后端画风", SelectedStyle: "style-id", SourceText: "完整原文"})
	if err != nil || project.SourceText != "完整原文" {
		t.Fatalf("create=%+v %v", project, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(strings.TrimSpace(err.Error()))
	}
}
