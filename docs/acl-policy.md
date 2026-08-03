# Deny-Overrides ACL Policy

GroundedSearch evaluates document access in a single central function.

## Order

1. Deleted document: deny.
2. Disabled document: deny.
3. Unknown user: deny.
4. Disabled user: deny.
5. Explicit user deny: deny.
6. Matching group deny: deny.
7. Public document: allow.
8. Explicit user allow: allow.
9. Matching group allow: allow.
10. No matching rule: deny.

The ordering is security-critical. A deny rule must override public, user allow,
and group allow conditions.

## Boundaries

PostgreSQL is authoritative for users, groups, memberships, document lifecycle,
and ACL rules. Search indices may carry ACL-derived filter fields for candidate
reduction, but the final decision must use the authoritative policy.

The evaluator is intentionally independent of retrieval and generation so it can
be reused before reranking, before answer generation, and in direct document
access endpoints.
