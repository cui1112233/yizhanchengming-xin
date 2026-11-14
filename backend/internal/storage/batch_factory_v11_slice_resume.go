package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var v11SliceForeignKey = regexp.MustCompile(`^ALTER TABLE (\w+) ADD CONSTRAINT (\w+) FOREIGN KEY \((\w+)\) REFERENCES (\w+)\((\w+)\) ON DELETE RESTRICT$`)

const sliceForeignKeyQuery = `SELECT k.TABLE_NAME,k.COLUMN_NAME,k.REFERENCED_TABLE_NAME,k.REFERENCED_COLUMN_NAME,r.DELETE_RULE,r.UPDATE_RULE
FROM information_schema.KEY_COLUMN_USAGE k
JOIN information_schema.REFERENTIAL_CONSTRAINTS r ON r.CONSTRAINT_SCHEMA=k.CONSTRAINT_SCHEMA AND r.CONSTRAINT_NAME=k.CONSTRAINT_NAME AND r.TABLE_NAME=k.TABLE_NAME
WHERE k.CONSTRAINT_SCHEMA=DATABASE() AND k.CONSTRAINT_NAME=? ORDER BY k.ORDINAL_POSITION`

// MySQL DDL commits independently of sql.Tx. An interrupted migration can
// therefore leave its foreign keys installed without its ledger row. Verify
// every existing key before writing, then complete only the missing steps.
// A same-named key on a different table is a conflict, never proof of success.
func adoptV11SliceOneLedger(ctx context.Context, tx *sql.Tx) (bool, error) {
	statements := V11SliceOneStatements()
	installed := make(map[string]bool)
	found := false
	for _, statement := range statements {
		parts := v11SliceForeignKey.FindStringSubmatch(statement)
		if parts == nil {
			continue
		}
		rows, err := tx.QueryContext(ctx, sliceForeignKeyQuery, parts[2])
		if err != nil {
			return false, err
		}
		count := 0
		valid := true
		for rows.Next() {
			var table, column, referencedTable, referencedColumn, deleteRule, updateRule string
			if err := rows.Scan(&table, &column, &referencedTable, &referencedColumn, &deleteRule, &updateRule); err != nil {
				rows.Close()
				return false, err
			}
			count++
			valid = valid && table == parts[1] && column == parts[3] && referencedTable == parts[4] && referencedColumn == parts[5] && restrictiveRule(deleteRule) && restrictiveRule(updateRule)
		}
		err = errors.Join(rows.Err(), rows.Close())
		if err != nil {
			return false, err
		}
		if count > 0 {
			if count != 1 || !valid {
				return false, fmt.Errorf("existing foreign key %s does not match migration 1100002; preserve the schema and repair its migration history", parts[2])
			}
			installed[statement] = true
			found = true
		}
	}
	if !found {
		return false, nil
	}
	for _, statement := range statements {
		if installed[statement] {
			continue
		}
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return false, err
		}
	}
	return true, nil
}

func restrictiveRule(rule string) bool {
	return strings.EqualFold(rule, "RESTRICT") || strings.EqualFold(rule, "NO ACTION")
}
