package batchfactorytext

import "context"

type PageCounts struct {
	Total int `json:"total"`
	Ready int `json:"ready"`
}

// BatchPageCounts excludes the cover, historical pagination and old artifacts.
// A single grouped read keeps the workbench from loading every full page body.
func (r *MySQLRepository) BatchPageCounts(ctx context.Context, owner, batchID string) (map[string]PageCounts, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT p.book_id, COUNT(*), SUM(CASE WHEN p.status='ready' AND a.id IS NOT NULL THEN 1 ELSE 0 END)
FROM `+pagesTable+` p JOIN `+documents+` d ON d.owner_username=p.owner_username AND d.book_id=p.book_id AND d.active_version_id=p.version_id AND d.pagination_version=p.pagination_version
LEFT JOIN `+artifacts+` a ON a.id=p.active_artifact_id AND a.page_id=p.id AND a.artifact_kind='page' AND a.render_version=p.render_version
WHERE d.owner_username=? AND d.batch_id=? GROUP BY p.book_id`, owner, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string]PageCounts{}
	for rows.Next() {
		var id string
		var counts PageCounts
		if err := rows.Scan(&id, &counts.Total, &counts.Ready); err != nil {
			return nil, err
		}
		result[id] = counts
	}
	return result, rows.Err()
}
