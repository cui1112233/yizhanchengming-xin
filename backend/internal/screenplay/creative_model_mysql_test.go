package screenplay

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"regexp"
	"testing"
	"time"
)

func TestCreativeModelConfigMySQLAtomicAspectAndFullJSON(t *testing.T) {
	db, m, e := sqlmock.New()
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	ctx := context.Background()
	c := CreativeConfig{ConfigVersion: 1, AspectRatio: "21:9", ImageResolution: "4K", AssetPromptType: "western", AssetProcessPrompt: "", AssetProcessPromptCleared: true, ImageCoreInstruction: "", ImageCoreInstructionCleared: true, VoiceDesignModel: "qwen-tts", ShowOriginalText: false, CharacterImageMode: "typedprompt:luoshui_character_design_212", CharacterDesignPromptID: "luoshui_character_design_212", SceneImageMode: "none"}
	raw, e := encodeCreativeConfig(&c)
	if e != nil {
		t.Fatal(e)
	}
	now := time.Now()
	m.ExpectBegin()
	m.ExpectQuery("SELECT revision FROM screenplay_projects").WithArgs("p", "alice").WillReturnRows(sqlmock.NewRows([]string{"revision"}).AddRow(7))
	m.ExpectExec("UPDATE screenplay_projects SET art_style=.*aspect_ratio=IF").WithArgs("", "", "", "", "", "21:9", "21:9", sqlmock.AnyArg(), "p", "alice", int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectExec("INSERT INTO screenplay_project_creative_configs").WithArgs("p", "alice", string(raw), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	m.ExpectCommit()
	columns := []string{"id", "owner_username", "title", "script_mode", "aspect_ratio", "art_style", "style_desc", "selected_style", "synopsis", "source_text", "model_id", "prompt_id", "asset_degree", "permission_mode", "revision", "created_at", "updated_at"}
	m.ExpectQuery("SELECT id, owner_username, title").WithArgs("p", "alice").WillReturnRows(sqlmock.NewRows(columns).AddRow("p", "alice", "标题", "b", "21:9", "", "", "", "", "原文", "", "", "master", "bypass", 8, now, now))
	m.ExpectQuery(regexp.QuoteMeta("SELECT config FROM screenplay_project_creative_configs WHERE project_id=? AND owner_username=?")).WithArgs("p", "alice").WillReturnRows(sqlmock.NewRows([]string{"config"}).AddRow(raw))
	saved, e := NewMySQLStore(db).SaveCreativeConfig(ctx, "alice", "p", c, 7)
	if e != nil || saved.Revision != 8 || saved.AspectRatio != "21:9" || saved.CreativeConfig.ImageResolution != "4K" || saved.CreativeConfig.VoiceDesignModel != "qwen-tts" || saved.CreativeConfig.ShowOriginalText || saved.CreativeConfig.AssetProcessPrompt != "" || !saved.CreativeConfig.AssetProcessPromptCleared || saved.CreativeConfig.CharacterImageMode != c.CharacterImageMode {
		t.Fatal(saved, e)
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
