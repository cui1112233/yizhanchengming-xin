package batchfactorytext

import (
	"context"
	"crypto/rand"
	"strings"
	"unicode/utf8"
)

type LibraryBackground struct {
	ID   string
	Name string
	ArtifactFile
}

const librarySelect = "SELECT id,name,storage_ref,media_type,byte_size,sha256 FROM batch_factory_v12_text_backgrounds WHERE owner_username=?"

func readBackground(row scanner) (LibraryBackground, error) {
	var b LibraryBackground
	err := row.Scan(&b.ID, &b.Name, &b.StorageRef, &b.MediaType, &b.ByteSize, &b.SHA256)
	return b, lookupErr(err)
}
func (r *MySQLRepository) GetBackground(ctx context.Context, owner, id string) (LibraryBackground, error) {
	return readBackground(r.db.QueryRowContext(ctx, librarySelect+" AND id=?", owner, id))
}
func (r *MySQLRepository) ListBackgrounds(ctx context.Context, owner string) ([]LibraryBackground, error) {
	return readRows(ctx, r.db, readBackground, librarySelect+" ORDER BY created_at DESC,id", owner)
}
func (r *MySQLRepository) SaveBackground(ctx context.Context, owner, name string, file ArtifactFile) (LibraryBackground, error) {
	name = strings.TrimSpace(name)
	if owner == "" || name == "" || utf8.RuneCountInString(name) > 191 || file.ByteSize < 1 || file.StorageRef == "" {
		return LibraryBackground{}, ErrInvalid
	}
	b := LibraryBackground{ID: rand.Text(), Name: name, ArtifactFile: file}
	_, err := r.db.ExecContext(ctx, "INSERT INTO batch_factory_v12_text_backgrounds(id,owner_username,name,storage_ref,media_type,byte_size,sha256,created_at) VALUES(?,?,?,?,?,?,?,NOW(6))", b.ID, owner, b.Name, b.StorageRef, b.MediaType, b.ByteSize, b.SHA256)
	return b, err
}
