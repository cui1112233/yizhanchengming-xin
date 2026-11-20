package novelcreation

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

func cloneAgentDefaults(a AgentSettings) AgentSettings {
	raw, _ := json.Marshal(a)
	type plain AgentSettings
	var out plain
	_ = json.Unmarshal(raw, &out)
	return AgentSettings(out)
}
func (s *MemoryStore) GetAgentDefaults(_ context.Context, owner string) (AgentSettings, error) {
	owner, e := requiredOwner(owner)
	if e != nil {
		return AgentSettings{}, e
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneAgentDefaults(s.agentDefaults[owner]), nil
}
func (s *MemoryStore) PutAgentDefaults(_ context.Context, owner string, a AgentSettings) (AgentSettings, error) {
	owner, e := requiredOwner(owner)
	if e != nil {
		return AgentSettings{}, e
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.agentDefaults[owner] = cloneAgentDefaults(a)
	return cloneAgentDefaults(a), nil
}
func (s *MySQLStore) GetAgentDefaults(ctx context.Context, owner string) (AgentSettings, error) {
	owner, e := requiredOwner(owner)
	if e != nil {
		return AgentSettings{}, e
	}
	var raw []byte
	var a AgentSettings
	e = s.db.QueryRowContext(ctx, `SELECT agent_json FROM novel_agent_defaults WHERE owner_username=?`, owner).Scan(&raw)
	if errors.Is(e, sql.ErrNoRows) {
		return a, nil
	}
	if e != nil {
		return a, persistenceError(e)
	}
	if e = json.Unmarshal(raw, &a); e != nil {
		return a, persistenceError(e)
	}
	return a, nil
}
func (s *MySQLStore) PutAgentDefaults(ctx context.Context, owner string, a AgentSettings) (AgentSettings, error) {
	owner, e := requiredOwner(owner)
	if e != nil {
		return AgentSettings{}, e
	}
	raw, e := json.Marshal(a)
	if e != nil {
		return AgentSettings{}, persistenceError(e)
	}
	_, e = s.db.ExecContext(ctx, `INSERT INTO novel_agent_defaults (owner_username,agent_json,updated_at) VALUES (?,?,?) ON DUPLICATE KEY UPDATE agent_json=VALUES(agent_json),updated_at=VALUES(updated_at)`, owner, string(raw), time.Now().UTC())
	if e != nil {
		return AgentSettings{}, persistenceError(e)
	}
	return a, nil
}
