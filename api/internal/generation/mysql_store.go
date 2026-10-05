package generation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

import "github.com/cui1112233/yizhanchengming-xin/api/internal/intake"

type MySQLStore struct{ db *sql.DB }

func NewMySQLStore(db *sql.DB) *MySQLStore { return &MySQLStore{db: db} }

func noRows(err error) error {
	if errors.Is(err, sql.ErrNoRows) { return ErrNotFound }
	return err
}

func (s *MySQLStore) GetBookForProject(ctx context.Context, projectID, bookID int64) (intake.Book, error) {
	var b intake.Book
	err := s.db.QueryRowContext(ctx, `SELECT b.id,b.intake_id,b.source,b.platform_id,b.external_book_id,b.title,b.body_ref,b.original_text,b.category,b.genre,b.gender,b.gender_source,b.style,b.status,b.error_message,b.created_at,b.updated_at FROM books b JOIN batch_projects p ON p.intake_id=b.intake_id WHERE p.id=? AND b.id=?`, projectID, bookID).Scan(
		&b.ID,&b.IntakeID,&b.Source,&b.PlatformID,&b.ExternalBookID,&b.Title,&b.BodyRef,&b.OriginalText,&b.Category,&b.Genre,&b.Gender,&b.GenderSource,&b.Style,&b.Status,&b.ErrorMessage,&b.CreatedAt,&b.UpdatedAt,
	)
	if err != nil { return intake.Book{}, noRows(err) }
	return b,nil
}

func (s *MySQLStore) ListBooksForProject(ctx context.Context, projectID int64) ([]intake.Book, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT b.id,b.intake_id,b.source,b.platform_id,b.external_book_id,b.title,b.body_ref,b.original_text,b.category,b.genre,b.gender,b.gender_source,b.style,b.status,b.error_message,b.created_at,b.updated_at FROM books b JOIN batch_projects p ON p.intake_id=b.intake_id WHERE p.id=? ORDER BY b.id`, projectID)
	if err != nil { return nil,err }
	defer rows.Close()
	out:=[]intake.Book{}
	for rows.Next(){ var b intake.Book; if err:=rows.Scan(&b.ID,&b.IntakeID,&b.Source,&b.PlatformID,&b.ExternalBookID,&b.Title,&b.BodyRef,&b.OriginalText,&b.Category,&b.Genre,&b.Gender,&b.GenderSource,&b.Style,&b.Status,&b.ErrorMessage,&b.CreatedAt,&b.UpdatedAt); err!=nil{return nil,err}; out=append(out,b)}
	return out,rows.Err()
}

func (s *MySQLStore) ResolvePrompt(ctx context.Context, key string) (Prompt,error){
	var p Prompt
	err:=s.db.QueryRowContext(ctx,`SELECT id,prompt_key,version,content,enabled,created_at,updated_at FROM generation_prompts WHERE prompt_key=? AND enabled=1 ORDER BY version DESC LIMIT 1`,key).Scan(&p.ID,&p.Key,&p.Version,&p.Content,&p.Enabled,&p.CreatedAt,&p.UpdatedAt)
	if err!=nil{return Prompt{},noRows(err)}
	return p,nil
}

func (s *MySQLStore) ListPrompts(ctx context.Context)([]Prompt,error){
	rows,err:=s.db.QueryContext(ctx,`SELECT id,prompt_key,version,content,enabled,created_at,updated_at FROM generation_prompts ORDER BY prompt_key,version DESC`); if err!=nil{return nil,err}; defer rows.Close()
	out:=[]Prompt{}; for rows.Next(){var p Prompt;if err:=rows.Scan(&p.ID,&p.Key,&p.Version,&p.Content,&p.Enabled,&p.CreatedAt,&p.UpdatedAt);err!=nil{return nil,err};out=append(out,p)};return out,rows.Err()
}

func (s *MySQLStore) CreateBookRun(ctx context.Context,v BookRun)(BookRun,error){
	r,err:=s.db.ExecContext(ctx,`INSERT INTO book_runs(batch_project_id,book_id,status,request_id,error_message,started_at,finished_at) VALUES(?,?,?,?,?,?,?)`,v.BatchProjectID,v.BookID,v.Status,v.RequestID,v.ErrorMessage,v.StartedAt,v.FinishedAt);if err!=nil{return BookRun{},fmt.Errorf("create book run: %w",err)};v.ID,_=r.LastInsertId();return s.bookRunByID(ctx,v.ID)
}

func (s *MySQLStore) UpdateBookRun(ctx context.Context,v BookRun)(BookRun,error){
	r,err:=s.db.ExecContext(ctx,`UPDATE book_runs SET status=?,request_id=?,error_message=?,started_at=?,finished_at=? WHERE id=?`,v.Status,v.RequestID,v.ErrorMessage,v.StartedAt,v.FinishedAt,v.ID);if err!=nil{return BookRun{},err};if n,_:=r.RowsAffected();n==0{return BookRun{},ErrNotFound};return s.bookRunByID(ctx,v.ID)
}

func (s *MySQLStore) bookRunByID(ctx context.Context,id int64)(BookRun,error){
	var v BookRun;err:=s.db.QueryRowContext(ctx,`SELECT id,batch_project_id,book_id,status,request_id,error_message,started_at,finished_at,created_at,updated_at FROM book_runs WHERE id=?`,id).Scan(&v.ID,&v.BatchProjectID,&v.BookID,&v.Status,&v.RequestID,&v.ErrorMessage,&v.StartedAt,&v.FinishedAt,&v.CreatedAt,&v.UpdatedAt);if err!=nil{return BookRun{},noRows(err)};return v,nil
}

func (s *MySQLStore) LatestBookRun(ctx context.Context,projectID,bookID int64)(BookRun,error){
	var v BookRun;err:=s.db.QueryRowContext(ctx,`SELECT id,batch_project_id,book_id,status,request_id,error_message,started_at,finished_at,created_at,updated_at FROM book_runs WHERE batch_project_id=? AND book_id=? ORDER BY id DESC LIMIT 1`,projectID,bookID).Scan(&v.ID,&v.BatchProjectID,&v.BookID,&v.Status,&v.RequestID,&v.ErrorMessage,&v.StartedAt,&v.FinishedAt,&v.CreatedAt,&v.UpdatedAt);if err!=nil{return BookRun{},noRows(err)};return v,nil
}

func (s *MySQLStore) ListBookRunsByProject(ctx context.Context,projectID int64)([]BookRun,error){
	rows,err:=s.db.QueryContext(ctx,`SELECT id,batch_project_id,book_id,status,request_id,error_message,started_at,finished_at,created_at,updated_at FROM book_runs WHERE batch_project_id=? ORDER BY id`,projectID);if err!=nil{return nil,err};defer rows.Close();out:=[]BookRun{};for rows.Next(){var v BookRun;if err:=rows.Scan(&v.ID,&v.BatchProjectID,&v.BookID,&v.Status,&v.RequestID,&v.ErrorMessage,&v.StartedAt,&v.FinishedAt,&v.CreatedAt,&v.UpdatedAt);err!=nil{return nil,err};out=append(out,v)};return out,rows.Err()
}

func (s *MySQLStore) CreateStageRun(ctx context.Context,v StageRun)(StageRun,error){
	r,err:=s.db.ExecContext(ctx,`INSERT INTO stage_runs(book_run_id,book_id,stage,status,attempt,request_id,prompt_key,prompt_version,input_snapshot,output_text,error_message,validation_result,started_at,finished_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,v.BookRunID,v.BookID,v.Stage,v.Status,v.Attempt,v.RequestID,v.PromptKey,v.PromptVersion,v.InputSnapshot,v.OutputText,v.ErrorMessage,v.ValidationResult,v.StartedAt,v.FinishedAt);if err!=nil{return StageRun{},fmt.Errorf("create stage run: %w",err)};v.ID,_=r.LastInsertId();return s.stageRunByID(ctx,v.ID)
}

