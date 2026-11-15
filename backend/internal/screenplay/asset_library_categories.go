package screenplay

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type AssetLibraryCategoryState struct {
	CustomCategories []string `json:"custom_categories"`
	Revision         int64    `json:"revision"`
}

// Category persistence stays separate so existing library adapters remain valid.
type AssetLibraryCategoryRepository interface {
	LoadAssetLibraryCategories(context.Context, string) (AssetLibraryCategoryState, error)
	ChangeAssetLibraryCategory(context.Context, string, string, string, int64) (AssetLibraryCategoryState, error)
}

var _ AssetLibraryCategoryRepository = (*MemoryAssetLibraryRepository)(nil)
var _ AssetLibraryCategoryRepository = (*MySQLStore)(nil)

func BuiltinAssetLibraryCategories() []string {
	return []string{"古言", "现言", "男生", "女生", "男主", "女主", "男配", "女配", "小孩", "中老年", "老人", "其他职业", "动物"}
}
func cloneLibraryCategories(state AssetLibraryCategoryState) AssetLibraryCategoryState {
	state.CustomCategories = append([]string{}, state.CustomCategories...)
	return state
}
func validCategoryOwner(owner string) bool {
	return strings.TrimSpace(owner) != "" && utf8.RuneCountInString(owner) <= 191
}
func validateLibraryCategoryChange(owner, action, name string, expected int64) error {
	if !validCategoryOwner(owner) || (action != "add" && action != "delete") || expected < 0 || strings.TrimSpace(name) == "" || !utf8.ValidString(name) || utf8.RuneCountInString(name) > 191 {
		return fmt.Errorf("%w: 分类名称须为1至191字，操作和版本须有效", ErrInvalidArgument)
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return fmt.Errorf("%w: 分类名称不能包含控制字符", ErrInvalidArgument)
		}
	}
	for _, builtin := range BuiltinAssetLibraryCategories() {
		if name == builtin && action == "delete" {
			return fmt.Errorf("%w: 内置分类不能删除", ErrInvalidArgument)
		}
	}
	return nil
}
func changeLibraryCategory(state AssetLibraryCategoryState, action, name string, expected int64) (AssetLibraryCategoryState, error) {
	state = cloneLibraryCategories(state)
	found := -1
	for i, v := range state.CustomCategories {
		if v == name {
			found = i
			break
		}
	}
	builtin := false
	for _, v := range BuiltinAssetLibraryCategories() {
		if v == name {
			builtin = true
			break
		}
	}
	// Retry after a lost response is safe when the requested result is confirmed.
	if (action == "add" && (found >= 0 || builtin)) || (action == "delete" && found < 0) {
		return state, nil
	}
	if expected != state.Revision {
		return AssetLibraryCategoryState{}, fmt.Errorf("%w: 分类已变化，请刷新后重试", ErrRevisionConflict)
	}
	if action == "add" {
		state.CustomCategories = append(state.CustomCategories, name)
	} else {
		state.CustomCategories = append(state.CustomCategories[:found], state.CustomCategories[found+1:]...)
	}
	state.Revision++
	return state, nil
}
func (m *MemoryAssetLibraryRepository) LoadAssetLibraryCategories(_ context.Context, owner string) (AssetLibraryCategoryState, error) {
	if !validCategoryOwner(owner) {
		return AssetLibraryCategoryState{}, ErrInvalidArgument
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return cloneLibraryCategories(m.categories[owner]), nil
}
func (m *MemoryAssetLibraryRepository) ChangeAssetLibraryCategory(_ context.Context, owner, action, name string, expected int64) (AssetLibraryCategoryState, error) {
	if err := validateLibraryCategoryChange(owner, action, name, expected); err != nil {
		return AssetLibraryCategoryState{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	state, err := changeLibraryCategory(m.categories[owner], action, name, expected)
	if err != nil {
		return AssetLibraryCategoryState{}, err
	}
	if m.categories == nil {
		m.categories = map[string]AssetLibraryCategoryState{}
	}
	m.categories[owner] = cloneLibraryCategories(state)
	return state, nil
}
func (s *MySQLStore) LoadAssetLibraryCategories(ctx context.Context, owner string) (AssetLibraryCategoryState, error) {
	if !validCategoryOwner(owner) {
		return AssetLibraryCategoryState{}, ErrInvalidArgument
	}
	var raw []byte
	err := s.db.QueryRowContext(ctx, `SELECT state_json FROM screenplay_asset_library_categories WHERE owner_username = ?`, owner).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return cloneLibraryCategories(AssetLibraryCategoryState{}), nil
	}
	if err != nil {
		return AssetLibraryCategoryState{}, err
	}
	var state AssetLibraryCategoryState
	if err = json.Unmarshal(raw, &state); err != nil {
		return AssetLibraryCategoryState{}, err
	}
	return cloneLibraryCategories(state), nil
}
func (s *MySQLStore) ChangeAssetLibraryCategory(ctx context.Context, owner, action, name string, expected int64) (AssetLibraryCategoryState, error) {
	if err := validateLibraryCategoryChange(owner, action, name, expected); err != nil {
		return AssetLibraryCategoryState{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return AssetLibraryCategoryState{}, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `INSERT IGNORE INTO screenplay_asset_library_categories(owner_username,state_json) VALUES (?,JSON_OBJECT('custom_categories',JSON_ARRAY(),'revision',0))`, owner); err != nil {
		return AssetLibraryCategoryState{}, err
	}
	var raw []byte
	if err = tx.QueryRowContext(ctx, `SELECT state_json FROM screenplay_asset_library_categories WHERE owner_username = ? FOR UPDATE`, owner).Scan(&raw); err != nil {
		return AssetLibraryCategoryState{}, err
	}
	var state AssetLibraryCategoryState
	if err = json.Unmarshal(raw, &state); err != nil {
		return AssetLibraryCategoryState{}, err
	}
	next, err := changeLibraryCategory(state, action, name, expected)
	if err != nil {
		return AssetLibraryCategoryState{}, err
	}
	if next.Revision != state.Revision {
		raw, err = json.Marshal(next)
		if err != nil {
			return AssetLibraryCategoryState{}, err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE screenplay_asset_library_categories SET state_json = ? WHERE owner_username = ?`, string(raw), owner); err != nil {
			return AssetLibraryCategoryState{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return AssetLibraryCategoryState{}, err
	}
	return next, nil
}
