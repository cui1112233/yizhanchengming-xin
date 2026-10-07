package generation

import (
	"context"
	"database/sql"
	"fmt"
)

func (s *MySQLStore) GetStoryboard(ctx context.Context, p, b int64) (StoryboardDocument, error) {
	var d StoryboardDocument
	var source sql.NullInt64
	err := s.db.QueryRowContext(ctx, `SELECT batch_project_id,book_id,version,source_stage_run_id FROM script_storyboard_documents WHERE batch_project_id=? AND book_id=?`, p, b).Scan(&d.BatchProjectID, &d.BookID, &d.Version, &source)
	if err != nil {
		return StoryboardDocument{}, noRows(err)
	}
	if source.Valid {
		d.SourceStageRunID = source.Int64
	}
	cards, err := s.storyboardCards(ctx, p, b)
	if err != nil {
		return StoryboardDocument{}, err
	}
	d.Cards = cards
	return d, nil
}
func (s *MySQLStore) storyboardCards(ctx context.Context, p, b int64) ([]StoryboardCard, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,position,title,content,version FROM script_storyboard_cards WHERE batch_project_id=? AND book_id=? ORDER BY position,id`, p, b)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []StoryboardCard{}
	for rows.Next() {
		var c StoryboardCard
		if err := rows.Scan(&c.ID, &c.Position, &c.Title, &c.Content, &c.Version); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
func (s *MySQLStore) CreateStoryboard(ctx context.Context, d StoryboardDocument) (StoryboardDocument, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return StoryboardDocument{}, err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `INSERT INTO script_storyboard_documents(batch_project_id,book_id,version,source_stage_run_id) VALUES(?,?,?,?)`, d.BatchProjectID, d.BookID, d.Version, d.SourceStageRunID)
	if err != nil {
		return StoryboardDocument{}, err
	}
	for i, c := range d.Cards {
		if c.Title == "" {
			c.Title = fmt.Sprintf("分镜%d", i+1)
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO script_storyboard_cards(batch_project_id,book_id,position,title,content,version) VALUES(?,?,?,?,?,1)`, d.BatchProjectID, d.BookID, i+1, c.Title, c.Content); err != nil {
			return StoryboardDocument{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return StoryboardDocument{}, err
	}
	return s.GetStoryboard(ctx, d.BatchProjectID, d.BookID)
}
func (s *MySQLStore) mutateStoryboard(ctx context.Context, p, b int64, expected int, mutate func(*sql.Tx) error) (StoryboardDocument, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return StoryboardDocument{}, err
	}
	defer tx.Rollback()
	r, err := tx.ExecContext(ctx, `UPDATE script_storyboard_documents SET version=version+1 WHERE batch_project_id=? AND book_id=? AND version=?`, p, b, expected)
	if err != nil {
		return StoryboardDocument{}, err
	}
	n, _ := r.RowsAffected()
	if n != 1 {
		return StoryboardDocument{}, ErrConflict
	}
	if err = mutate(tx); err != nil {
		return StoryboardDocument{}, err
	}
	if err = tx.Commit(); err != nil {
		return StoryboardDocument{}, err
	}
	return s.GetStoryboard(ctx, p, b)
}
func (s *MySQLStore) SaveStoryboardCard(ctx context.Context, p, b int64, c StoryboardCard, expected int) (StoryboardDocument, error) {
	return s.mutateStoryboard(ctx, p, b, expected, func(tx *sql.Tx) error {
		if c.ID == 0 {
			var pos int
			if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(position),0)+1 FROM script_storyboard_cards WHERE batch_project_id=? AND book_id=?`, p, b).Scan(&pos); err != nil {
				return err
			}
			if c.Title == "" {
				c.Title = fmt.Sprintf("分镜%d", pos)
			}
			_, err := tx.ExecContext(ctx, `INSERT INTO script_storyboard_cards(batch_project_id,book_id,position,title,content,version) VALUES(?,?,?,?,?,1)`, p, b, pos, c.Title, c.Content)
			return err
		}
		r, err := tx.ExecContext(ctx, `UPDATE script_storyboard_cards SET title=?,content=?,version=version+1 WHERE id=? AND batch_project_id=? AND book_id=? AND version=?`, c.Title, c.Content, c.ID, p, b, c.Version)
		if err != nil {
			return err
		}
		n, _ := r.RowsAffected()
		if n != 1 {
			return ErrConflict
		}
		return nil
	})
}
func (s *MySQLStore) DeleteStoryboardCard(ctx context.Context, p, b, id int64, expected int) (StoryboardDocument, error) {
	return s.mutateStoryboard(ctx, p, b, expected, func(tx *sql.Tx) error {
		r, err := tx.ExecContext(ctx, `DELETE FROM script_storyboard_cards WHERE id=? AND batch_project_id=? AND book_id=?`, id, p, b)
		if err != nil {
			return err
		}
		n, _ := r.RowsAffected()
		if n != 1 {
			return ErrNotFound
		}
		if _, err = tx.ExecContext(ctx, `UPDATE script_storyboard_cards SET position=-position WHERE batch_project_id=? AND book_id=?`, p, b); err != nil {
			return err
		}
		rows, err := tx.QueryContext(ctx, `SELECT id FROM script_storyboard_cards WHERE batch_project_id=? AND book_id=? ORDER BY position DESC,id`, p, b)
		if err != nil {
			return err
		}
		defer rows.Close()
		i := 1
		for rows.Next() {
			var cardID int64
			if err := rows.Scan(&cardID); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `UPDATE script_storyboard_cards SET position=? WHERE id=?`, i, cardID); err != nil {
				return err
			}
			i++
		}
		return rows.Err()
	})
}
func (s *MySQLStore) ReorderStoryboard(ctx context.Context, p, b int64, ids []int64, expected int) (StoryboardDocument, error) {
	return s.mutateStoryboard(ctx, p, b, expected, func(tx *sql.Tx) error {
		var total int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM script_storyboard_cards WHERE batch_project_id=? AND book_id=?`, p, b).Scan(&total); err != nil {
			return err
		}
		if total != len(ids) {
			return ErrInvalid
		}
		for i, id := range ids {
			r, err := tx.ExecContext(ctx, `UPDATE script_storyboard_cards SET position=?,version=version+1 WHERE id=? AND batch_project_id=? AND book_id=?`, -(i + 1), id, p, b)
			if err != nil {
				return err
			}
			n, _ := r.RowsAffected()
			if n != 1 {
				return ErrInvalid
			}
		}
		for i, id := range ids {
			if _, err := tx.ExecContext(ctx, `UPDATE script_storyboard_cards SET position=? WHERE id=?`, i+1, id); err != nil {
				return err
			}
		}
		return nil
	})
}

var _ StoryboardStore = (*MySQLStore)(nil)
