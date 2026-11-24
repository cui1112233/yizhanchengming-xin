package yzcmusage

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type Store struct{ DB *sql.DB }
type Pending struct {
	Request
	CredentialHash string
}

func (s Store) TeamMonthTokens(ctx context.Context, owners []string) (int, error) {
	if len(owners) == 0 {
		return 0, nil
	}
	zone := time.FixedZone("Asia/Shanghai", 8*60*60)
	now := time.Now().In(zone)
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, zone).UTC()
	args := []any{start}
	marks := make([]string, len(owners))
	for i, owner := range owners {
		marks[i] = "?"
		args = append(args, owner)
	}
	var total int
	err := s.DB.QueryRowContext(ctx, "SELECT COALESCE(SUM(input_tokens+output_tokens),0) FROM yzcm_usage_requests WHERE created_at>=? AND owner IN ("+strings.Join(marks, ",")+")", args...).Scan(&total)
	return total, err
}

func SchemaStatements() []string {
	return []string{`CREATE TABLE IF NOT EXISTS yzcm_usage_requests (
 request_id VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin PRIMARY KEY,
 owner VARCHAR(32) NOT NULL, feature VARCHAR(64) NOT NULL,
 credential_hash CHAR(64) CHARACTER SET ascii NOT NULL,
 created_at DATETIME(6) NOT NULL, http_status INT NOT NULL,
 model VARCHAR(255) NOT NULL DEFAULT '', quota BIGINT NULL,
 input_tokens BIGINT NULL, output_tokens BIGINT NULL,
 actual_cost DECIMAL(24,10) NULL, cost_currency VARCHAR(8) NULL,
 reconciled_at DATETIME(6) NULL, retry_after DATETIME(6) NULL,
 INDEX yzcm_usage_owner_time(owner,created_at),
 INDEX yzcm_usage_pending(reconciled_at,retry_after,created_at)
 ) ENGINE=InnoDB`}
}

