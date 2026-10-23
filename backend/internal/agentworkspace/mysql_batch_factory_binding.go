package agentworkspace

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

const batchFactoryBindingColumns = "owner_username, batch_id, project_id, created_at, updated_at"

func (s *MySQLStore) BindBatchFactoryProject(ctx context.Context, binding BatchFactoryBinding) (BatchFactoryBinding, error) {
	binding, err := normalizeBatchFactoryBinding(binding, time.Now().UTC())
	if err != nil {
		return BatchFactoryBinding{}, err
	}
	existing, err := s.BatchFactoryBinding(ctx, binding.Owner, binding.BatchID)
	if err == nil {
		if existing.ProjectID != binding.ProjectID {
			return BatchFactoryBinding{}, ErrBatchFactoryBindingConflict
		}
		return existing, nil
	}
	if !errors.Is(err, ErrBatchFactoryBindingNotFound) {
		return BatchFactoryBinding{}, err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO agent_batch_factory_bindings (owner_username, batch_id, project_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, binding.Owner, binding.BatchID, binding.ProjectID, binding.CreatedAt, binding.UpdatedAt)
	if err != nil && isBatchFactoryBindingDuplicate(err) {
		return BatchFactoryBinding{}, ErrBatchFactoryBindingConflict
	}
	if err != nil {
		return BatchFactoryBinding{}, err
	}
	return binding, nil
}

func (s *MySQLStore) BatchFactoryBinding(ctx context.Context, owner, batchID string) (BatchFactoryBinding, error) {
	owner, batchID = strings.TrimSpace(owner), strings.TrimSpace(batchID)
	if owner == "" || batchID == "" {
		return BatchFactoryBinding{}, ErrBatchFactoryBindingInvalid
	}
	var binding BatchFactoryBinding
	err := s.db.QueryRowContext(ctx, `SELECT `+batchFactoryBindingColumns+` FROM agent_batch_factory_bindings WHERE owner_username = ? AND batch_id = ?`, owner, batchID).Scan(&binding.Owner, &binding.BatchID, &binding.ProjectID, &binding.CreatedAt, &binding.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return BatchFactoryBinding{}, ErrBatchFactoryBindingNotFound
	}
	return binding, err
}

func (s *MySQLStore) RebindBatchFactoryProject(ctx context.Context, owner, batchID, previousProjectID, projectID string, now time.Time) (BatchFactoryBinding, error) {
	owner, batchID = strings.TrimSpace(owner), strings.TrimSpace(batchID)
	previousProjectID, projectID = strings.TrimSpace(previousProjectID), strings.TrimSpace(projectID)
	if owner == "" || batchID == "" || previousProjectID == "" || projectID == "" {
		return BatchFactoryBinding{}, ErrBatchFactoryBindingInvalid
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return BatchFactoryBinding{}, err
	}
	defer tx.Rollback()
	var binding BatchFactoryBinding
	err = tx.QueryRowContext(ctx, `SELECT `+batchFactoryBindingColumns+` FROM agent_batch_factory_bindings WHERE owner_username = ? AND batch_id = ? FOR UPDATE`, owner, batchID).
		Scan(&binding.Owner, &binding.BatchID, &binding.ProjectID, &binding.CreatedAt, &binding.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return BatchFactoryBinding{}, ErrBatchFactoryBindingNotFound
	}
	if err != nil {
		return BatchFactoryBinding{}, err
	}
	if binding.ProjectID == projectID {
		return binding, nil
	}
	if binding.ProjectID != previousProjectID {
		return BatchFactoryBinding{}, ErrBatchFactoryBindingConflict
	}
	var exists int
	err = tx.QueryRowContext(ctx, `SELECT 1 FROM agent_projects WHERE owner_username = ? AND id = ? AND deleted_at IS NULL FOR UPDATE`, owner, projectID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return BatchFactoryBinding{}, ErrProjectNotFound
	}
	if err != nil {
		return BatchFactoryBinding{}, err
	}
	now = now.UTC().Truncate(time.Microsecond)
	if _, err := tx.ExecContext(ctx, `UPDATE agent_batch_factory_bindings SET project_id = ?, updated_at = ? WHERE owner_username = ? AND batch_id = ? AND project_id = ?`, projectID, now, owner, batchID, previousProjectID); err != nil {
		if isBatchFactoryBindingDuplicate(err) {
			return BatchFactoryBinding{}, ErrBatchFactoryBindingConflict
		}
		return BatchFactoryBinding{}, err
	}
	if err := tx.Commit(); err != nil {
		return BatchFactoryBinding{}, err
	}
	binding.ProjectID, binding.UpdatedAt = projectID, now
	return binding, nil
}

func (s *MySQLStore) BindBatchFactoryNode(ctx context.Context, owner string, origin BatchFactoryNodeOrigin) error {
	origin, err := normalizeBatchFactoryNodeOrigin(owner, origin)
	if err != nil {
		return err
	}
	var exists int
	err = s.db.QueryRowContext(ctx, `SELECT 1 FROM agent_batch_factory_bindings WHERE owner_username = ? AND batch_id = ? AND project_id = ?`, strings.TrimSpace(owner), origin.BatchID, origin.ProjectID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrBatchFactoryBindingNotFound
	}
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO agent_batch_factory_node_origins (project_id, node_id, owner_username, batch_id, book_id, resource_type, resource_id, version_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE batch_id = VALUES(batch_id), book_id = VALUES(book_id), resource_type = VALUES(resource_type), resource_id = VALUES(resource_id), version_id = VALUES(version_id)`, origin.ProjectID, origin.NodeID, strings.TrimSpace(owner), origin.BatchID, origin.BookID, origin.ResourceType, origin.ResourceID, nullableBatchFactoryVersionID(origin.VersionID))
	return err
}

func (s *MySQLStore) BatchFactoryNodeOrigins(ctx context.Context, owner, projectID, bookID string) ([]BatchFactoryNodeOrigin, error) {
	owner, projectID, bookID = strings.TrimSpace(owner), strings.TrimSpace(projectID), strings.TrimSpace(bookID)
	if owner == "" || projectID == "" || bookID == "" {
		return nil, ErrBatchFactoryNodeOriginInvalid
	}
	rows, err := s.db.QueryContext(ctx, `SELECT project_id, node_id, batch_id, book_id, resource_type, resource_id, version_id FROM agent_batch_factory_node_origins WHERE owner_username = ? AND project_id = ? AND book_id = ? ORDER BY node_id`, owner, projectID, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	origins := make([]BatchFactoryNodeOrigin, 0)
	for rows.Next() {
		var origin BatchFactoryNodeOrigin
		var versionID sql.NullString
		if err := rows.Scan(&origin.ProjectID, &origin.NodeID, &origin.BatchID, &origin.BookID, &origin.ResourceType, &origin.ResourceID, &versionID); err != nil {
			return nil, err
		}
		origin.VersionID = versionID.String
		origins = append(origins, origin)
	}
	return origins, rows.Err()
}

func (s *MySQLStore) RemoveBatchFactoryNodeBinding(ctx context.Context, owner, projectID, nodeID string) error {
	owner, projectID, nodeID = strings.TrimSpace(owner), strings.TrimSpace(projectID), strings.TrimSpace(nodeID)
	if owner == "" || projectID == "" || nodeID == "" {
		return ErrBatchFactoryNodeOriginInvalid
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM agent_batch_factory_node_origins WHERE owner_username = ? AND project_id = ? AND node_id = ?`, owner, projectID, nodeID)
	return err
}

func nullableBatchFactoryVersionID(versionID string) any {
	if strings.TrimSpace(versionID) == "" {
		return nil
	}
	return strings.TrimSpace(versionID)
}

func isBatchFactoryBindingDuplicate(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "duplicate")
}

var _ BatchFactoryBindingStore = (*MySQLStore)(nil)