func (s *MySQLStore) UpdateStageRun(ctx context.Context,v StageRun)(StageRun,error){
	r,err:=s.db.ExecContext(ctx,`UPDATE stage_runs SET status=?,request_id=?,prompt_key=?,prompt_version=?,input_snapshot=?,output_text=?,error_message=?,validation_result=?,started_at=?,finished_at=? WHERE id=?`,v.Status,v.RequestID,v.PromptKey,v.PromptVersion,v.InputSnapshot,v.OutputText,v.ErrorMessage,v.ValidationResult,v.StartedAt,v.FinishedAt,v.ID);if err!=nil{return StageRun{},err};if n,_:=r.RowsAffected();n==0{return StageRun{},ErrNotFound};return s.stageRunByID(ctx,v.ID)
}

func (s *MySQLStore) stageRunByID(ctx context.Context,id int64)(StageRun,error){
	var v StageRun;err:=s.db.QueryRowContext(ctx,`SELECT id,book_run_id,book_id,stage,status,attempt,request_id,prompt_key,prompt_version,input_snapshot,output_text,error_message,validation_result,started_at,finished_at,created_at,updated_at FROM stage_runs WHERE id=?`,id).Scan(&v.ID,&v.BookRunID,&v.BookID,&v.Stage,&v.Status,&v.Attempt,&v.RequestID,&v.PromptKey,&v.PromptVersion,&v.InputSnapshot,&v.OutputText,&v.ErrorMessage,&v.ValidationResult,&v.StartedAt,&v.FinishedAt,&v.CreatedAt,&v.UpdatedAt);if err!=nil{return StageRun{},noRows(err)};return v,nil
}