func (s Store) Record(ctx context.Context, r Request, keyHash string, status int) error {
	return s.recordAt(ctx, r, keyHash, status, time.Now().UTC())
}
func (s Store) recordAt(ctx context.Context, r Request, keyHash string, status int, at time.Time) error {
	if r.ID == "" || r.Owner == "" || len(r.ID) > 64 || len(r.Owner) > 32 {
		return errors.New("invalid usage identity")
	}
	_, err := s.DB.ExecContext(ctx, `INSERT IGNORE INTO yzcm_usage_requests(request_id,owner,feature,credential_hash,created_at,http_status) VALUES(?,?,?,?,?,?)`, r.ID, r.Owner, r.Feature, keyHash, at.UTC(), status)
	return err
}
func (s Store) Pending(ctx context.Context) ([]Pending, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT request_id,owner,feature,credential_hash FROM yzcm_usage_requests WHERE reconciled_at IS NULL AND created_at >= ? AND (retry_after IS NULL OR retry_after<=?) ORDER BY COALESCE(retry_after,created_at) LIMIT 100`, time.Now().UTC().Add(-31*24*time.Hour), time.Now().UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Pending{}
	for rows.Next() {
		var r Pending
		if err := rows.Scan(&r.ID, &r.Owner, &r.Feature, &r.CredentialHash); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()
	if len(result) > 0 {
		args := []any{time.Now().UTC().Add(2 * time.Minute)}
		marks := make([]string, len(result))
		for i, p := range result {
			marks[i] = "?"
			args = append(args, p.ID)
		}
		if _, err = s.DB.ExecContext(ctx, "UPDATE yzcm_usage_requests SET retry_after=? WHERE request_id IN ("+strings.Join(marks, ",")+")", args...); err != nil {
			return nil, err
		}
	}
	return result, nil
}
func (s Store) Settle(ctx context.Context, e Entry) error {
	_, err := s.DB.ExecContext(ctx, `UPDATE yzcm_usage_requests SET model=?,quota=?,input_tokens=?,output_tokens=?,actual_cost=?,cost_currency=?,reconciled_at=? WHERE request_id=? AND owner=?`, e.Model, e.Quota, e.InputTokens, e.OutputTokens, e.Cost, e.Currency, time.Now().UTC(), e.ID, e.Owner)
	return err
}

// ReadSummary reads only the selected member's indexed current-month rows.
// Requests that have no billing receipt remain pending, never free/estimated.
func (s Store) ReadSummary(ctx context.Context, owner string) (map[string]any, error) {
	zone := time.FixedZone("Asia/Shanghai", 8*60*60)
	now := time.Now().In(zone)
	month := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, zone).UTC()
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, zone).UTC()
	summary := func(since time.Time) (map[string]any, error) {
		var calls, priced, success, input, output int64
		var cost sql.NullFloat64
		err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*),COUNT(actual_cost),COALESCE(SUM(http_status>=200 AND http_status<300),0),COALESCE(SUM(input_tokens),0),COALESCE(SUM(output_tokens),0),SUM(actual_cost) FROM yzcm_usage_requests WHERE owner=? AND created_at>=?`, owner, since).Scan(&calls, &priced, &success, &input, &output, &cost)
		if err != nil {
			return nil, err
		}
		var amount any
		if cost.Valid {
			amount = cost.Float64
		}
		return map[string]any{"calls": calls, "successfulCalls": success, "billableCalls": calls, "pricedCalls": priced, "pendingCalls": calls - priced, "knownUsageCalls": priced, "inputTokens": input, "outputTokens": output, "totalTokens": input + output, "actualCost": amount, "estimatedCost": amount, "costCurrency": "CNY", "byFeature": map[string]int64{}, "callsByFeature": map[string]int64{}, "byModel": map[string]int64{}, "source": "yzcm_billing_log"}, nil
	}
	m, err := summary(month)
	if err != nil {
		return nil, err
	}
	d, err := summary(day)
	if err != nil {
		return nil, err
	}
	rows, err := s.DB.QueryContext(ctx, `SELECT feature,COUNT(*),COALESCE(SUM(input_tokens+output_tokens),0) FROM yzcm_usage_requests WHERE owner=? AND created_at>=? GROUP BY feature`, owner, month)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var feature string
		var calls, tokens int64
		if err = rows.Scan(&feature, &calls, &tokens); err != nil {
			rows.Close()
			return nil, err
		}
		m["byFeature"].(map[string]int64)[feature] = tokens
		m["callsByFeature"].(map[string]int64)[feature] = calls
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	rows, err = s.DB.QueryContext(ctx, `SELECT request_id,created_at,feature,model,input_tokens,output_tokens,actual_cost,cost_currency,http_status FROM yzcm_usage_requests WHERE owner=? ORDER BY created_at DESC,request_id DESC LIMIT 20`, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	recent := []map[string]any{}
	for rows.Next() {
		var id, feature, model string
		var at time.Time
		var input, output sql.NullInt64
		var cost sql.NullFloat64
		var currency sql.NullString
		var status int
		if err = rows.Scan(&id, &at, &feature, &model, &input, &output, &cost, &currency, &status); err != nil {
			return nil, err
		}
		var amount any
		if cost.Valid {
			amount = cost.Float64
		}
		recent = append(recent, map[string]any{"id": id, "at": at.UTC().Format(time.RFC3339), "feature": feature, "model": model, "inputTokens": input.Int64, "outputTokens": output.Int64, "totalTokens": input.Int64 + output.Int64, "actualCost": amount, "estimatedCost": amount, "costCurrency": currency.String, "costKnown": cost.Valid, "usageKnown": input.Valid && output.Valid, "source": "yzcm_billing_log", "httpStatus": status})
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return map[string]any{"day": d, "month": m, "recent": recent, "source": "yzcm_billing_log", "historyCoverage": "tracked_requests", "timeZone": "Asia/Shanghai"}, nil
}
