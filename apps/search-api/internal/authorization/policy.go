package authorization

const (
	ReasonDocumentDeleted    = "document_deleted"
	ReasonDocumentDisabled   = "document_disabled"
	ReasonUnknownUser        = "unknown_user"
	ReasonUserDisabled       = "user_disabled"
	ReasonExplicitUserDeny   = "explicit_user_deny"
	ReasonMatchingGroupDeny  = "matching_group_deny"
	ReasonPublicDocument     = "public_document"
	ReasonExplicitUserAllow  = "explicit_user_allow"
	ReasonMatchingGroupAllow = "matching_group_allow"
	ReasonDefaultDeny        = "default_deny"
)

// Evaluate applies the authoritative deny-overrides ACL policy.
//
// Evaluation order:
//
//  1. deleted document -> deny;
//  2. disabled document -> deny;
//  3. unknown user -> deny;
//  4. disabled user -> deny;
//  5. explicit user deny -> deny;
//  6. matching group deny -> deny;
//  7. public document -> allow;
//  8. explicit user allow -> allow;
//  9. matching group allow -> allow;
//
// 10. otherwise -> deny.
func Evaluate(userID string, user User, document Document) Result {
	switch document.Status {
	case DocumentDeleted:
		return deny(ReasonDocumentDeleted)
	case DocumentDisabled:
		return deny(ReasonDocumentDisabled)
	}

	if !user.Exists {
		return deny(ReasonUnknownUser)
	}
	if !user.Enabled {
		return deny(ReasonUserDisabled)
	}

	if document.UserRules[userID] == DecisionDeny {
		return deny(ReasonExplicitUserDeny)
	}

	if hasMatchingGroupDecision(
		user.GroupIDs,
		document.GroupRules,
		DecisionDeny,
	) {
		return deny(ReasonMatchingGroupDeny)
	}

	if document.Public {
		return allow(ReasonPublicDocument)
	}

	if document.UserRules[userID] == DecisionAllow {
		return allow(ReasonExplicitUserAllow)
	}

	if hasMatchingGroupDecision(
		user.GroupIDs,
		document.GroupRules,
		DecisionAllow,
	) {
		return allow(ReasonMatchingGroupAllow)
	}

	return deny(ReasonDefaultDeny)
}

func hasMatchingGroupDecision(
	groupIDs map[string]struct{},
	rules map[string]Decision,
	target Decision,
) bool {
	for groupID := range groupIDs {
		if rules[groupID] == target {
			return true
		}
	}
	return false
}

func allow(reason string) Result {
	return Result{Allowed: true, Reason: reason}
}

func deny(reason string) Result {
	return Result{Allowed: false, Reason: reason}
}
