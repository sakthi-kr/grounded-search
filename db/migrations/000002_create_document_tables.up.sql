BEGIN;

CREATE TABLE documents (
    id uuid PRIMARY KEY,
    source_type text NOT NULL,
    source_id text NOT NULL,
    title text NOT NULL,
    content_hash text NOT NULL,
    status text NOT NULL DEFAULT 'active',
    is_public boolean NOT NULL DEFAULT false,
    source_updated_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT documents_source_type_not_blank
        CHECK (btrim(source_type) <> ''),
    CONSTRAINT documents_source_id_not_blank
        CHECK (btrim(source_id) <> ''),
    CONSTRAINT documents_title_not_blank
        CHECK (btrim(title) <> ''),
    CONSTRAINT documents_content_hash_not_blank
        CHECK (btrim(content_hash) <> ''),
    CONSTRAINT documents_status_valid
        CHECK (status IN ('active', 'disabled', 'deleted')),
    UNIQUE (source_type, source_id)
);

CREATE INDEX documents_status_idx
    ON documents (status);

CREATE INDEX documents_public_active_idx
    ON documents (is_public)
    WHERE status = 'active';

CREATE TABLE document_chunks (
    id uuid PRIMARY KEY,
    document_id uuid NOT NULL
        REFERENCES documents(id)
        ON DELETE CASCADE,
    ordinal integer NOT NULL,
    content text NOT NULL,
    content_hash text NOT NULL,
    start_offset integer NOT NULL,
    end_offset integer NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT document_chunks_ordinal_non_negative
        CHECK (ordinal >= 0),
    CONSTRAINT document_chunks_content_not_blank
        CHECK (btrim(content) <> ''),
    CONSTRAINT document_chunks_content_hash_not_blank
        CHECK (btrim(content_hash) <> ''),
    CONSTRAINT document_chunks_offsets_valid
        CHECK (
            start_offset >= 0
            AND end_offset >= start_offset
        ),
    UNIQUE (document_id, ordinal)
);

CREATE INDEX document_chunks_document_id_idx
    ON document_chunks (document_id);

COMMIT;
