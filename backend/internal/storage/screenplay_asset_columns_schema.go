package storage

import (
	"context"
	"database/sql"
	"fmt"
)

// Released databases may have the old 8800101 ledger without these later-added
// CREATE TABLE fields. A new version repairs only absent fields, including a
// partially completed DDL run; it never rewrites existing asset values.
func screenplayAssetColumnsMigration() Migration {
	return Migration{Version: 8800113, CallbackChecksum: "v88-screenplay-asset-fields-reconcile-v1", Adopt: func(ctx context.Context, tx *sql.Tx) (bool, error) {
		err := reconcileScreenplayAssetColumns(ctx, tx)
		return err == nil, err
	}, Reconcile: reconcileScreenplayAssetColumns}
}

const screenplayAssetColumnsQuery = `SELECT column_name, data_type, character_maximum_length, is_nullable FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'screenplay_assets' AND column_name IN ('episodes','tag','extra')`

func reconcileScreenplayAssetColumns(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx, screenplayAssetColumnsQuery)
	if err != nil {
		return err
	}
	found := map[string]bool{}
	for rows.Next() {
		var name, kind, nullable string
		var length sql.NullInt64
		if err = rows.Scan(&name, &kind, &length, &nullable); err != nil {
			rows.Close()
			return err
		}
		valid := (name == "extra" && kind == "json" && nullable == "YES") || (kind == "varchar" && nullable == "NO" && ((name == "episodes" && length.Valid && length.Int64 >= 191) || (name == "tag" && length.Valid && length.Int64 >= 64)))
		if !valid {
			rows.Close()
			return fmt.Errorf("screenplay_assets.%s has incompatible definition; preserve existing data and repair schema", name)
		}
		found[name] = true
	}
	err = rows.Err()
	if closeErr := rows.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	for _, column := range []struct{ name, ddl string }{
		{"episodes", "ALTER TABLE screenplay_assets ADD COLUMN episodes VARCHAR(191) NOT NULL DEFAULT ''"},
		{"tag", "ALTER TABLE screenplay_assets ADD COLUMN tag VARCHAR(64) NOT NULL DEFAULT ''"},
		{"extra", "ALTER TABLE screenplay_assets ADD COLUMN extra JSON NULL"},
	} {
		if !found[column.name] {
			if _, err := tx.ExecContext(ctx, column.ddl); err != nil {
				return err
			}
		}
	}
	return nil
}
