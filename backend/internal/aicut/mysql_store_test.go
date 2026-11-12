package aicut

import (
	"context"
	"database/sql"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestEncodeExportJobSnapshotsUsesTextForMySQLJSONColumns(t *testing.T) {
	options, project, timeline, err := encodeExportJobSnapshots(ExportJob{
		Options:          ExportOptions{Resolution: "1080p", FrameRate: 30, Codec: "h264", BitDepth: 8},
		ProjectSnapshot:  Project{ID: "cut-1", Title: "导出工程"},
		TimelineSnapshot: Timeline{Revision: 2, Tracks: []Track{}},
	})
	if err != nil {
		t.Fatal(err)
	}
	for name, value := range map[string]string{"options": options, "project": project, "timeline": timeline} {
		if !strings.HasPrefix(value, "{") {
			t.Fatalf("%s snapshot must be textual JSON, got %q", name, value)
		}
	}
}

func TestMySQLStoreTimelineHistoryReadsPersistentCursor(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	mock.ExpectQuery("SELECT cursor_no, max_no FROM aicut_timeline_history_state").WithArgs("cut-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"cursor_no", "max_no"}).AddRow(1, 2))
	history, err := store.TimelineHistory(context.Background(), "alice", "cut-1")
	if err != nil {
		t.Fatal(err)
	}
	if !history.CanUndo || !history.CanRedo {
		t.Fatalf("history=%#v", history)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreTimelineHistoryTreatsExistingProjectWithoutHistoryAsBoundary(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	mock.ExpectQuery("SELECT cursor_no, max_no FROM aicut_timeline_history_state").WithArgs("cut-1", "alice").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT 1 FROM aicut_projects").WithArgs("cut-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	history, err := store.TimelineHistory(context.Background(), "alice", "cut-1")
	if err != nil {
		t.Fatal(err)
	}
	if history.CanUndo || history.CanRedo {
		t.Fatalf("history=%#v", history)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreSaveTimelineReplacesTracksAtomically(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 11, 9, 16, 0, 0, 0, time.UTC)
	timeline := Timeline{Script: "第一场", Tracks: []Track{{ID: "track-1", Kind: TrackVideo, Name: "主画面", Clips: []Clip{{ID: "clip-1", AssetID: "asset-1", Kind: ClipVideo, Name: "开场", DurationMS: 2000, SourceDurationMS: 3000, Transition: "flash-black", TransitionMS: 500, ReviewStatus: "keep", ReviewReason: "可用", LinkGroupID: "link-1", CompoundID: "compound-1", CompoundName: "开场复合片段", Volume: 0.25, FadeInMS: 300, FadeOutMS: 500}}}}}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT revision FROM aicut_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL FOR UPDATE`)).WithArgs("cut-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"revision"}).AddRow(0))
	mock.ExpectQuery("SELECT cursor_no, max_no FROM aicut_timeline_history_state").WithArgs("cut-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"cursor_no", "max_no"}).AddRow(0, 0))
	mock.ExpectExec("DELETE FROM aicut_timeline_history").WithArgs("cut-1", "alice", int64(0)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO aicut_timeline_history").WithArgs("cut-1", "alice", int64(1), sqlmock.AnyArg(), now).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE aicut_timeline_history_state SET cursor_no").WithArgs(int64(1), int64(1), now, "cut-1", "alice").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE aicut_projects SET revision = ?, script_text = ?, markers_json = ?, updated_at = ? WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`)).WithArgs(int64(1), "第一场", "null", now, "cut-1", "alice").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM aicut_tracks WHERE project_id = ? AND owner_username = ?`)).WithArgs("cut-1", "alice").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO aicut_tracks").WithArgs("track-1", "cut-1", "alice", TrackVideo, "主画面", 0, false, false, `{}`, now, now).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO aicut_clips").WithArgs("clip-1", "track-1", "cut-1", "alice", "asset-1", ClipVideo, "开场", int64(0), int64(2000), int64(0), int64(3000), "", "flash-black", int64(500), "keep", "可用", "link-1", float64(1), 0.25, int64(300), int64(500), nil, `{"compoundId":"compound-1","compoundName":"开场复合片段"}`, now, now).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	saved, err := store.SaveTimeline(context.Background(), "alice", "cut-1", timeline, 0, now)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Revision != 1 {
		t.Fatalf("revision=%d", saved.Revision)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreClaimsOldestQueuedASRJob(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 11, 9, 18, 0, 0, 0, time.UTC)
	created := now.Add(-time.Minute)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, owner_username, project_id, asset_id, model_id, language, expected_revision, state, progress").WillReturnRows(sqlmock.NewRows([]string{"id", "owner_username", "project_id", "asset_id", "model_id", "language", "expected_revision", "state", "progress", "error_message", "imported_count", "track_id", "created_at", "updated_at", "timeline_id"}).AddRow("asr-1", "alice", "cut-1", "asset-1", "asr-main", "zh", 2, ASRQueued, 0, "", 0, "", created, created, ""))
	mock.ExpectExec("UPDATE aicut_asr_jobs SET state = \\?, progress = \\?, updated_at = \\?").WithArgs(ASRRunning, 5, now, "asr-1", "cut-1", "alice", ASRQueued).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	job, err := store.ClaimNextASRJob(context.Background(), now)
	if err != nil {
		t.Fatal(err)
	}
	if job.ID != "asr-1" || job.State != ASRRunning || job.Progress != 5 {
		t.Fatalf("job=%#v", job)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreClaimsOldestQueuedExportJobWithSnapshots(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 11, 11, 14, 0, 0, 0, time.UTC)
	created := now.Add(-time.Minute)
	columns := []string{"id", "owner_username", "project_id", "format", "expected_revision", "state", "progress", "error_message", "output_path", "options_json", "project_snapshot_json", "timeline_snapshot_json", "created_at", "updated_at"}
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, owner_username, project_id, format, expected_revision, state, progress").WithArgs(ExportQueued).WillReturnRows(sqlmock.NewRows(columns).AddRow("export-1", "alice", "cut-1", "mp4", 3, ExportQueued, 0, "", "", `{"resolution":"1080p","frameRate":30,"codec":"h264","bitDepth":8}`, `{"id":"cut-1","title":"队列工程","aspectRatio":"16:9"}`, `{"revision":3,"tracks":[]}`, created, created))
	mock.ExpectExec("UPDATE aicut_export_jobs SET state = \\?, progress = \\?, updated_at = \\?").WithArgs(ExportRunning, 5, now, "export-1", "cut-1", "alice", ExportQueued).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	job, err := store.ClaimNextExportJob(context.Background(), now)
	if err != nil {
		t.Fatal(err)
	}
	if job.ID != "export-1" || job.State != ExportRunning || job.TimelineSnapshot.Revision != 3 || job.Options.Resolution != "1080p" {
		t.Fatalf("job=%#v", job)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestClipEditorJSONAudioProcessingRoundTrip(t *testing.T) {
	clip := Clip{AudioProcessing: &AudioProcessing{Denoise: .7, EQLowDB: 2, EQMidDB: -3, EQHighDB: 4, ChannelMode: "swap", VoiceEffect: "telephone", PitchSemitones: 5.5}}
	encoded, err := encodeClipEditorJSON(clip)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Clip
	if err := decodeClipEditorJSON(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.AudioProcessing == nil || decoded.AudioProcessing.Denoise != .7 || decoded.AudioProcessing.ChannelMode != "swap" || decoded.AudioProcessing.EQMidDB != -3 || decoded.AudioProcessing.VoiceEffect != "telephone" || decoded.AudioProcessing.PitchSemitones != 5.5 {
		t.Fatalf("decoded audio processing = %#v; json=%s", decoded.AudioProcessing, encoded)
	}
}

func TestClipEditorJSONVisualProcessingRoundTrip(t *testing.T) {
	freezeSourceMS := int64(125)
	clip := Clip{VisualProcessing: &VisualProcessing{
		FrameInterpolation: "blend",
		Stabilization:      VideoStabilization{Enabled: true, Smoothness: 40, Zoom: 8, Tripod: true},
		Color:              ColorAdjustment{Exposure: .6, Temperature: -.2, Tint: .1, Highlights: .3, Shadows: -.4, Gamma: 1.2, Curve: []float64{0, .2, .55, .85, 1}, CurvePoints: []ColorCurvePoint{{X: 0, Y: 0}, {X: .37, Y: .82}, {X: 1, Y: 1}}},
		ChromaKey:          ChromaKey{Enabled: true, RemoveBlack: true, Color: "#00ff00", Similarity: .25, Blend: .08},
		Mask:               VisualMask{Type: "bezier", X: .4, Y: .6, Width: .8, Height: .7, Feather: .12, Invert: true, BezierPoints: []VisualBezierPoint{{X: .1, Y: .1, InX: .05, InY: .2, OutX: .2, OutY: .05}, {X: .9, Y: .2, InX: .8, InY: .1, OutX: .95, OutY: .3}, {X: .7, Y: .9, InX: .8, InY: .8, OutX: .6, OutY: .95}}},
		Geometry:           VisualGeometry{CropLeft: .1, CropTop: .05, CropRight: .2, CropBottom: .15, FlipHorizontal: true, FlipVertical: true},
	}, FreezeSourceMS: &freezeSourceMS, BlendMode: "screen"}
	encoded, err := encodeClipEditorJSON(clip)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Clip
	if err := decodeClipEditorJSON(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	got := decoded.VisualProcessing
	if got == nil || got.FrameInterpolation != "blend" || !got.Stabilization.Enabled || got.Stabilization.Smoothness != 40 || got.Stabilization.Zoom != 8 || !got.Stabilization.Tripod || got.Color.Exposure != .6 || got.Color.Gamma != 1.2 || len(got.Color.Curve) != 5 || got.Color.Curve[2] != .55 || len(got.Color.CurvePoints) != 3 || got.Color.CurvePoints[1].X != .37 || got.Color.CurvePoints[1].Y != .82 || !got.ChromaKey.Enabled || !got.ChromaKey.RemoveBlack || got.ChromaKey.Color != "#00ff00" || got.Mask.Type != "bezier" || len(got.Mask.BezierPoints) != 3 || got.Mask.BezierPoints[1].X != .9 || got.Mask.Feather != .12 || !got.Mask.Invert || got.Geometry.CropLeft != .1 || got.Geometry.CropBottom != .15 || !got.Geometry.FlipHorizontal || !got.Geometry.FlipVertical || decoded.FreezeSourceMS == nil || *decoded.FreezeSourceMS != 125 || decoded.BlendMode != "screen" {
		t.Fatalf("decoded visual processing = %#v; json=%s", got, encoded)
	}
}
