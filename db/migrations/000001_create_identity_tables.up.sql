BEGIN;

CREATE TABLE users (
    id uuid PRIMARY KEY,
    external_id text NOT NULL UNIQUE,
    display_name text NOT NULL,
    email text,
    enabled boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT users_external_id_not_blank
        CHECK (btrim(external_id) <> ''),
    CONSTRAINT users_display_name_not_blank
        CHECK (btrim(display_name) <> ''),
    CONSTRAINT users_email_not_blank
        CHECK (email IS NULL OR btrim(email) <> '')
);

CREATE UNIQUE INDEX users_email_lower_unique
    ON users (lower(email))
    WHERE email IS NOT NULL;

CREATE TABLE groups (
    id uuid PRIMARY KEY,
    external_id text NOT NULL UNIQUE,
    display_name text NOT NULL,
    enabled boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT groups_external_id_not_blank
        CHECK (btrim(external_id) <> ''),
    CONSTRAINT groups_display_name_not_blank
        CHECK (btrim(display_name) <> '')
);

CREATE TABLE group_memberships (
    user_id uuid NOT NULL
        REFERENCES users(id)
        ON DELETE CASCADE,
    group_id uuid NOT NULL
        REFERENCES groups(id)
        ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, group_id)
);

CREATE INDEX group_memberships_group_id_idx
    ON group_memberships (group_id);

COMMIT;