func (s *MySQLStore) ListStageRuns(ctx context.Context,bookRunID int64)([]StageRun,error){
	rows,err:=s.db.QueryContext(ctx,`SELECT id,book_run_id,book_id,stage,status,attempt,request_id,prompt_key,prompt_version,input_snapshot,output_text,error_message,validation_result,started_at,finished_at,created_at,updated_at FROM stage_runs WHERE book_run_id=? ORDER BY id`,bookRunID);if err!=nil{return nil,err};defer rows.Close();out:=[]StageRun{};for rows.Next(){var v StageRun;if err:=rows.Scan(&v.ID,&v.BookRunID,&v.BookID,&v.Stage,&v.Status,&v.Attempt,&v.RequestID,&v.PromptKey,&v.PromptVersion,&v.InputSnapshot,&v.OutputText,&v.ErrorMessage,&v.ValidationResult,&v.StartedAt,&v.FinishedAt,&v.CreatedAt,&v.UpdatedAt);err!=nil{return nil,err};out=append(out,v)};return out,rows.Err()
}

func (s *MySQLStore) LatestStageRun(ctx context.Context,bookRunID int64,stage Stage)(StageRun,error){
	var v StageRun;err:=s.db.QueryRowContext(ctx,`SELECT id,book_run_id,book_id,stage,status,attempt,request_id,prompt_key,prompt_version,input_snapshot,output_text,error_message,validation_result,started_at,finished_at,created_at,updated_at FROM stage_runs WHERE book_run_id=? AND stage=? ORDER BY attempt DESC,id DESC LIMIT 1`,bookRunID,stage).Scan(&v.ID,&v.BookRunID,&v.BookID,&v.Stage,&v.Status,&v.Attempt,&v.RequestID,&v.PromptKey,&v.PromptVersion,&v.InputSnapshot,&v.OutputText,&v.ErrorMessage,&v.ValidationResult,&v.StartedAt,&v.FinishedAt,&v.CreatedAt,&v.UpdatedAt);if err!=nil{return StageRun{},noRows(err)};return v,nil
}

func (s *MySQLStore) AudioMeasurementByAsset(ctx context.Context,projectID,bookID int64,assetHash string)(AudioMeasurement,error){
	var v AudioMeasurement
	err:=s.db.QueryRowContext(ctx,`SELECT id,batch_project_id,book_id,audio_asset,asset_hash,duration_ms,measured_at,created_at FROM audio_measurements WHERE batch_project_id=? AND book_id=? AND asset_hash=? LIMIT 1`,projectID,bookID,assetHash).Scan(&v.ID,&v.BatchProjectID,&v.BookID,&v.AudioAsset,&v.AssetHash,&v.DurationMS,&v.MeasuredAt,&v.CreatedAt)
	if err!=nil{return AudioMeasurement{},noRows(err)}
	return v,nil
}

func (s *MySQLStore) LatestAudioMeasurement(ctx context.Context,projectID,bookID int64)(AudioMeasurement,error){
	var v AudioMeasurement
	err:=s.db.QueryRowContext(ctx,`SELECT id,batch_project_id,book_id,audio_asset,asset_hash,duration_ms,measured_at,created_at FROM audio_measurements WHERE batch_project_id=? AND book_id=? ORDER BY measured_at DESC,id DESC LIMIT 1`,projectID,bookID).Scan(&v.ID,&v.BatchProjectID,&v.BookID,&v.AudioAsset,&v.AssetHash,&v.DurationMS,&v.MeasuredAt,&v.CreatedAt)
	if err!=nil{return AudioMeasurement{},noRows(err)}
	return v,nil
}

func (s *MySQLStore) CreateAudioMeasurement(ctx context.Context,v AudioMeasurement)(AudioMeasurement,error){
	r,err:=s.db.ExecContext(ctx,`INSERT INTO audio_measurements(batch_project_id,book_id,audio_asset,asset_hash,duration_ms,measured_at) VALUES(?,?,?,?,?,?)`,v.BatchProjectID,v.BookID,v.AudioAsset,v.AssetHash,v.DurationMS,v.MeasuredAt)
	if err!=nil{return AudioMeasurement{},fmt.Errorf("create audio measurement: %w",err)}
	v.ID,_=r.LastInsertId()
	err=s.db.QueryRowContext(ctx,`SELECT id,batch_project_id,book_id,audio_asset,asset_hash,duration_ms,measured_at,created_at FROM audio_measurements WHERE id=?`,v.ID).Scan(&v.ID,&v.BatchProjectID,&v.BookID,&v.AudioAsset,&v.AssetHash,&v.DurationMS,&v.MeasuredAt,&v.CreatedAt)
	if err!=nil{return AudioMeasurement{},err}
	return v,nil
}
