package repository

import "time"

type User struct {
	ID, ExternalID, DisplayName string
	Email                       *string
	Enabled                     bool
	CreatedAt, UpdatedAt        time.Time
}
type Group struct {
	ID, ExternalID, DisplayName string
	Enabled                     bool
	CreatedAt, UpdatedAt        time.Time
}
type Document struct {
	ID, SourceType, SourceID, Title, ContentHash, Status string
	Public                                               bool
	SourceUpdatedAt                                      *time.Time
	CreatedAt, UpdatedAt                                 time.Time
}
type Chunk struct {
	ID, DocumentID, Content, ContentHash string
	Ordinal, StartOffset, EndOffset      int
}
type ACLRule struct{ PrincipalID, Decision string }
