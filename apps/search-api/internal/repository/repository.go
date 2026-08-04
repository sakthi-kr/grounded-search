package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sakthi-kr/grounded-search/apps/search-api/internal/authorization"
)

var ErrNotFound = errors.New("repository entity not found")

type Repository struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) (*Repository, error) {
	if pool == nil {
		return nil, fmt.Errorf("repository pool is nil")
	}
	return &Repository{pool: pool}, nil
}

func (r *Repository) GetUser(ctx context.Context, id string) (User, error) {
	var user User
	err := r.pool.QueryRow(ctx, `SELECT id::text,external_id,display_name,email,enabled,created_at,updated_at FROM users WHERE id=$1`, id).Scan(&user.ID, &user.ExternalID, &user.DisplayName, &user.Email, &user.Enabled, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

func (r *Repository) GetDocument(ctx context.Context, id string) (Document, error) {
	var d Document
	err := r.pool.QueryRow(ctx, `SELECT id::text,source_type,source_id,title,content_hash,status,is_public,source_updated_at,created_at,updated_at FROM documents WHERE id=$1`, id).Scan(&d.ID, &d.SourceType, &d.SourceID, &d.Title, &d.ContentHash, &d.Status, &d.Public, &d.SourceUpdatedAt, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Document{}, ErrNotFound
	}
	if err != nil {
		return Document{}, fmt.Errorf("get document: %w", err)
	}
	return d, nil
}

func (r *Repository) GroupIDsForUser(ctx context.Context, userID string) (map[string]struct{}, error) {
	rows, err := r.pool.Query(ctx, `SELECT g.id::text FROM groups g JOIN group_memberships gm ON gm.group_id=g.id WHERE gm.user_id=$1 AND g.enabled=true`, userID)
	if err != nil {
		return nil, fmt.Errorf("list user groups: %w", err)
	}
	defer rows.Close()
	result := map[string]struct{}{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan user group: %w", err)
		}
		result[id] = struct{}{}
	}
	return result, rows.Err()
}

func (r *Repository) LoadAuthorization(ctx context.Context, userID, documentID string) (authorization.User, authorization.Document, error) {
	user, err := r.GetUser(ctx, userID)
	if errors.Is(err, ErrNotFound) {
		return authorization.User{Exists: false}, authorization.Document{}, nil
	}
	if err != nil {
		return authorization.User{}, authorization.Document{}, err
	}
	doc, err := r.GetDocument(ctx, documentID)
	if err != nil {
		return authorization.User{}, authorization.Document{}, err
	}
	groups, err := r.GroupIDsForUser(ctx, userID)
	if err != nil {
		return authorization.User{}, authorization.Document{}, err
	}
	userRules, err := r.userRules(ctx, documentID)
	if err != nil {
		return authorization.User{}, authorization.Document{}, err
	}
	groupRules, err := r.groupRules(ctx, documentID)
	if err != nil {
		return authorization.User{}, authorization.Document{}, err
	}
	return authorization.User{Exists: true, Enabled: user.Enabled, GroupIDs: groups}, authorization.Document{Status: authorization.DocumentStatus(doc.Status), Public: doc.Public, UserRules: userRules, GroupRules: groupRules}, nil
}

func (r *Repository) userRules(ctx context.Context, documentID string) (map[string]authorization.Decision, error) {
	rows, err := r.pool.Query(ctx, `SELECT user_id::text,decision FROM document_acl_users WHERE document_id=$1`, documentID)
	if err != nil {
		return nil, fmt.Errorf("list user ACLs: %w", err)
	}
	defer rows.Close()
	result := map[string]authorization.Decision{}
	for rows.Next() {
		var id, decision string
		if err := rows.Scan(&id, &decision); err != nil {
			return nil, err
		}
		result[id] = authorization.Decision(decision)
	}
	return result, rows.Err()
}
func (r *Repository) groupRules(ctx context.Context, documentID string) (map[string]authorization.Decision, error) {
	rows, err := r.pool.Query(ctx, `SELECT group_id::text,decision FROM document_acl_groups WHERE document_id=$1`, documentID)
	if err != nil {
		return nil, fmt.Errorf("list group ACLs: %w", err)
	}
	defer rows.Close()
	result := map[string]authorization.Decision{}
	for rows.Next() {
		var id, decision string
		if err := rows.Scan(&id, &decision); err != nil {
			return nil, err
		}
		result[id] = authorization.Decision(decision)
	}
	return result, rows.Err()
}
