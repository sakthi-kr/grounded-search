BEGIN;

CREATE TABLE document_acl_users (
    document_id uuid NOT NULL
        REFERENCES documents(id)
        ON DELETE CASCADE,
    user_id uuid NOT NULL
        REFERENCES users(id)
        ON DELETE CASCADE,
    decision text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (document_id, user_id),
    CONSTRAINT document_acl_users_decision_valid
        CHECK (decision IN ('allow', 'deny'))
);

CREATE INDEX document_acl_users_user_id_idx
    ON document_acl_users (user_id);

CREATE TABLE document_acl_groups (
    document_id uuid NOT NULL
        REFERENCES documents(id)
        ON DELETE CASCADE,
    group_id uuid NOT NULL
        REFERENCES groups(id)
        ON DELETE CASCADE,
    decision text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (document_id, group_id),
    CONSTRAINT document_acl_groups_decision_valid
        CHECK (decision IN ('allow', 'deny'))
);

CREATE INDEX document_acl_groups_group_id_idx
    ON document_acl_groups (group_id);

COMMIT;
