package aicut

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"time"
)

// TimelineCollectionSchemaStatements is registered by the application's normal
// migration ledger. It never rewrites legacy tracks, clips, or history.
func TimelineCollectionSchemaStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS aicut_timelines (
 id VARCHAR(64) NOT NULL,
 project_id VARCHAR(64) NOT NULL,
 owner_username VARCHAR(191) NOT NULL,
 title VARCHAR(400) NOT NULL,
 sort_index INT NOT NULL,
 revision BIGINT NOT NULL DEFAULT 0,
 cursor_no BIGINT NOT NULL DEFAULT 0,
 timeline_json JSON NOT NULL,
 created_at DATETIME(6) NOT NULL,
 updated_at DATETIME(6) NOT NULL,
 PRIMARY KEY (project_id, owner_username, id),
 CONSTRAINT fk_aicut_timelines_project FOREIGN KEY (project_id, owner_username) REFERENCES aicut_projects(id, owner_username) ON DELETE CASCADE
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS aicut_named_timeline_history (
 project_id VARCHAR(64) NOT NULL,
 owner_username VARCHAR(191) NOT NULL,
 timeline_id VARCHAR(64) NOT NULL,
 sequence_no BIGINT NOT NULL,
 timeline_json JSON NOT NULL,
 PRIMARY KEY (project_id, owner_username, timeline_id, sequence_no),
 CONSTRAINT fk_aicut_named_history_timeline FOREIGN KEY (project_id, owner_username, timeline_id) REFERENCES aicut_timelines(project_id, owner_username, id) ON DELETE CASCADE
) ENGINE=InnoDB`}
}

type mysqlCollectionRepository struct{ db *sql.DB }

func NewMySQLTimelineCollectionStore(db *sql.DB, base Store) *TimelineCollectionStore {
	return &TimelineCollectionStore{base: base, repo: &mysqlCollectionRepository{db: db}}
}

func (r *mysqlCollectionRepository) access(ctx context.Context, o, p, target string, fn func(*[]timelineCollectionRecord) error) error {
	tx, e := r.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	// Collection mutations lock the parent to serialize count and ordering. Edits
	// lock only their selected timeline, so independent timelines can save at once.
	parentQuery := `SELECT id FROM aicut_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`
	if target == "" {
		parentQuery += " FOR UPDATE"
	}
	var parent string
	if e = tx.QueryRowContext(ctx, parentQuery, p, o).Scan(&parent); errors.Is(e, sql.ErrNoRows) {
		return ErrProjectNotFound
	} else if e != nil {
		return e
	}
	rows, e := tx.QueryContext(ctx, `SELECT id, title, sort_index, created_at, updated_at FROM aicut_timelines WHERE project_id = ? AND owner_username = ? ORDER BY sort_index, id`, p, o)
	if e != nil {
		return e
	}
	items := []timelineCollectionRecord{}
	for rows.Next() {
		var item timelineCollectionRecord
		if e = rows.Scan(&item.ID, &item.Title, &item.Order, &item.CreatedAt, &item.UpdatedAt); e != nil {
			rows.Close()
			return e
		}
		item.Default = item.ID == DefaultTimelineID
		items = append(items, item)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return e
	}
	before := map[string]timelineCollectionRecord{}
	for _, item := range items {
		before[item.ID] = item
	}
	if collectionIndex(items, DefaultTimelineID) < 0 {
		d := defaultCollectionRecord()
		d.Order = len(items)
		items = append(items, d)
	}
	if target != "" {
		i := collectionIndex(items, target)
		if i < 0 {
			return ErrTimelineNotFound
		}
		var raw []byte
		e = tx.QueryRowContext(ctx, `SELECT revision, cursor_no, timeline_json FROM aicut_timelines WHERE project_id = ? AND owner_username = ? AND id = ? FOR UPDATE`, p, o, target).Scan(&items[i].Value.Revision, &items[i].Cursor, &raw)
		if errors.Is(e, sql.ErrNoRows) {
			return ErrTimelineNotFound
		}
		if e != nil {
			return e
		}
		revision := items[i].Value.Revision
		if e = json.Unmarshal(raw, &items[i].Value); e != nil {
			return e
		}
		items[i].Value.Revision = revision
		hr, e := tx.QueryContext(ctx, `SELECT sequence_no, timeline_json FROM aicut_named_timeline_history WHERE project_id = ? AND owner_username = ? AND timeline_id = ? ORDER BY sequence_no`, p, o, target)
		if e != nil {
			return e
		}
		for hr.Next() {
			var seq int
			var raw []byte
			if e = hr.Scan(&seq, &raw); e != nil {
				hr.Close()
				return e
			}
			if seq != len(items[i].History) {
				hr.Close()
				return ErrInvalidTimelineCollection
			}
			var v Timeline
			if e = json.Unmarshal(raw, &v); e != nil {
				hr.Close()
				return e
			}
			items[i].History = append(items[i].History, v)
		}
		e = hr.Err()
		hr.Close()
		if e != nil {
			return e
		}
		if len(items[i].History) == 0 || items[i].Cursor < 0 || items[i].Cursor >= len(items[i].History) {
			return ErrInvalidTimelineCollection
		}
		before[target] = items[i]
		before[target] = cloneCollectionRecord(items[i])
	}
	if e = fn(&items); e != nil {
		return e
	}
	after := map[string]bool{}
	for _, item := range items {
		after[item.ID] = true
		old, exists := before[item.ID]
		if exists && reflect.DeepEqual(old, item) {
			continue
		}
		if !exists {
			if item.CreatedAt.IsZero() {
				item.CreatedAt = time.Now().UTC()
				item.UpdatedAt = item.CreatedAt
			}
			raw, e := json.Marshal(item.Value)
			if e != nil {
				return e
			}
			if _, e = tx.ExecContext(ctx, `INSERT INTO aicut_timelines (id, project_id, owner_username, title, sort_index, revision, cursor_no, timeline_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, item.ID, p, o, item.Title, item.Order, item.Value.Revision, item.Cursor, string(raw), item.CreatedAt, item.UpdatedAt); e != nil {
				return e
			}
		} else if target == item.ID {
			raw, e := json.Marshal(item.Value)
			if e != nil {
				return e
			}
			if _, e = tx.ExecContext(ctx, `UPDATE aicut_timelines SET revision = ?, cursor_no = ?, timeline_json = ?, updated_at = ? WHERE project_id = ? AND owner_username = ? AND id = ?`, item.Value.Revision, item.Cursor, string(raw), item.UpdatedAt, p, o, item.ID); e != nil {
				return e
			}
		} else {
			if _, e = tx.ExecContext(ctx, `UPDATE aicut_timelines SET title = ?, sort_index = ?, updated_at = ? WHERE project_id = ? AND owner_username = ? AND id = ?`, item.Title, item.Order, item.UpdatedAt, p, o, item.ID); e != nil {
				return e
			}
		}
		if !reflect.DeepEqual(old.History, item.History) {
			if _, e = tx.ExecContext(ctx, `DELETE FROM aicut_named_timeline_history WHERE project_id = ? AND owner_username = ? AND timeline_id = ? AND sequence_no >= ?`, p, o, item.ID, len(item.History)); e != nil {
				return e
			}
			for seq, v := range item.History {
				if seq < len(old.History) && reflect.DeepEqual(old.History[seq], v) {
					continue
				}
				raw, e := json.Marshal(v)
				if e != nil {
					return e
				}
				if _, e = tx.ExecContext(ctx, `INSERT INTO aicut_named_timeline_history (project_id, owner_username, timeline_id, sequence_no, timeline_json) VALUES (?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE timeline_json = VALUES(timeline_json)`, p, o, item.ID, seq, string(raw)); e != nil {
					return e
				}
			}
		}
	}
	for id := range before {
		if !after[id] {
			if _, e = tx.ExecContext(ctx, `DELETE FROM aicut_timelines WHERE project_id = ? AND owner_username = ? AND id = ?`, p, o, id); e != nil {
				return e
			}
		}
	}
	return tx.Commit()
}
func cloneCollectionRecord(r timelineCollectionRecord) timelineCollectionRecord {
	r.Value = cloneTimeline(r.Value)
	history := make([]Timeline, len(r.History))
	for i := range r.History {
		history[i] = cloneTimeline(r.History[i])
	}
	r.History = history
	return r
}
