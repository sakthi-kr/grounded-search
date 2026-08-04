package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type DocumentWrite struct {
	Document              Document
	Chunks                []Chunk
	UserRules, GroupRules []ACLRule
}

func (r *Repository) ReplaceDocument(ctx context.Context, input DocumentWrite) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin document transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `INSERT INTO documents(id,source_type,source_id,title,content_hash,status,is_public,source_updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8)
 ON CONFLICT(id) DO UPDATE SET source_type=EXCLUDED.source_type,source_id=EXCLUDED.source_id,title=EXCLUDED.title,content_hash=EXCLUDED.content_hash,status=EXCLUDED.status,is_public=EXCLUDED.is_public,source_updated_at=EXCLUDED.source_updated_at,updated_at=now()`, input.Document.ID, input.Document.SourceType, input.Document.SourceID, input.Document.Title, input.Document.ContentHash, input.Document.Status, input.Document.Public, input.Document.SourceUpdatedAt)
	if err != nil {
		return fmt.Errorf("upsert document: %w", err)
	}
	for _, table := range []string{"document_chunks", "document_acl_users", "document_acl_groups"} {
		if _, err = tx.Exec(ctx, "DELETE FROM "+table+" WHERE document_id=$1", input.Document.ID); err != nil {
			return fmt.Errorf("clear %s: %w", table, err)
		}
	}
	for _, chunk := range input.Chunks {
		if _, err = tx.Exec(ctx, `INSERT INTO document_chunks(id,document_id,ordinal,content,content_hash,start_offset,end_offset) VALUES($1,$2,$3,$4,$5,$6,$7)`, chunk.ID, input.Document.ID, chunk.Ordinal, chunk.Content, chunk.ContentHash, chunk.StartOffset, chunk.EndOffset); err != nil {
			return fmt.Errorf("insert chunk: %w", err)
		}
	}
	for _, rule := range input.UserRules {
		if _, err = tx.Exec(ctx, `INSERT INTO document_acl_users(document_id,user_id,decision) VALUES($1,$2,$3)`, input.Document.ID, rule.PrincipalID, rule.Decision); err != nil {
			return fmt.Errorf("insert user ACL: %w", err)
		}
	}
	for _, rule := range input.GroupRules {
		if _, err = tx.Exec(ctx, `INSERT INTO document_acl_groups(document_id,group_id,decision) VALUES($1,$2,$3)`, input.Document.ID, rule.PrincipalID, rule.Decision); err != nil {
			return fmt.Errorf("insert group ACL: %w", err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit document transaction: %w", err)
	}
	return nil
}
