package intake

import (
	"context"
	"fmt"
	"time"
)

func (s *MySQLStore) ListRunnableRunIDs(ctx context.Context, now, staleBefore time.Time, limit int) ([]int64, error) {
	if limit <= 0 {
		limit = 100
	}
	const query = "SELECT id FROM runs WHERE (status = 'pending' AND run_at <= ?) OR (status = 'running' AND updated_at <= ?) ORDER BY run_at ASC, id ASC LIMIT ?"
	rows, err := s.db.QueryContext(ctx, query, now, staleBefore, limit)
	if err != nil {
		return nil, fmt.Errorf("list runnable runs: %w", err)
	}
	defer rows.Close()
	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan runnable run: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate runnable runs: %w", err)
	}
	return ids, nil
}
