# PostgreSQL Metadata and ACL Design

## 1. Purpose

PostgreSQL is the authoritative source for identity, group membership, document
lifecycle, document chunks, and access-control rules.

Bleve and Qdrant will be derived indices. They must never replace PostgreSQL for
the final authorisation decision.

## 2. Identifier strategy

All domain entities use UUID primary keys supplied by the application or fixture
generator.

Source systems keep their own stable identifiers in dedicated columns such as
`external_id` and `source_id`.

This avoids database-specific UUID-generation extensions in the initial schema.

## 3. Identity model

### `users`

Stores one row per resolved identity.

Important fields:

- `id`: internal UUID;
- `external_id`: stable identity-provider or fixture identifier;
- `display_name`;
- `email`;
- `enabled`: disabled identities fail closed;
- timestamps.

### `groups`

Stores named access-control groups.

### `group_memberships`

Maps users to groups.

The composite primary key prevents duplicate memberships. Foreign keys use
`ON DELETE CASCADE` so deleted users or groups cannot leave orphaned rows.

## 4. Document model

### `documents`

Stores authoritative document metadata and lifecycle state.

Lifecycle states:

- `active`;
- `disabled`;
- `deleted`.

The row remains available after logical deletion so connector reconciliation and
audit logic can distinguish deletion from absence.

`is_public` is an allow condition only. Explicit deny rules still take
precedence.

### `document_chunks`

Stores deterministic chunks and source offsets.

A document cannot contain two chunks with the same ordinal. Character offsets
must be non-negative and the end offset must not precede the start offset.

## 5. ACL model

### `document_acl_users`

Stores explicit user-level `allow` or `deny` decisions.

### `document_acl_groups`

Stores explicit group-level `allow` or `deny` decisions.

The database prevents duplicate rules for the same document and principal.

## 6. Authorisation order

The application must evaluate rules in this order:

```text
1. Deleted document -> deny.
2. Disabled document -> deny.
3. Unknown or disabled user -> deny.
4. Explicit user deny -> deny.
5. Matching group deny -> deny.
6. Public document -> allow.
7. Explicit user allow -> allow.
8. Matching group allow -> allow.
9. Otherwise -> deny.
```

Database constraints preserve valid data shapes. The final deny-overrides
decision remains application logic and will be implemented in a later Phase 2
step.

## 7. Transaction boundaries

The following operations must be transactional:

- creating or updating a document and its chunks;
- replacing document ACL rules;
- updating group membership used for an authorisation change;
- connector checkpoint advancement after a complete metadata update.

## 8. Migration rules

- migrations are immutable after merging;
- every upward migration has a matching downward migration;
- upward migrations apply in filename order;
- downward migrations apply in reverse filename order;
- migration validation must run against a real PostgreSQL container;
- production data migrations must not silently discard rows.

## 9. Current exclusions

This schema does not yet include:

- connector checkpoints;
- interaction events;
- audit-event storage;
- search index versions;
- embedding model versions;
- row-level security policies.

Those will be added only when their owning components are implemented.
