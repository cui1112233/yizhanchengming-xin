package storage

import (
	"context"
	"database/sql"
)

// adoptAgentCanvasNodeConfigLedger records the historical node-config
// migration only after confirming the exact nullable JSON column already
// exists. A partially recorded local database must never run this ALTER twice.
func adoptAgentCanvasNodeConfigLedger(ctx context.Context, tx *sql.Tx) (bool, error) {
	var count int
	err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'agent_canvas_nodes' AND column_name = 'config_json' AND data_type = 'json' AND is_nullable = 'YES'`).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 1, nil
}
