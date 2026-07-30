# ADR-003: Use Authoritative Deny-Overrides ACL Evaluation

- Status: Accepted
- Date: 2026-07-30
- Decision owners: GroundedSearch project
- Scope: Document-level authorisation

## Context

GroundedSearch must ensure that protected documents never reach users or
downstream model services without authorisation.

The system supports:

- public documents;
- explicit user allow rules;
- explicit group allow rules;
- explicit user deny rules;
- explicit group deny rules;
- disabled users;
- disabled or deleted documents.

Search indices may contain stale or incomplete payload metadata. Client-provided
identity or group claims cannot be treated as authoritative.

## Decision

The Go search API will perform the final authorisation decision using identity,
group, document, and ACL state from PostgreSQL.

The policy evaluation order is:

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

Deny rules override allow rules. Missing or malformed ACL data fails closed.

## Identity rules

1. The client may submit a user identifier.
2. The server resolves the user from PostgreSQL.
3. The server resolves group memberships from PostgreSQL.
4. Client-provided group lists are ignored.
5. Unknown users cannot access protected content.
6. Disabled users cannot access protected content.
7. Every request re-evaluates authorisation using current state.

## Secure processing order

The required request sequence is:

```text
validate request
-> resolve identity and groups
-> retrieve candidates
-> evaluate ACLs
-> rerank authorised candidates
-> generate from authorised evidence
-> validate authorised citations
-> build response
```

Unauthorised candidates must not reach:

- the cross-encoder reranker;
- the interaction-aware ranker;
- the answer provider;
- the citation builder;
- protected-response caches;
- logs containing content;
- traces containing content.

## Caching rules

Protected response caches must include:

- user identity or an equivalent security-context identifier;
- group-membership version;
- query and filters;
- retrieval mode;
- index version;
- ACL or document-policy version where needed.

Permission changes must invalidate or bypass stale protected cache entries.

Caching raw retrieval candidates before final authorisation is preferable to
sharing final protected responses across users.

## Result-count rules

Public API result counts must be calculated after ACL filtering.

The API must not expose:

- raw candidate counts;
- hidden document identifiers;
- snippets from denied documents;
- score explanations for denied documents.

## Direct-document access

`GET /v1/documents/{document_id}` must independently authorise the request.

A document appearing in a previous result does not create permanent access.

Unauthorised and absent document responses should avoid unnecessary enumeration
differences where practical.

## Audit rules

Audit records may include:

- request or trace identifier;
- user identifier where allowed;
- decision outcome;
- policy rule category;
- timestamp;
- component and status.

Audit records must not include raw protected document content.

## Consequences

### Positive

- The policy is explicit and testable.
- Deny precedence prevents accidental broad grants.
- Authorisation remains independent of stale index payloads.
- Downstream ML services receive only authorised content.
- Permission changes can take effect without rebuilding every index immediately.

### Negative

- PostgreSQL becomes a critical dependency for protected requests.
- ACL checks add latency.
- Cache design becomes more complex.
- Cross-store lifecycle consistency still requires monitoring.
- Timing differences cannot be eliminated completely.

## Alternatives considered

### Allow rules override deny rules

Rejected because explicit revocation must take precedence.

### ACL filtering inside each search engine only

Rejected because index payloads are derived and may be stale or inconsistent.

### Filter after reranking

Rejected because protected content would cross the policy boundary into the ML
service.

### Filter only before response serialisation

Rejected because protected content could enter prompts, caches, traces, or model
memory.

### Trust client-provided group memberships

Rejected because clients are untrusted.

## Validation

This decision is successful when automated tests demonstrate:

- zero unauthorised documents across at least 10,000 generated ACL cases;
- zero unauthorised snippets;
- zero unauthorised citations;
- zero unauthorised reranker inputs;
- zero unauthorised answer-provider inputs;
- zero cross-user cache leaks;
- immediate safe behaviour after membership or ACL revocation.

Any non-zero leakage blocks release.

## Revisit triggers

Reconsider this model if:

- production identity-provider integration requires richer policies;
- attribute-based access control is required;
- policy evaluation moves to a dedicated verified policy engine;
- multi-tenant isolation requires stronger boundaries;
- measured ACL latency requires a new safe optimisation.
