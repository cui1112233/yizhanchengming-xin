package imagecapacity

import (
	"context"
	"database/sql"
	"fmt"
)

type MySQLRepository struct{ db *sql.DB }

func NewMySQLRepository(db *sql.DB) *MySQLRepository { return &MySQLRepository{db: db} }
func SchemaStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS image_capacity_accounts (account_key VARCHAR(64) PRIMARY KEY, group_key VARCHAR(128) NOT NULL DEFAULT '', account_limit INT NOT NULL DEFAULT 4, updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS image_capacity_policies (account_key VARCHAR(64) NOT NULL, profile_key VARCHAR(64) NOT NULL, model VARCHAR(255) NOT NULL, size VARCHAR(64) NOT NULL DEFAULT '', quality VARCHAR(64) NOT NULL DEFAULT '', resolution VARCHAR(64) NOT NULL DEFAULT '', has_references BOOLEAN NOT NULL DEFAULT FALSE, item_count INT NOT NULL DEFAULT 1, model_limit INT NOT NULL DEFAULT 4, initial_limit INT NOT NULL DEFAULT 4, verified BOOLEAN NOT NULL DEFAULT FALSE, updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, PRIMARY KEY(account_key, profile_key))`,
	}
}
func (r *MySQLRepository) EnsureSchema(ctx context.Context) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("image capacity MySQL unavailable")
	}
	for _, s := range SchemaStatements() {
		if _, e := r.db.ExecContext(ctx, s); e != nil {
			return e
		}
	}
	var columns int
	if e := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='image_capacity_policies' AND COLUMN_NAME='item_count'`).Scan(&columns); e != nil {
		return e
	}
	if columns == 0 {
		if _, e := r.db.ExecContext(ctx, `ALTER TABLE image_capacity_policies ADD COLUMN item_count INT NOT NULL DEFAULT 1 AFTER has_references`); e != nil {
			return e
		}
	}
	return nil
}
func (r *MySQLRepository) Resolve(ctx context.Context, q Request) (Policy, error) {
	if r == nil || r.db == nil {
		return defaultPolicy(q), fmt.Errorf("image capacity MySQL unavailable")
	}
	// 并发完全由运行时自动判定：起始 20 路、上限 30 路，遇到限流自动降档。
	// 历史表里手工保存过的 4 路配置不再参与计算，避免旧值把自动并发压回 4。
	return normalize(defaultPolicy(q), q), nil
}
func (r *MySQLRepository) Save(ctx context.Context, p Policy) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("image capacity MySQL unavailable")
	}
	if len(p.AccountKey) != 64 || p.ProfileKey != ProfileKey(Request{Model: p.Model, Size: p.Size, References: p.References, Quality: p.Quality, Resolution: p.Resolution, Count: p.Count}) || p.AccountLimit < 1 || p.ModelLimit < 1 || p.InitialLimit < 1 || p.InitialLimit > p.ModelLimit {
		return fmt.Errorf("invalid image capacity policy")
	}
	tx, e := r.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	_, e = tx.ExecContext(ctx, `INSERT INTO image_capacity_accounts (account_key,group_key,account_limit) VALUES (?,?,?) ON DUPLICATE KEY UPDATE group_key=VALUES(group_key),account_limit=VALUES(account_limit)`, p.AccountKey, p.GroupKey, p.AccountLimit)
	if e != nil {
		return e
	}
	_, e = tx.ExecContext(ctx, `INSERT INTO image_capacity_policies (account_key,profile_key,model,size,quality,resolution,has_references,item_count,model_limit,initial_limit,verified) VALUES (?,?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE model=VALUES(model),size=VALUES(size),quality=VALUES(quality),resolution=VALUES(resolution),has_references=VALUES(has_references),item_count=VALUES(item_count),model_limit=VALUES(model_limit),initial_limit=VALUES(initial_limit),verified=VALUES(verified)`, p.AccountKey, p.ProfileKey, p.Model, p.Size, p.Quality, p.Resolution, p.References, imageCount(p.Count), p.ModelLimit, p.InitialLimit, p.Verified)
	if e != nil {
		return e
	}
	return tx.Commit()
}
